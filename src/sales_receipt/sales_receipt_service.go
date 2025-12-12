package sales_receipt

import (
	"database/sql"
	"fmt"
	"math"

	"github.com/mysecodgit/go_accounting/src/accounts"
	"github.com/mysecodgit/go_accounting/src/items"
	"github.com/mysecodgit/go_accounting/src/receipt_items"
	"github.com/mysecodgit/go_accounting/src/splits"
	"github.com/mysecodgit/go_accounting/src/transactions"
)

type SalesReceiptService struct {
	receiptRepo     SalesReceiptRepository
	transactionRepo transactions.TransactionRepository
	splitRepo       splits.SplitRepository
	receiptItemRepo receipt_items.ReceiptItemRepository
	itemRepo        items.ItemRepository
	accountRepo     accounts.AccountRepository
	db              *sql.DB
}

func (s *SalesReceiptService) GetReceiptRepo() SalesReceiptRepository {
	return s.receiptRepo
}

func NewSalesReceiptService(
	receiptRepo SalesReceiptRepository,
	transactionRepo transactions.TransactionRepository,
	splitRepo splits.SplitRepository,
	receiptItemRepo receipt_items.ReceiptItemRepository,
	itemRepo items.ItemRepository,
	accountRepo accounts.AccountRepository,
	db *sql.DB,
) *SalesReceiptService {
	return &SalesReceiptService{
		receiptRepo:     receiptRepo,
		transactionRepo: transactionRepo,
		splitRepo:       splitRepo,
		receiptItemRepo: receiptItemRepo,
		itemRepo:        itemRepo,
		accountRepo:     accountRepo,
		db:              db,
	}
}

// CalculateSplitsForSalesReceipt calculates the double-entry accounting splits for a sales receipt
// For sales receipt:
// 1. Debit: Asset Account (cash/bank account where payment is received)
// 2. Credit: Income/Revenue account (from item's income_account)
func (s *SalesReceiptService) CalculateSplitsForSalesReceipt(req CreateSalesReceiptRequest, userID int) ([]SplitPreview, error) {
	splits := []SplitPreview{}

	// Get all items with their account information
	itemMap := make(map[int]*items.Item)
	itemIncomeAccounts := make(map[int]*accounts.Account)
	itemAssetAccounts := make(map[int]*accounts.Account)
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

	// Get Asset Account from request (cash/bank account)
	assetAccount, _, _, err := s.accountRepo.GetByID(req.AccountID)
	if err != nil {
		return nil, fmt.Errorf("asset account not found: %v", err)
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
				rate = item.AvgCost
			}
		} else {
			rate = item.AvgCost
		}

		// For discount/payment items, use absolute value of rate (qty is implicitly 1)
		if item.Type == "discount" || item.Type == "payment" {
			itemTotal = math.Abs(rate)
		} else if itemInput.Qty != nil {
			itemTotal = *itemInput.Qty * rate
		} else {
			itemTotal = rate
		}

		// Categorize by item type
		if item.Type == "discount" {
			discountTotal += itemTotal
			if incomeAccount, exists := itemIncomeAccounts[itemInput.ItemID]; exists && incomeAccount != nil {
				discountIncomeAccount = incomeAccount
			}
		} else if item.Type == "payment" {
			paymentTotal += itemTotal
			if assetAccountItem, exists := itemAssetAccounts[itemInput.ItemID]; exists && assetAccountItem != nil {
				paymentAssetAccount = assetAccountItem
			}
		} else if item.Type == "service" {
			serviceTotalAmount += itemTotal
			if incomeAccount, exists := itemIncomeAccounts[itemInput.ItemID]; exists && incomeAccount != nil {
				if itemTotal >= 0 {
					// Positive rate: credit income account
					serviceIncomeByAccount[incomeAccount.ID] += itemTotal
				} else {
					// Negative rate: debit income account
					serviceDebitByAccount[incomeAccount.ID] += math.Abs(itemTotal)
				}
			} else {
				return nil, fmt.Errorf("service item '%s' (ID: %d) must have an income account configured", item.Name, item.ID)
			}
		}
	}

	// Calculate Asset Account amount = service total - discount - payment
	assetAmount := serviceTotalAmount - discountTotal - paymentTotal

	// Create splits
	// 1. Debit or Credit: Asset Account (depending on net amount)
	if assetAmount > 0 {
		// Net positive: debit asset account
		splits = append(splits, SplitPreview{
			AccountID:   assetAccount.ID,
			AccountName: assetAccount.AccountName,
			PeopleID:    req.PeopleID,
			Debit:       &assetAmount,
			Credit:      nil,
			Status:      "1",
		})
	} else if assetAmount < 0 {
		// Net negative: credit asset account (refund/reversal)
		assetCreditAmount := math.Abs(assetAmount)
		splits = append(splits, SplitPreview{
			AccountID:   assetAccount.ID,
			AccountName: assetAccount.AccountName,
			PeopleID:    req.PeopleID,
			Debit:       nil,
			Credit:      &assetCreditAmount,
			Status:      "1",
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
			Status:      "1",
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
			Status:      "1",
		})
	}

	// 4. Credit: Service Income Accounts (positive rates)
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
			PeopleID:    req.PeopleID,
			Debit:       nil,
			Credit:      &creditAmount,
			Status:      "1",
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
			PeopleID:    req.PeopleID,
			Debit:       &debitAmount,
			Credit:      nil,
			Status:      "1",
		})
	}

	// Calculate totals and balance
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
		if len(serviceIncomeByAccount) > 0 {
			for i := range splits {
				if splits[i].Credit != nil && splits[i].Debit == nil {
					diff := totalDebit - totalCredit
					newCredit := *splits[i].Credit + diff
					splits[i].Credit = &newCredit
					totalCredit = totalDebit
					break
				}
			}
		}
	}

	// Validate: Must have at least 2 splits and be balanced
	if len(splits) < 2 {
		return nil, fmt.Errorf("sales receipt must have at least 2 splits for double-entry accounting, got %d", len(splits))
	}

	if totalDebit != totalCredit {
		return nil, fmt.Errorf("splits are not balanced: total debit %.2f != total credit %.2f", totalDebit, totalCredit)
	}

	return splits, nil
}

// PreviewSalesReceipt calculates and returns the splits that will be created
func (s *SalesReceiptService) PreviewSalesReceipt(req CreateSalesReceiptRequest, userID int) (*SalesReceiptPreviewResponse, error) {
	if req.Amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than 0")
	}

	if len(req.Items) == 0 {
		return nil, fmt.Errorf("sales receipt must have at least one item")
	}

	splitPreviews, err := s.CalculateSplitsForSalesReceipt(req, userID)
	if err != nil {
		return nil, err
	}

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

	return &SalesReceiptPreviewResponse{
		Receipt:     req,
		Splits:      splitPreviews,
		TotalDebit:  totalDebit,
		TotalCredit: totalCredit,
		IsBalanced:  isBalanced,
	}, nil
}

// CreateSalesReceipt creates the sales receipt with transaction and splits
// All operations are wrapped in a database transaction to ensure atomicity
func (s *SalesReceiptService) CreateSalesReceipt(req CreateSalesReceiptRequest, userID int) (*SalesReceiptResponse, error) {
	// Check for duplicate receipt number
	exists, err := s.receiptRepo.CheckDuplicateReceiptNo(req.BuildingID, req.ReceiptNo, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("receipt number already exists for this building")
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

	// Create transaction record - always use status "1" (active)
	transactionStatus := "1"
	var unitID interface{}
	if req.UnitID != nil {
		unitID = *req.UnitID
	} else {
		unitID = nil
	}

	result, err := tx.Exec("INSERT INTO transactions (type, transaction_date, memo, status, building_id, user_id, unit_id) VALUES (?, ?, ?, ?, ?, ?, ?)",
		"sales receipt", req.ReceiptDate, req.Description, transactionStatus, req.BuildingID, userID, unitID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %v", err)
	}

	transactionID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction ID: %v", err)
	}

	// Create sales receipt - always use status "1" (active)
	receiptStatus := "1"
	var peopleID interface{}
	if req.PeopleID != nil {
		peopleID = *req.PeopleID
	} else {
		peopleID = nil
	}

	result, err = tx.Exec("INSERT INTO sales_receipt (receipt_no, transaction_id, receipt_date, unit_id, people_id, user_id, account_id, amount, description, cancel_reason, status, building_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		req.ReceiptNo, transactionID, req.ReceiptDate, unitID, peopleID, userID, req.AccountID, req.Amount, req.Description, nil, receiptStatus, req.BuildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to create sales receipt: %v", err)
	}

	receiptID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get receipt ID: %v", err)
	}

	// Create receipt items
	for _, itemInput := range req.Items {
		item, _, _, _, _, _, err := s.itemRepo.GetByID(itemInput.ItemID)
		if err != nil {
			return nil, fmt.Errorf("item %d not found: %v", itemInput.ItemID, err)
		}

		var total float64
		var rate float64

		if itemInput.Rate != nil && *itemInput.Rate != "" {
			if _, err := fmt.Sscanf(*itemInput.Rate, "%f", &rate); err != nil {
				rate = item.AvgCost
			}
		} else {
			rate = item.AvgCost
		}

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

		// Always use status "1" (active) when creating receipt items
		itemStatus := "1"
		_, err = tx.Exec("INSERT INTO receipt_items (receipt_id, item_id, item_name, previous_value, current_value, qty, rate, total, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
			receiptID, itemInput.ItemID, item.Name, previousValue, currentValue, qty, rateStr, total, itemStatus)
		if err != nil {
			return nil, fmt.Errorf("failed to create receipt item: %v", err)
		}
	}

	// Calculate and create splits
	splitPreviews, err := s.CalculateSplitsForSalesReceipt(req, userID)
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
	committed = true

	// Fetch created records after successful commit
	createdTransaction, err := s.transactionRepo.GetByID(int(transactionID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transaction: %v", err)
	}

	createdReceipt, err := s.receiptRepo.GetByID(int(receiptID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sales receipt: %v", err)
	}

	createdSplits, err := s.splitRepo.GetByTransactionID(int(transactionID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch splits: %v", err)
	}

	createdReceiptItems, err := s.receiptItemRepo.GetByReceiptID(int(receiptID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch receipt items: %v", err)
	}

	return &SalesReceiptResponse{
		Receipt:     createdReceipt,
		Items:       createdReceiptItems,
		Splits:      createdSplits,
		Transaction: createdTransaction,
	}, nil
}

// UpdateSalesReceipt updates the sales receipt with transaction and splits
// All operations are wrapped in a database transaction to ensure atomicity
// Soft deletes existing receipt_items and splits by setting status='0', then recreates them
func (s *SalesReceiptService) UpdateSalesReceipt(req UpdateSalesReceiptRequest, userID int) (*SalesReceiptResponse, error) {
	// Validate receipt exists
	existingReceipt, err := s.receiptRepo.GetByID(req.ID)
	if err != nil {
		return nil, fmt.Errorf("sales receipt not found: %v", err)
	}

	// Validate receipt belongs to the building
	if existingReceipt.BuildingID != req.BuildingID {
		return nil, fmt.Errorf("sales receipt does not belong to the specified building")
	}

	// Check for duplicate receipt number (excluding current receipt)
	exists, err := s.receiptRepo.CheckDuplicateReceiptNo(req.BuildingID, req.ReceiptNo, req.ID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("receipt number already exists for this building")
	}

	// Validate request
	if req.Amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than 0")
	}

	if len(req.Items) == 0 {
		return nil, fmt.Errorf("sales receipt must have at least one item")
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

	_, err = tx.Exec("UPDATE transactions SET transaction_date = ?, memo = ?, unit_id = ? WHERE id = ?",
		req.ReceiptDate, req.Description, unitID, existingReceipt.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to update transaction: %v", err)
	}

	// Update sales receipt
	var peopleID interface{}
	if req.PeopleID != nil {
		peopleID = *req.PeopleID
	} else {
		peopleID = nil
	}

	_, err = tx.Exec("UPDATE sales_receipt SET receipt_no = ?, receipt_date = ?, unit_id = ?, people_id = ?, account_id = ?, amount = ?, description = ? WHERE id = ?",
		req.ReceiptNo, req.ReceiptDate, unitID, peopleID, req.AccountID, req.Amount, req.Description, req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update sales receipt: %v", err)
	}

	// Soft delete existing receipt_items (set status='0')
	_, err = tx.Exec("UPDATE receipt_items SET status = '0' WHERE receipt_id = ?", req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to soft delete receipt items: %v", err)
	}

	// Soft delete existing splits (set status='0')
	_, err = tx.Exec("UPDATE splits SET status = '0' WHERE transaction_id = ?", existingReceipt.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to soft delete splits: %v", err)
	}

	// Recreate receipt items
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
				rate = item.AvgCost
			}
		} else {
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

		// Always use status "1" (active) when creating receipt items
		itemStatus := "1"
		_, err = tx.Exec("INSERT INTO receipt_items (receipt_id, item_id, item_name, previous_value, current_value, qty, rate, total, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
			req.ID, itemInput.ItemID, item.Name, previousValue, currentValue, qty, rateStr, total, itemStatus)
		if err != nil {
			return nil, fmt.Errorf("failed to create receipt item: %v", err)
		}
	}

	// Calculate splits (convert UpdateSalesReceiptRequest to CreateSalesReceiptRequest for calculation)
	createReq := CreateSalesReceiptRequest{
		ReceiptNo:   req.ReceiptNo,
		ReceiptDate: req.ReceiptDate,
		UnitID:      req.UnitID,
		PeopleID:    req.PeopleID,
		AccountID:   req.AccountID,
		Amount:      req.Amount,
		Description: req.Description,
		Status:      req.Status,
		BuildingID:  req.BuildingID,
		Items:       req.Items,
	}

	splitPreviews, err := s.CalculateSplitsForSalesReceipt(createReq, userID)
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
			existingReceipt.TransactionID, preview.AccountID, peopleIDSplit, debit, credit, "1")
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
	updatedTransaction, err := s.transactionRepo.GetByID(existingReceipt.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transaction: %v", err)
	}

	updatedReceipt, err := s.receiptRepo.GetByID(req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sales receipt: %v", err)
	}

	// Get only active splits (status='1')
	updatedSplits, err := s.splitRepo.GetByTransactionID(existingReceipt.TransactionID)
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

	// Get only active receipt items (status='1')
	updatedReceiptItems, err := s.receiptItemRepo.GetByReceiptID(req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch receipt items: %v", err)
	}
	// Filter to only active items
	activeItems := []receipt_items.ReceiptItem{}
	for _, item := range updatedReceiptItems {
		if item.Status == "1" {
			activeItems = append(activeItems, item)
		}
	}

	return &SalesReceiptResponse{
		Receipt:     updatedReceipt,
		Items:       activeItems,
		Splits:      activeSplits,
		Transaction: updatedTransaction,
	}, nil
}

// GetReceiptItemRepo exposes the ReceiptItemRepository
func (s *SalesReceiptService) GetReceiptItemRepo() receipt_items.ReceiptItemRepository {
	return s.receiptItemRepo
}

// GetSplitRepo exposes the SplitRepository
func (s *SalesReceiptService) GetSplitRepo() splits.SplitRepository {
	return s.splitRepo
}

// GetTransactionRepo exposes the TransactionRepository
func (s *SalesReceiptService) GetTransactionRepo() transactions.TransactionRepository {
	return s.transactionRepo
}
