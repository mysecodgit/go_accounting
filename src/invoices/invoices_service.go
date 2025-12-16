package invoices

import (
	"database/sql"
	"fmt"
	"math"

	"github.com/mysecodgit/go_accounting/src/accounts"
	"github.com/mysecodgit/go_accounting/src/invoice_items"
	"github.com/mysecodgit/go_accounting/src/items"
	"github.com/mysecodgit/go_accounting/src/splits"
	"github.com/mysecodgit/go_accounting/src/transactions"
)

type InvoiceService struct {
	invoiceRepo     InvoiceRepository
	transactionRepo transactions.TransactionRepository
	splitRepo       splits.SplitRepository
	invoiceItemRepo invoice_items.InvoiceItemRepository
	itemRepo        items.ItemRepository
	accountRepo     accounts.AccountRepository
	db              *sql.DB
}

// Expose invoiceRepo for handler access
func (s *InvoiceService) GetInvoiceRepo() InvoiceRepository {
	return s.invoiceRepo
}

// Expose invoiceItemRepo for handler access
func (s *InvoiceService) GetInvoiceItemRepo() invoice_items.InvoiceItemRepository {
	return s.invoiceItemRepo
}

// Expose splitRepo for handler access
func (s *InvoiceService) GetSplitRepo() splits.SplitRepository {
	return s.splitRepo
}

// Expose transactionRepo for handler access
func (s *InvoiceService) GetTransactionRepo() transactions.TransactionRepository {
	return s.transactionRepo
}

func NewInvoiceService(
	invoiceRepo InvoiceRepository,
	transactionRepo transactions.TransactionRepository,
	splitRepo splits.SplitRepository,
	invoiceItemRepo invoice_items.InvoiceItemRepository,
	itemRepo items.ItemRepository,
	accountRepo accounts.AccountRepository,
	db *sql.DB,
) *InvoiceService {
	return &InvoiceService{
		invoiceRepo:     invoiceRepo,
		transactionRepo: transactionRepo,
		splitRepo:       splitRepo,
		invoiceItemRepo: invoiceItemRepo,
		itemRepo:        itemRepo,
		accountRepo:     accountRepo,
		db:              db,
	}
}

// CalculateSplitsForInvoice calculates the double-entry accounting splits for an invoice
func (s *InvoiceService) CalculateSplitsForInvoice(req CreateInvoiceRequest, userID int) ([]SplitPreview, error) {
	splits := []SplitPreview{}

	// Get all items with their account information
	itemMap := make(map[int]*items.Item)
	itemIncomeAccounts := make(map[int]*accounts.Account) // Store income accounts by item ID
	itemAssetAccounts := make(map[int]*accounts.Account)  // Store asset accounts by item ID
	for _, itemInput := range req.Items {
		item, _, assetAccount, incomeAccount, _, _, err := s.itemRepo.GetByID(itemInput.ItemID)
		if err != nil {
			return nil, fmt.Errorf("item %d not found: %v", itemInput.ItemID, err)
		}
		itemMap[itemInput.ItemID] = &item
		if incomeAccount != nil {
			itemIncomeAccounts[itemInput.ItemID] = incomeAccount
		}
		if assetAccount != nil {
			itemAssetAccounts[itemInput.ItemID] = assetAccount
		}
	}

	// For invoice (accounts receivable):
	// 1. Debit: Accounts Receivable (or customer account if people_id has account)
	// 2. Credit: Income/Revenue account (from item's income_account)

	// Get accounts for the building
	accountsList, _, _, err := s.accountRepo.GetByBuildingID(req.BuildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %v", err)
	}

	// Get Accounts Receivable account from request
	if req.ARAccountID == nil {
		return nil, fmt.Errorf("A/R account is required")
	}

	var accountsReceivableAccount *accounts.Account
	for i := range accountsList {
		if accountsList[i].ID == *req.ARAccountID {
			accountsReceivableAccount = &accountsList[i]
			break
		}
	}

	if accountsReceivableAccount == nil {
		return nil, fmt.Errorf("A/R account not found")
	}

	// Calculate totals by item type
	discountTotal := 0.0
	paymentTotal := 0.0
	serviceTotalAmount := 0.0
	serviceIncomeByAccount := make(map[int]float64) // For positive amounts (credits)
	serviceDebitByAccount := make(map[int]float64)  // For negative amounts (debits)
	var discountIncomeAccount *accounts.Account
	var paymentAssetAccount *accounts.Account

	for _, itemInput := range req.Items {
		item := itemMap[itemInput.ItemID]

		// Calculate item total using rate from input
		var itemTotal float64
		var rate float64

		// Parse rate from string input
		if itemInput.Rate != nil && *itemInput.Rate != "" {
			if _, err := fmt.Sscanf(*itemInput.Rate, "%f", &rate); err != nil {
				// If parsing fails, use avg_cost as fallback
				rate = item.AvgCost
			}
		} else {
			// If no rate provided, use avg_cost
			rate = item.AvgCost
		}

		// Use manually edited total if provided, otherwise calculate from qty * rate
		if itemInput.Total != nil {
			// Use the manually edited total from the input
			itemTotal = *itemInput.Total
		} else {
			// Calculate total: qty * rate
			// For discount/payment items, qty should be 1 (enforced by frontend, but ensure here too)
			if item.Type == "discount" || item.Type == "payment" {
				// Force qty to 1 for discount/payment items
				itemTotal = math.Abs(rate) // Use absolute value of rate, qty is implicitly 1
			} else if itemInput.Qty != nil {
				itemTotal = *itemInput.Qty * rate
			} else {
				itemTotal = rate
			}
		}

		// Categorize by item type
		if item.Type == "discount" {
			// Use absolute value for discount (even if rate is negative)
			discountTotal += itemTotal // itemTotal is already absolute value
			// Get discount income account from fetched accounts
			if incomeAccount, exists := itemIncomeAccounts[itemInput.ItemID]; exists && incomeAccount != nil {
				discountIncomeAccount = incomeAccount
			}
		} else if item.Type == "payment" {
			// Use absolute value for payment (even if rate is negative)
			paymentTotal += itemTotal // itemTotal is already absolute value
			// Get payment asset account from fetched accounts
			if assetAccount, exists := itemAssetAccounts[itemInput.ItemID]; exists && assetAccount != nil {
				paymentAssetAccount = assetAccount
			}
		} else if item.Type == "service" {
			serviceTotalAmount += itemTotal
			// Group service income by account - handle negative rates as debits
			if incomeAccount, exists := itemIncomeAccounts[itemInput.ItemID]; exists && incomeAccount != nil {
				if itemTotal >= 0 {
					// Positive rate: credit income account
					serviceIncomeByAccount[incomeAccount.ID] += itemTotal
				} else {
					// Negative rate: debit income account
					serviceDebitByAccount[incomeAccount.ID] += math.Abs(itemTotal)
				}
			} else {
				// Service items must have an income account for proper accounting
				return nil, fmt.Errorf("service item '%s' (ID: %d) must have an income account configured", item.Name, item.ID)
			}
		}
	}

	// Use the amount from the request (which matches the amount input field)
	// This ensures the AR split matches exactly what the user entered
	arAmount := req.Amount

	// Create splits
	// Use unit_id from request for all splits
	// 1. Debit or Credit: Accounts Receivable (depending on net amount)
	if arAmount > 0 {
		// Net positive: debit A/R
		splits = append(splits, SplitPreview{
			AccountID:   accountsReceivableAccount.ID,
			AccountName: accountsReceivableAccount.AccountName,
			PeopleID:    req.PeopleID,
			UnitID:      req.UnitID, // Use unit_id from request
			Debit:       &arAmount,
			Credit:      nil,
			Status:      "1", // 1 = active, 0 = inactive/deleted
		})
	} else if arAmount < 0 {
		// Net negative: credit A/R (refund/reversal)
		arCreditAmount := math.Abs(arAmount)
		splits = append(splits, SplitPreview{
			AccountID:   accountsReceivableAccount.ID,
			AccountName: accountsReceivableAccount.AccountName,
			PeopleID:    req.PeopleID,
			UnitID:      req.UnitID, // Use unit_id from request
			Debit:       nil,
			Credit:      &arCreditAmount,
			Status:      "1", // 1 = active, 0 = inactive/deleted
		})
	}

	// 2. Debit: Discount Income Account (if discount items exist)
	// Note: people_id is only set for AR account, not for income/expense accounts
	if discountTotal > 0 && discountIncomeAccount != nil {
		splits = append(splits, SplitPreview{
			AccountID:   discountIncomeAccount.ID,
			AccountName: discountIncomeAccount.AccountName,
			PeopleID:    req.PeopleID, // Link people_id to all splits
			UnitID:      req.UnitID, // Use unit_id from request
			Debit:       &discountTotal,
			Credit:      nil,
			Status:      "1", // 1 = active, 0 = inactive/deleted
		})
	}

	// 3. Debit: Payment Asset Account (if payment items exist)
	// Note: people_id is only set for AR account, not for asset accounts
	if paymentTotal > 0 && paymentAssetAccount != nil {
		splits = append(splits, SplitPreview{
			AccountID:   paymentAssetAccount.ID,
			AccountName: paymentAssetAccount.AccountName,
			PeopleID:    req.PeopleID, // Link people_id to all splits
			UnitID:      req.UnitID, // Use unit_id from request
			Debit:       &paymentTotal,
			Credit:      nil,
			Status:      "1", // 1 = active, 0 = inactive/deleted
		})
	}

	// 4. Credit: Service Income Accounts (positive rates)
	// This should always have entries if we have service items (validated above)
	if len(serviceIncomeByAccount) == 0 && len(serviceDebitByAccount) == 0 && serviceTotalAmount > 0 {
		return nil, fmt.Errorf("no income account found for service items - service items must have an income account configured")
	}

	for accountID, amount := range serviceIncomeByAccount {
		account, _, _, err := s.accountRepo.GetByID(accountID)
		if err != nil {
			return nil, fmt.Errorf("income account %d not found: %v", accountID, err)
		}
		creditAmount := amount
		splits = append(splits, SplitPreview{
			AccountID:   accountID,
			AccountName: account.AccountName,
			PeopleID:    req.PeopleID, // Link people_id to all splits
			UnitID:      req.UnitID, // Use unit_id from request
			Debit:       nil,
			Credit:      &creditAmount,
			Status:      "1", // 1 = active, 0 = inactive/deleted
		})
	}

	// 5. Debit: Service Income Accounts (negative rates - refunds/reversals)
	for accountID, amount := range serviceDebitByAccount {
		account, _, _, err := s.accountRepo.GetByID(accountID)
		if err != nil {
			return nil, fmt.Errorf("income account %d not found: %v", accountID, err)
		}
		debitAmount := amount
		splits = append(splits, SplitPreview{
			AccountID:   accountID,
			AccountName: account.AccountName,
			PeopleID:    req.PeopleID, // Link people_id to all splits
			UnitID:      req.UnitID, // Use unit_id from request
			Debit:       &debitAmount,
			Credit:      nil,
			Status:      "1", // 1 = active, 0 = inactive/deleted
		})
	}

	// If total credits don't match total debits, adjust
	totalDebit := 0.0
	totalCredit := 0.0
	for _, split := range splits {
		if split.Debit != nil {
			totalDebit += *split.Debit
		}
		if split.Credit != nil {
			totalCredit += *split.Credit
		}
	}

	// Balance the splits if needed
	if totalDebit != totalCredit {
		// Adjust the service income account to balance
		if len(serviceIncomeByAccount) > 0 {
			// Find the first service income split and adjust it
			for i := range splits {
				if splits[i].Credit != nil && splits[i].Debit == nil {
					diff := totalDebit - totalCredit
					newCredit := *splits[i].Credit + diff
					splits[i].Credit = &newCredit
					totalCredit = totalDebit // Update total credit
					break
				}
			}
		}
	}

	// Validate: Must have at least 2 splits and be balanced for double-entry accounting
	if len(splits) < 2 {
		return nil, fmt.Errorf("invoice must have at least 2 splits for double-entry accounting, got %d", len(splits))
	}

	if totalDebit != totalCredit {
		return nil, fmt.Errorf("splits are not balanced: total debit %.2f != total credit %.2f", totalDebit, totalCredit)
	}

	return splits, nil
}

// PreviewInvoice calculates and returns the splits that will be created
func (s *InvoiceService) PreviewInvoice(req CreateInvoiceRequest, userID int) (*InvoicePreviewResponse, error) {
	// Validate request
	if req.Amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than 0")
	}

	if len(req.Items) == 0 {
		return nil, fmt.Errorf("invoice must have at least one item")
	}

	// Calculate splits
	splitPreviews, err := s.CalculateSplitsForInvoice(req, userID)
	if err != nil {
		return nil, err
	}

	// Calculate totals
	totalDebit := 0.0
	totalCredit := 0.0
	for _, split := range splitPreviews {
		if split.Debit != nil {
			totalDebit += *split.Debit
		}
		if split.Credit != nil {
			totalCredit += *split.Credit
		}
	}

	isBalanced := totalDebit == totalCredit

	return &InvoicePreviewResponse{
		Invoice:     req,
		Splits:      splitPreviews,
		TotalDebit:  totalDebit,
		TotalCredit: totalCredit,
		IsBalanced:  isBalanced,
	}, nil
}

// CreateInvoice creates the invoice with transaction and splits
// All operations are wrapped in a database transaction to ensure atomicity
func (s *InvoiceService) CreateInvoice(req CreateInvoiceRequest, userID int) (*InvoiceResponse, error) {
	// Check for duplicate invoice number (before transaction)
	exists, err := s.invoiceRepo.CheckDuplicateInvoiceNo(req.BuildingID, req.InvoiceNo, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("invoice number already exists for this building")
	}

	// Start database transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %v", err)
	}

	// Track if transaction was committed to avoid unnecessary rollback
	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()

	// Create transaction record - always use status 1 (active) when creating
	var unitID interface{}
	if req.UnitID != nil {
		unitID = *req.UnitID
	} else {
		unitID = nil
	}

	// Always use status "1" (active) when creating transactions
	transactionStatus := "1"
	result, err := tx.Exec("INSERT INTO transactions (type, transaction_date, transaction_number, memo, status, building_id, user_id, unit_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		"invoice", req.SalesDate, req.InvoiceNo, req.Description, transactionStatus, req.BuildingID, userID, unitID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %v", err)
	}

	transactionID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction ID: %v", err)
	}

	// Create invoice
	var peopleID interface{}
	if req.PeopleID != nil {
		peopleID = *req.PeopleID
	} else {
		peopleID = nil
	}

	var arAccountID interface{}
	if req.ARAccountID != nil {
		arAccountID = *req.ARAccountID
	} else {
		arAccountID = nil
	}

	// Always use status "1" (active) when creating invoices
	invoiceStatus := "1"
	result, err = tx.Exec("INSERT INTO invoices (invoice_no, transaction_id, sales_date, due_date, ar_account_id, unit_id, people_id, user_id, amount, description, cancel_reason, status, building_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		req.InvoiceNo, transactionID, req.SalesDate, req.DueDate, arAccountID, unitID, peopleID, userID, req.Amount, req.Description, nil, invoiceStatus, req.BuildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to create invoice: %v", err)
	}

	invoiceID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice ID: %v", err)
	}

	// Create invoice items
	for _, itemInput := range req.Items {
		item, _, _, _, _, _, err := s.itemRepo.GetByID(itemInput.ItemID)
		if err != nil {
			return nil, fmt.Errorf("item %d not found: %v", itemInput.ItemID, err)
		}

		// Calculate total using rate from input
		var total float64
		var rate float64

		// Parse rate from string input
		if itemInput.Rate != nil && *itemInput.Rate != "" {
			if _, err := fmt.Sscanf(*itemInput.Rate, "%f", &rate); err != nil {
				// If parsing fails, use avg_cost as fallback
				rate = item.AvgCost
			}
		} else {
			// If no rate provided, use avg_cost
			rate = item.AvgCost
		}

		// For discount/payment items, use absolute value of rate (qty is implicitly 1)
		if item.Type == "discount" || item.Type == "payment" {
			total = math.Abs(rate)
		} else if itemInput.Qty != nil {
			total = *itemInput.Qty * rate
		} else {
			total = rate
		}

		var previousValue interface{}
		if itemInput.PreviousValue != nil {
			previousValue = *itemInput.PreviousValue
		} else {
			previousValue = nil
		}

		var currentValue interface{}
		if itemInput.CurrentValue != nil {
			currentValue = *itemInput.CurrentValue
		} else {
			currentValue = nil
		}

		var qty interface{}
		if itemInput.Qty != nil {
			qty = *itemInput.Qty
		} else {
			qty = nil
		}

		var rateStr interface{}
		if itemInput.Rate != nil {
			rateStr = *itemInput.Rate
		} else {
			rateStr = nil
		}

		// Always use status "1" (active) when creating invoice items
		itemStatus := "1"
		_, err = tx.Exec("INSERT INTO invoice_items (invoice_id, item_id, item_name, previous_value, current_value, qty, rate, total, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
			invoiceID, itemInput.ItemID, item.Name, previousValue, currentValue, qty, rateStr, total, itemStatus)
		if err != nil {
			return nil, fmt.Errorf("failed to create invoice item: %v", err)
		}
	}

	// Calculate and create splits
	splitPreviews, err := s.CalculateSplitsForInvoice(req, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate splits: %v", err)
	}

	// Create splits within transaction
	for _, preview := range splitPreviews {
		var peopleIDSplit interface{}
		if preview.PeopleID != nil {
			peopleIDSplit = *preview.PeopleID
		} else {
			peopleIDSplit = nil
		}

		var unitIDSplit interface{}
		if preview.UnitID != nil {
			unitIDSplit = *preview.UnitID
		} else {
			unitIDSplit = nil
		}

		var debit interface{}
		if preview.Debit != nil {
			debit = *preview.Debit
		} else {
			debit = nil
		}

		var credit interface{}
		if preview.Credit != nil {
			credit = *preview.Credit
		} else {
			credit = nil
		}

		// Always set status to "1" (active) when creating splits
		_, err = tx.Exec("INSERT INTO splits (transaction_id, account_id, people_id, unit_id, debit, credit, status) VALUES (?, ?, ?, ?, ?, ?, ?)",
			transactionID, preview.AccountID, peopleIDSplit, unitIDSplit, debit, credit, "1")
		if err != nil {
			return nil, fmt.Errorf("failed to create split: %v", err)
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}
	committed = true

	// Fetch created records after successful commit
	createdTransaction, err := s.transactionRepo.GetByID(int(transactionID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transaction: %v", err)
	}

	createdInvoice, err := s.invoiceRepo.GetByID(int(invoiceID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch invoice: %v", err)
	}

	createdSplits, err := s.splitRepo.GetByTransactionID(int(transactionID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch splits: %v", err)
	}

	createdInvoiceItems, err := s.invoiceItemRepo.GetByInvoiceID(int(invoiceID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch invoice items: %v", err)
	}

	return &InvoiceResponse{
		Invoice:     createdInvoice,
		Items:       createdInvoiceItems,
		Splits:      createdSplits,
		Transaction: createdTransaction,
	}, nil
}

// UpdateInvoice updates the invoice with transaction and splits
// All operations are wrapped in a database transaction to ensure atomicity
// Soft deletes existing invoice_items and splits by setting status='0', then recreates them
func (s *InvoiceService) UpdateInvoice(req UpdateInvoiceRequest, userID int) (*InvoiceResponse, error) {
	// Validate invoice exists
	existingInvoice, err := s.invoiceRepo.GetByID(req.ID)
	if err != nil {
		return nil, fmt.Errorf("invoice not found: %v", err)
	}

	// Validate invoice belongs to the building
	if existingInvoice.BuildingID != req.BuildingID {
		return nil, fmt.Errorf("invoice does not belong to the specified building")
	}

	// Check for duplicate invoice number (excluding current invoice)
	exists, err := s.invoiceRepo.CheckDuplicateInvoiceNo(req.BuildingID, req.InvoiceNo, req.ID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("invoice number already exists for this building")
	}

	// Validate request
	if req.Amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than 0")
	}

	if len(req.Items) == 0 {
		return nil, fmt.Errorf("invoice must have at least one item")
	}

	// Start database transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %v", err)
	}

	// Track if transaction was committed to avoid unnecessary rollback
	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()

	// Update transaction record
	var unitID interface{}
	if req.UnitID != nil {
		unitID = *req.UnitID
	} else {
		unitID = nil
	}

	_, err = tx.Exec("UPDATE transactions SET transaction_date = ?, transaction_number = ?, memo = ? WHERE id = ?",
		req.SalesDate, req.InvoiceNo, req.Description, existingInvoice.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to update transaction: %v", err)
	}

	// Update invoice
	var peopleID interface{}
	if req.PeopleID != nil {
		peopleID = *req.PeopleID
	} else {
		peopleID = nil
	}

	var arAccountID interface{}
	if req.ARAccountID != nil {
		arAccountID = *req.ARAccountID
	} else {
		arAccountID = nil
	}

	_, err = tx.Exec("UPDATE invoices SET invoice_no = ?, sales_date = ?, due_date = ?, ar_account_id = ?, unit_id = ?, people_id = ?, amount = ?, description = ? WHERE id = ?",
		req.InvoiceNo, req.SalesDate, req.DueDate, arAccountID, unitID, peopleID, req.Amount, req.Description, req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update invoice: %v", err)
	}

	// Soft delete existing invoice_items (set status='0')
	_, err = tx.Exec("UPDATE invoice_items SET status = '0' WHERE invoice_id = ?", req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to soft delete invoice items: %v", err)
	}

	// Soft delete existing splits (set status='0')
	_, err = tx.Exec("UPDATE splits SET status = '0' WHERE transaction_id = ?", existingInvoice.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to soft delete splits: %v", err)
	}

	// Recreate invoice items
	for _, itemInput := range req.Items {
		item, _, _, _, _, _, err := s.itemRepo.GetByID(itemInput.ItemID)
		if err != nil {
			return nil, fmt.Errorf("item %d not found: %v", itemInput.ItemID, err)
		}

		// Calculate total using rate from input
		var total float64
		var rate float64

		// Parse rate from string input
		if itemInput.Rate != nil && *itemInput.Rate != "" {
			if _, err := fmt.Sscanf(*itemInput.Rate, "%f", &rate); err != nil {
				// If parsing fails, use avg_cost as fallback
				rate = item.AvgCost
			}
		} else {
			// If no rate provided, use avg_cost
			rate = item.AvgCost
		}

		// Use manually edited total if provided, otherwise calculate from qty * rate
		if itemInput.Total != nil {
			// Use the manually edited total from the input
			total = *itemInput.Total
		} else {
			// Calculate total: qty * rate
			// For discount/payment items, use absolute value of rate (qty is implicitly 1)
			if item.Type == "discount" || item.Type == "payment" {
				total = math.Abs(rate)
			} else if itemInput.Qty != nil {
				total = *itemInput.Qty * rate
			} else {
				total = rate
			}
		}

		var previousValue interface{}
		if itemInput.PreviousValue != nil {
			previousValue = *itemInput.PreviousValue
		} else {
			previousValue = nil
		}

		var currentValue interface{}
		if itemInput.CurrentValue != nil {
			currentValue = *itemInput.CurrentValue
		} else {
			currentValue = nil
		}

		var qty interface{}
		if itemInput.Qty != nil {
			qty = *itemInput.Qty
		} else {
			qty = nil
		}

		var rateStr interface{}
		if itemInput.Rate != nil {
			rateStr = *itemInput.Rate
		} else {
			rateStr = nil
		}

		// Always use status 1 (active) when creating invoice items
		itemStatus := "1"
		_, err = tx.Exec("INSERT INTO invoice_items (invoice_id, item_id, item_name, previous_value, current_value, qty, rate, total, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
			req.ID, itemInput.ItemID, item.Name, previousValue, currentValue, qty, rateStr, total, itemStatus)
		if err != nil {
			return nil, fmt.Errorf("failed to create invoice item: %v", err)
		}
	}

	// Calculate splits (convert UpdateInvoiceRequest to CreateInvoiceRequest for calculation)
	createReq := CreateInvoiceRequest{
		InvoiceNo:   req.InvoiceNo,
		SalesDate:   req.SalesDate,
		DueDate:     req.DueDate,
		UnitID:      req.UnitID,
		PeopleID:    req.PeopleID,
		ARAccountID: req.ARAccountID,
		Amount:      req.Amount,
		Description: req.Description,
		Status:      req.Status,
		BuildingID:  req.BuildingID,
		Items:       req.Items,
	}

	splitPreviews, err := s.CalculateSplitsForInvoice(createReq, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate splits: %v", err)
	}

	// Recreate splits within transaction
	for _, preview := range splitPreviews {
		var peopleIDSplit interface{}
		if preview.PeopleID != nil {
			peopleIDSplit = *preview.PeopleID
		} else {
			peopleIDSplit = nil
		}

		var unitIDSplit interface{}
		if preview.UnitID != nil {
			unitIDSplit = *preview.UnitID
		} else {
			unitIDSplit = nil
		}

		var debit interface{}
		if preview.Debit != nil {
			debit = *preview.Debit
		} else {
			debit = nil
		}

		var credit interface{}
		if preview.Credit != nil {
			credit = *preview.Credit
		} else {
			credit = nil
		}

		// Always set status to "1" (active) when creating splits
		_, err = tx.Exec("INSERT INTO splits (transaction_id, account_id, people_id, unit_id, debit, credit, status) VALUES (?, ?, ?, ?, ?, ?, ?)",
			existingInvoice.TransactionID, preview.AccountID, peopleIDSplit, unitIDSplit, debit, credit, "1")
		if err != nil {
			return nil, fmt.Errorf("failed to create split: %v", err)
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}
	committed = true

	// Fetch updated records after successful commit
	updatedTransaction, err := s.transactionRepo.GetByID(existingInvoice.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transaction: %v", err)
	}

	updatedInvoice, err := s.invoiceRepo.GetByID(req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch invoice: %v", err)
	}

	// Get only active splits (status='1')
	updatedSplits, err := s.splitRepo.GetByTransactionID(existingInvoice.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch splits: %v", err)
	}
	// Filter to only active splits
	activeSplits := []splits.Split{}
	for _, split := range updatedSplits {
		if split.Status == "1" {
			activeSplits = append(activeSplits, split)
		}
	}

	// Get only active invoice items (status='1')
	updatedInvoiceItems, err := s.invoiceItemRepo.GetByInvoiceID(req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch invoice items: %v", err)
	}
	// Filter to only active items
	activeItems := []invoice_items.InvoiceItem{}
	for _, item := range updatedInvoiceItems {
		if item.Status == "1" {
			activeItems = append(activeItems, item)
		}
	}

	return &InvoiceResponse{
		Invoice:     updatedInvoice,
		Items:       activeItems,
		Splits:      activeSplits,
		Transaction: updatedTransaction,
	}, nil
}
