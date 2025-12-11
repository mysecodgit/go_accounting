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
	serviceIncomeByAccount := make(map[int]float64)
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
			// Group service income by account - use fetched income account
			if incomeAccount, exists := itemIncomeAccounts[itemInput.ItemID]; exists && incomeAccount != nil {
				serviceIncomeByAccount[incomeAccount.ID] += itemTotal
			} else {
				// Service items must have an income account for proper accounting
				return nil, fmt.Errorf("service item '%s' (ID: %d) must have an income account configured", item.Name, item.ID)
			}
		}
	}

	// Calculate A/R debit amount = service total - discount - payment
	arDebitAmount := serviceTotalAmount - discountTotal - paymentTotal

	// Create splits
	// 1. Debit: Accounts Receivable (full service amount)
	if arDebitAmount > 0 {
		splits = append(splits, SplitPreview{
			AccountID:   accountsReceivableAccount.ID,
			AccountName: accountsReceivableAccount.AccountName,
			PeopleID:    req.PeopleID,
			Debit:       &arDebitAmount,
			Credit:      nil,
			Status:      "1", // 1 = active, 0 = inactive/deleted
		})
	}

	// 2. Debit: Discount Income Account (if discount items exist)
	if discountTotal > 0 && discountIncomeAccount != nil {
		splits = append(splits, SplitPreview{
			AccountID:   discountIncomeAccount.ID,
			AccountName: discountIncomeAccount.AccountName,
			PeopleID:    req.PeopleID,
			Debit:       &discountTotal,
			Credit:      nil,
			Status:      "1", // 1 = active, 0 = inactive/deleted
		})
	}

	// 3. Debit: Payment Asset Account (if payment items exist)
	if paymentTotal > 0 && paymentAssetAccount != nil {
		splits = append(splits, SplitPreview{
			AccountID:   paymentAssetAccount.ID,
			AccountName: paymentAssetAccount.AccountName,
			PeopleID:    req.PeopleID,
			Debit:       &paymentTotal,
			Credit:      nil,
			Status:      "1", // 1 = active, 0 = inactive/deleted
		})
	}

	// 4. Credit: Service Income Accounts
	// This should always have entries if we have service items (validated above)
	if len(serviceIncomeByAccount) == 0 && serviceTotalAmount > 0 {
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
			PeopleID:    req.PeopleID,
			Debit:       nil,
			Credit:      &creditAmount,
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
	defer tx.Rollback() // Rollback on any error

	// Create transaction record - always use status 1 (active) when creating
	var unitID interface{}
	if req.UnitID != nil {
		unitID = *req.UnitID
	} else {
		unitID = nil
	}

	// Always use status 1 (active) when creating transactions
	transactionStatus := 1
	result, err := tx.Exec("INSERT INTO transactions (type, transaction_date, memo, status, building_id, user_id, unit_id) VALUES (?, ?, ?, ?, ?, ?, ?)",
		"invoice", req.SalesDate, req.Description, transactionStatus, req.BuildingID, userID, unitID)
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

	// Always use status 1 (active) when creating invoices
	invoiceStatus := 1
	result, err = tx.Exec("INSERT INTO invoices (invoice_no, transaction_id, sales_date, due_date, unit_id, people_id, user_id, amount, description, refrence, cancel_reason, status, building_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		req.InvoiceNo, transactionID, req.SalesDate, req.DueDate, unitID, peopleID, userID, req.Amount, req.Description, req.Reference, nil, invoiceStatus, req.BuildingID)
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

		// Always use status 1 (active) when creating invoice items
		itemStatus := 1
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
		_, err = tx.Exec("INSERT INTO splits (transaction_id, account_id, people_id, debit, credit, status) VALUES (?, ?, ?, ?, ?, ?)",
			transactionID, preview.AccountID, peopleIDSplit, debit, credit, "1")
		if err != nil {
			return nil, fmt.Errorf("failed to create split: %v", err)
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

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
