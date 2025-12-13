package checks

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/mysecodgit/go_accounting/src/account_types"
	"github.com/mysecodgit/go_accounting/src/accounts"
	"github.com/mysecodgit/go_accounting/src/expense_lines"
	"github.com/mysecodgit/go_accounting/src/splits"
	"github.com/mysecodgit/go_accounting/src/transactions"
)

type CheckService struct {
	checkRepo       CheckRepository
	expenseLineRepo expense_lines.ExpenseLineRepository
	transactionRepo transactions.TransactionRepository
	splitRepo       splits.SplitRepository
	accountRepo     accounts.AccountRepository
	accountTypeRepo account_types.AccountTypeRepository
	db              *sql.DB
}

func NewCheckService(
	checkRepo CheckRepository,
	expenseLineRepo expense_lines.ExpenseLineRepository,
	transactionRepo transactions.TransactionRepository,
	splitRepo splits.SplitRepository,
	accountRepo accounts.AccountRepository,
	accountTypeRepo account_types.AccountTypeRepository,
	db *sql.DB,
) *CheckService {
	return &CheckService{
		checkRepo:       checkRepo,
		expenseLineRepo: expenseLineRepo,
		transactionRepo: transactionRepo,
		splitRepo:       splitRepo,
		accountRepo:     accountRepo,
		accountTypeRepo: accountTypeRepo,
		db:              db,
	}
}

// CalculateSplitsForCheck calculates the double-entry accounting splits for a check
// For checks: Debit expense accounts, Credit payment account
func (s *CheckService) CalculateSplitsForCheck(req CreateCheckRequest, userID int) ([]SplitPreview, error) {
	splits := []SplitPreview{}

	if len(req.ExpenseLines) == 0 {
		return nil, fmt.Errorf("check must have at least one expense line")
	}

	// Get payment account
	paymentAccount, _, _, err := s.accountRepo.GetByID(req.PaymentAccountID)
	if err != nil {
		return nil, fmt.Errorf("payment account not found: %v", err)
	}

	// Validate expense lines and check for A/R or A/P accounts requiring people_id
	for _, expenseLine := range req.ExpenseLines {
		account, _, _, err := s.accountRepo.GetByID(expenseLine.AccountID)
		if err != nil {
			return nil, fmt.Errorf("expense account %d not found: %v", expenseLine.AccountID, err)
		}

		// Get account type
		accountType, err := s.accountTypeRepo.GetByID(account.AccountType)
		if err != nil {
			return nil, fmt.Errorf("account type not found: %v", err)
		}

		// Validate: if account type is "Account Receivable" or "Account Payable", people_id must be selected
		typeLower := strings.ToLower(accountType.Type)
		if (typeLower == "account receivable" || typeLower == "account payable") && expenseLine.PeopleID == nil {
			return nil, fmt.Errorf("people_id is required when account type is %s", accountType.TypeName)
		}

		// Debit: Expense account
		debitAmount := expenseLine.Amount
		splits = append(splits, SplitPreview{
			AccountID:   expenseLine.AccountID,
			AccountName: account.AccountName,
			PeopleID:    expenseLine.PeopleID,
			UnitID:      expenseLine.UnitID, // Include unit_id if provided
			Debit:       &debitAmount,
			Credit:      nil,
			Status:      "1",
		})
	}

	// Credit: Payment account (bank/credit account)
	creditAmount := req.TotalAmount
	splits = append(splits, SplitPreview{
		AccountID:   req.PaymentAccountID,
		AccountName: paymentAccount.AccountName,
		PeopleID:    nil,
		UnitID:      nil, // Payment account doesn't have unit_id
		Debit:       nil,
		Credit:      &creditAmount,
		Status:      "1",
	})

	// Validate: Must have at least 2 splits and be balanced for double-entry accounting
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

	if len(splits) < 2 {
		return nil, fmt.Errorf("check must have at least 2 splits for double-entry accounting, got %d", len(splits))
	}

	if totalDebit != totalCredit {
		return nil, fmt.Errorf("splits are not balanced: total debit %.2f != total credit %.2f", totalDebit, totalCredit)
	}

	return splits, nil
}

// PreviewCheck calculates and returns the splits that will be created
func (s *CheckService) PreviewCheck(req CreateCheckRequest, userID int) (*CheckPreviewResponse, error) {
	if req.TotalAmount <= 0 {
		return nil, fmt.Errorf("total amount must be greater than 0")
	}

	if len(req.ExpenseLines) == 0 {
		return nil, fmt.Errorf("check must have at least one expense line")
	}

	// Calculate splits
	splitPreviews, err := s.CalculateSplitsForCheck(req, userID)
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

	return &CheckPreviewResponse{
		Check:       req,
		Splits:      splitPreviews,
		TotalDebit:  totalDebit,
		TotalCredit: totalCredit,
		IsBalanced:  isBalanced,
	}, nil
}

// CreateCheck creates the check with transaction and splits
// All operations are wrapped in a database transaction to ensure atomicity
func (s *CheckService) CreateCheck(req CreateCheckRequest, userID int) (*CheckResponse, error) {
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

	// Create transaction record - always use status "1" (active) when creating
	transactionStatus := "1"
	memo := ""
	if req.Memo != nil {
		memo = *req.Memo
	}
	var transactionNumber string
	if req.ReferenceNumber != nil {
		transactionNumber = *req.ReferenceNumber
	} else {
		transactionNumber = ""
	}
	result, err := tx.Exec("INSERT INTO transactions (type, transaction_date, transaction_number, memo, status, building_id, user_id, unit_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		"check", req.CheckDate, transactionNumber, memo, transactionStatus, req.BuildingID, userID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %v", err)
	}

	transactionID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction ID: %v", err)
	}

	// Create check
	var referenceNumber interface{}
	if req.ReferenceNumber != nil {
		referenceNumber = *req.ReferenceNumber
	} else {
		referenceNumber = nil
	}

	var memoInterface interface{}
	if req.Memo != nil {
		memoInterface = *req.Memo
	} else {
		memoInterface = nil
	}

	result, err = tx.Exec("INSERT INTO checks (transaction_id, check_date, reference_number, payment_account_id, building_id, memo, total_amount) VALUES (?, ?, ?, ?, ?, ?, ?)",
		transactionID, req.CheckDate, referenceNumber, req.PaymentAccountID, req.BuildingID, memoInterface, req.TotalAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to create check: %v", err)
	}

	checkID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get check ID: %v", err)
	}

	// Create expense lines
	for _, expenseLineInput := range req.ExpenseLines {
		var unitID interface{}
		if expenseLineInput.UnitID != nil {
			unitID = *expenseLineInput.UnitID
		} else {
			unitID = nil
		}

		var peopleID interface{}
		if expenseLineInput.PeopleID != nil {
			peopleID = *expenseLineInput.PeopleID
		} else {
			peopleID = nil
		}

		var description interface{}
		if expenseLineInput.Description != nil {
			description = *expenseLineInput.Description
		} else {
			description = nil
		}

		_, err = tx.Exec("INSERT INTO expense_lines (check_id, account_id, unit_id, people_id, description, amount) VALUES (?, ?, ?, ?, ?, ?)",
			checkID, expenseLineInput.AccountID, unitID, peopleID, description, expenseLineInput.Amount)
		if err != nil {
			return nil, fmt.Errorf("failed to create expense line: %v", err)
		}
	}

	// Calculate and create splits
	splitPreviews, err := s.CalculateSplitsForCheck(req, userID)
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

	createdCheck, err := s.checkRepo.GetByID(int(checkID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch check: %v", err)
	}

	createdExpenseLines, err := s.expenseLineRepo.GetByCheckID(int(checkID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch expense lines: %v", err)
	}

	createdSplits, err := s.splitRepo.GetByTransactionID(int(transactionID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch splits: %v", err)
	}

	// Filter to only active splits
	activeSplits := []splits.Split{}
	for _, split := range createdSplits {
		if split.Status == "1" {
			activeSplits = append(activeSplits, split)
		}
	}

	return &CheckResponse{
		Check:        createdCheck,
		ExpenseLines: createdExpenseLines,
		Splits:       activeSplits,
		Transaction:  createdTransaction,
	}, nil
}

// UpdateCheck updates the check with transaction and splits
// All operations are wrapped in a database transaction to ensure atomicity
func (s *CheckService) UpdateCheck(req UpdateCheckRequest, userID int) (*CheckResponse, error) {
	// Get existing check
	existingCheck, err := s.checkRepo.GetByID(req.ID)
	if err != nil {
		return nil, fmt.Errorf("check not found: %v", err)
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
	memo := ""
	if req.Memo != nil {
		memo = *req.Memo
	}
	var transactionNumber string
	if req.ReferenceNumber != nil {
		transactionNumber = *req.ReferenceNumber
	} else {
		transactionNumber = ""
	}
	_, err = tx.Exec("UPDATE transactions SET transaction_date = ?, transaction_number = ?, memo = ? WHERE id = ?",
		req.CheckDate, transactionNumber, memo, existingCheck.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to update transaction: %v", err)
	}

	// Update check
	var referenceNumber interface{}
	if req.ReferenceNumber != nil {
		referenceNumber = *req.ReferenceNumber
	} else {
		referenceNumber = nil
	}

	var memoInterface interface{}
	if req.Memo != nil {
		memoInterface = *req.Memo
	} else {
		memoInterface = nil
	}

	_, err = tx.Exec("UPDATE checks SET check_date = ?, reference_number = ?, payment_account_id = ?, memo = ?, total_amount = ? WHERE id = ?",
		req.CheckDate, referenceNumber, req.PaymentAccountID, memoInterface, req.TotalAmount, req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update check: %v", err)
	}

	// Soft delete existing expense_lines (set status='0' if there's a status column, otherwise delete)
	// Since expense_lines doesn't have a status column, we'll delete and recreate
	_, err = tx.Exec("DELETE FROM expense_lines WHERE check_id = ?", req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete expense lines: %v", err)
	}

	// Soft delete existing splits (set status='0')
	_, err = tx.Exec("UPDATE splits SET status = '0' WHERE transaction_id = ?", existingCheck.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to soft delete splits: %v", err)
	}

	// Recreate expense lines
	for _, expenseLineInput := range req.ExpenseLines {
		var unitID interface{}
		if expenseLineInput.UnitID != nil {
			unitID = *expenseLineInput.UnitID
		} else {
			unitID = nil
		}

		var peopleID interface{}
		if expenseLineInput.PeopleID != nil {
			peopleID = *expenseLineInput.PeopleID
		} else {
			peopleID = nil
		}

		var description interface{}
		if expenseLineInput.Description != nil {
			description = *expenseLineInput.Description
		} else {
			description = nil
		}

		_, err = tx.Exec("INSERT INTO expense_lines (check_id, account_id, unit_id, people_id, description, amount) VALUES (?, ?, ?, ?, ?, ?)",
			req.ID, expenseLineInput.AccountID, unitID, peopleID, description, expenseLineInput.Amount)
		if err != nil {
			return nil, fmt.Errorf("failed to create expense line: %v", err)
		}
	}

	// Calculate and recreate splits
	createReq := CreateCheckRequest{
		CheckDate:        req.CheckDate,
		ReferenceNumber:  req.ReferenceNumber,
		PaymentAccountID: req.PaymentAccountID,
		BuildingID:       req.BuildingID,
		Memo:             req.Memo,
		TotalAmount:      req.TotalAmount,
		ExpenseLines:     req.ExpenseLines,
	}
	splitPreviews, err := s.CalculateSplitsForCheck(createReq, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate splits: %v", err)
	}

	// Recreate splits
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
			existingCheck.TransactionID, preview.AccountID, peopleIDSplit, unitIDSplit, debit, credit, "1")
		if err != nil {
			return nil, fmt.Errorf("failed to create split: %v", err)
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}
	committed = true

	// Fetch updated records, filtering for active status
	updatedTransaction, err := s.transactionRepo.GetByID(existingCheck.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transaction: %v", err)
	}

	updatedCheck, err := s.checkRepo.GetByID(req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch check: %v", err)
	}

	updatedExpenseLines, err := s.expenseLineRepo.GetByCheckID(req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch expense lines: %v", err)
	}

	updatedSplits, err := s.splitRepo.GetByTransactionID(existingCheck.TransactionID)
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

	return &CheckResponse{
		Check:        updatedCheck,
		ExpenseLines: updatedExpenseLines,
		Splits:       activeSplits,
		Transaction:  updatedTransaction,
	}, nil
}

// GetCheckDetails returns the check with expense lines, splits, and transaction
func (s *CheckService) GetCheckDetails(id int) (*CheckResponse, error) {
	check, err := s.checkRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	expenseLines, err := s.expenseLineRepo.GetByCheckID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch expense lines: %v", err)
	}

	splitsList, err := s.splitRepo.GetByTransactionID(check.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch splits: %v", err)
	}

	transaction, err := s.transactionRepo.GetByID(check.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transaction: %v", err)
	}

	return &CheckResponse{
		Check:        check,
		ExpenseLines: expenseLines,
		Splits:       splitsList,
		Transaction:  transaction,
	}, nil
}

// GetChecks returns all checks for a building
func (s *CheckService) GetChecks(buildingID int) ([]Check, error) {
	return s.checkRepo.GetByBuildingID(buildingID)
}
