package credit_memo

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/mysecodgit/go_accounting/src/account_types"
	"github.com/mysecodgit/go_accounting/src/accounts"
	"github.com/mysecodgit/go_accounting/src/people"
	"github.com/mysecodgit/go_accounting/src/splits"
	"github.com/mysecodgit/go_accounting/src/transactions"
)

type CreditMemoService struct {
	creditMemoRepo  CreditMemoRepository
	transactionRepo transactions.TransactionRepository
	splitRepo        splits.SplitRepository
	accountRepo      accounts.AccountRepository
	accountTypeRepo  account_types.AccountTypeRepository
	peopleRepo       people.PersonRepository
	db               *sql.DB
}

func NewCreditMemoService(
	creditMemoRepo CreditMemoRepository,
	transactionRepo transactions.TransactionRepository,
	splitRepo splits.SplitRepository,
	accountRepo accounts.AccountRepository,
	accountTypeRepo account_types.AccountTypeRepository,
	peopleRepo people.PersonRepository,
	db *sql.DB,
) *CreditMemoService {
	return &CreditMemoService{
		creditMemoRepo:  creditMemoRepo,
		transactionRepo: transactionRepo,
		splitRepo:       splitRepo,
		accountRepo:     accountRepo,
		accountTypeRepo: accountTypeRepo,
		peopleRepo:      peopleRepo,
		db:              db,
	}
}

// CalculateSplitsForCreditMemo calculates the double-entry accounting splits for a credit memo
// For credit memos: Debit deposit_to account (asset increases), Credit liability account (liability decreases)
func (s *CreditMemoService) CalculateSplitsForCreditMemo(req CreateCreditMemoRequest, userID int) ([]SplitPreview, error) {
	splits := []SplitPreview{}

	if req.Amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than 0")
	}

	// Get liability account
	liabilityAccount, _, _, err := s.accountRepo.GetByID(req.LiabilityAccount)
	if err != nil {
		return nil, fmt.Errorf("liability account not found: %v", err)
	}

	// Get account type for liability account
	liabilityAccountType, err := s.accountTypeRepo.GetByID(liabilityAccount.AccountType)
	if err != nil {
		return nil, fmt.Errorf("liability account type not found: %v", err)
	}

	// Validate: if liability account type is "Account Receivable" or "Account Payable", people_id must be selected
	typeLower := strings.ToLower(liabilityAccountType.Type)
	if (typeLower == "account receivable" || typeLower == "account payable") && req.PeopleID <= 0 {
		return nil, fmt.Errorf("people_id is required when liability account type is %s", liabilityAccountType.TypeName)
	}

	// Get deposit_to account
	depositAccount, _, _, err := s.accountRepo.GetByID(req.DepositTo)
	if err != nil {
		return nil, fmt.Errorf("deposit to account not found: %v", err)
	}

	// Debit: Deposit to account (asset increases)
	debitAmount := req.Amount
	splits = append(splits, SplitPreview{
		AccountID:   req.DepositTo,
		AccountName: depositAccount.AccountName,
		PeopleID:    nil,
		Debit:       &debitAmount,
		Credit:      nil,
		Status:      "1",
	})

	// Credit: Liability account (liability decreases, with people_id if applicable)
	creditAmount := req.Amount
	peopleIDPtr := &req.PeopleID
	splits = append(splits, SplitPreview{
		AccountID:   req.LiabilityAccount,
		AccountName: liabilityAccount.AccountName,
		PeopleID:    peopleIDPtr,
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
		return nil, fmt.Errorf("credit memo must have at least 2 splits for double-entry accounting, got %d", len(splits))
	}

	if totalDebit != totalCredit {
		return nil, fmt.Errorf("splits are not balanced: total debit %.2f != total credit %.2f", totalDebit, totalCredit)
	}

	return splits, nil
}

// PreviewCreditMemo calculates and returns the splits that will be created
func (s *CreditMemoService) PreviewCreditMemo(req CreateCreditMemoRequest, userID int) (*CreditMemoPreviewResponse, error) {
	if req.Amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than 0")
	}

	// Calculate splits
	splitPreviews, err := s.CalculateSplitsForCreditMemo(req, userID)
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

	return &CreditMemoPreviewResponse{
		CreditMemo:  req,
		Splits:      splitPreviews,
		TotalDebit:  totalDebit,
		TotalCredit: totalCredit,
		IsBalanced:  isBalanced,
	}, nil
}

// CreateCreditMemo creates the credit memo with transaction and splits
// All operations are wrapped in a database transaction to ensure atomicity
func (s *CreditMemoService) CreateCreditMemo(req CreateCreditMemoRequest, userID int) (*CreditMemoResponse, error) {
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
	result, err := tx.Exec("INSERT INTO transactions (type, transaction_date, memo, status, building_id, user_id, unit_id) VALUES (?, ?, ?, ?, ?, ?, ?)",
		"credit memo", req.Date, req.Description, transactionStatus, req.BuildingID, userID, req.UnitID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %v", err)
	}

	transactionID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction ID: %v", err)
	}

	// Create credit memo - always use status "1" (active) when creating
	creditMemoStatus := "1"
	result, err = tx.Exec("INSERT INTO credit_memo (transaction_id, date, user_id, deposit_to, liability_account, people_id, building_id, unit_id, amount, description, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		transactionID, req.Date, userID, req.DepositTo, req.LiabilityAccount, req.PeopleID, req.BuildingID, req.UnitID, req.Amount, req.Description, creditMemoStatus)
	if err != nil {
		return nil, fmt.Errorf("failed to create credit memo: %v", err)
	}

	creditMemoID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get credit memo ID: %v", err)
	}

	// Calculate and create splits
	splitPreviews, err := s.CalculateSplitsForCreditMemo(req, userID)
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
	committed = true

	// Fetch created records after successful commit
	createdTransaction, err := s.transactionRepo.GetByID(int(transactionID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transaction: %v", err)
	}

	createdCreditMemo, err := s.creditMemoRepo.GetByID(int(creditMemoID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch credit memo: %v", err)
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

	return &CreditMemoResponse{
		CreditMemo:  createdCreditMemo,
		Splits:      activeSplits,
		Transaction: createdTransaction,
	}, nil
}

// UpdateCreditMemo updates the credit memo with transaction and splits
// All operations are wrapped in a database transaction to ensure atomicity
func (s *CreditMemoService) UpdateCreditMemo(req UpdateCreditMemoRequest, userID int) (*CreditMemoResponse, error) {
	// Get existing credit memo
	existingCreditMemo, err := s.creditMemoRepo.GetByID(req.ID)
	if err != nil {
		return nil, fmt.Errorf("credit memo not found: %v", err)
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

	// Update transaction
	_, err = tx.Exec("UPDATE transactions SET transaction_date = ?, memo = ?, unit_id = ? WHERE id = ?",
		req.Date, req.Description, req.UnitID, existingCreditMemo.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to update transaction: %v", err)
	}

	// Update credit memo
	_, err = tx.Exec("UPDATE credit_memo SET date = ?, deposit_to = ?, liability_account = ?, people_id = ?, unit_id = ?, amount = ?, description = ? WHERE id = ?",
		req.Date, req.DepositTo, req.LiabilityAccount, req.PeopleID, req.UnitID, req.Amount, req.Description, req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update credit memo: %v", err)
	}

	// Soft delete existing splits (set status to '0')
	_, err = tx.Exec("UPDATE splits SET status = '0' WHERE transaction_id = ?", existingCreditMemo.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to soft delete splits: %v", err)
	}

	// Calculate and create new splits
	createReq := CreateCreditMemoRequest{
		Date:             req.Date,
		DepositTo:        req.DepositTo,
		LiabilityAccount: req.LiabilityAccount,
		PeopleID:         req.PeopleID,
		BuildingID:       req.BuildingID,
		UnitID:           req.UnitID,
		Amount:           req.Amount,
		Description:      req.Description,
	}
	splitPreviews, err := s.CalculateSplitsForCreditMemo(createReq, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate splits: %v", err)
	}

	// Create new splits within transaction
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
			existingCreditMemo.TransactionID, preview.AccountID, peopleIDSplit, debit, credit, "1")
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
	updatedTransaction, err := s.transactionRepo.GetByID(existingCreditMemo.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transaction: %v", err)
	}

	updatedCreditMemo, err := s.creditMemoRepo.GetByID(req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch credit memo: %v", err)
	}

	updatedSplits, err := s.splitRepo.GetByTransactionID(existingCreditMemo.TransactionID)
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

	return &CreditMemoResponse{
		CreditMemo:  updatedCreditMemo,
		Splits:      activeSplits,
		Transaction: updatedTransaction,
	}, nil
}

// GetCreditMemoByID retrieves a credit memo by ID with all related data
func (s *CreditMemoService) GetCreditMemoByID(id int) (*CreditMemoResponse, error) {
	creditMemo, err := s.creditMemoRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("credit memo not found: %v", err)
	}

	transaction, err := s.transactionRepo.GetByID(creditMemo.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transaction: %v", err)
	}

	allSplits, err := s.splitRepo.GetByTransactionID(creditMemo.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch splits: %v", err)
	}

	return &CreditMemoResponse{
		CreditMemo:  creditMemo,
		Splits:      allSplits, // Include both active and inactive splits for view
		Transaction: transaction,
	}, nil
}

// GetCreditMemosByBuildingID retrieves all credit memos for a building
func (s *CreditMemoService) GetCreditMemosByBuildingID(buildingID int) ([]CreditMemo, error) {
	return s.creditMemoRepo.GetByBuildingID(buildingID)
}

