package journal

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/mysecodgit/go_accounting/src/accounts"
	"github.com/mysecodgit/go_accounting/src/account_types"
	"github.com/mysecodgit/go_accounting/src/journal_lines"
	"github.com/mysecodgit/go_accounting/src/splits"
	"github.com/mysecodgit/go_accounting/src/transactions"
)

type JournalService struct {
	journalRepo     JournalRepository
	journalLineRepo journal_lines.JournalLineRepository
	transactionRepo transactions.TransactionRepository
	splitRepo       splits.SplitRepository
	accountRepo     accounts.AccountRepository
	accountTypeRepo account_types.AccountTypeRepository
	db              *sql.DB
}

func NewJournalService(
	journalRepo JournalRepository,
	journalLineRepo journal_lines.JournalLineRepository,
	transactionRepo transactions.TransactionRepository,
	splitRepo splits.SplitRepository,
	accountRepo accounts.AccountRepository,
	accountTypeRepo account_types.AccountTypeRepository,
	db *sql.DB,
) *JournalService {
	return &JournalService{
		journalRepo:     journalRepo,
		journalLineRepo: journalLineRepo,
		transactionRepo: transactionRepo,
		splitRepo:       splitRepo,
		accountRepo:     accountRepo,
		accountTypeRepo: accountTypeRepo,
		db:              db,
	}
}

// CalculateSplitsForJournal calculates the double-entry accounting splits for a journal
// For journals: Use debit/credit directly from journal_lines
func (s *JournalService) CalculateSplitsForJournal(req CreateJournalRequest, userID int) ([]SplitPreview, error) {
	splits := []SplitPreview{}

	if len(req.Lines) == 0 {
		return nil, fmt.Errorf("journal must have at least one line")
	}

	// Validate journal lines and check for A/R or A/P accounts requiring people_id
	for _, line := range req.Lines {
		account, _, _, err := s.accountRepo.GetByID(line.AccountID)
		if err != nil {
			return nil, fmt.Errorf("account %d not found: %v", line.AccountID, err)
		}

		// Get account type
		accountType, err := s.accountTypeRepo.GetByID(account.AccountType)
		if err != nil {
			return nil, fmt.Errorf("account type not found: %v", err)
		}

		// Validate: if account type is "Account Receivable" or "Account Payable", people_id must be selected
		typeLower := strings.ToLower(accountType.Type)
		if (typeLower == "account receivable" || typeLower == "account payable") && line.PeopleID == nil {
			return nil, fmt.Errorf("people_id is required when account type is %s", accountType.TypeName)
		}

		// Validate: must have either debit or credit, but not both
		if (line.Debit == nil || *line.Debit == 0) && (line.Credit == nil || *line.Credit == 0) {
			return nil, fmt.Errorf("each journal line must have either a debit or credit amount")
		}

		if line.Debit != nil && *line.Debit > 0 && line.Credit != nil && *line.Credit > 0 {
			return nil, fmt.Errorf("journal line cannot have both debit and credit")
		}

		var debitAmount *float64
		var creditAmount *float64

		if line.Debit != nil && *line.Debit > 0 {
			debitAmount = line.Debit
		}

		if line.Credit != nil && *line.Credit > 0 {
			creditAmount = line.Credit
		}

		splits = append(splits, SplitPreview{
			AccountID:   line.AccountID,
			AccountName: account.AccountName,
			PeopleID:    line.PeopleID,
			Debit:       debitAmount,
			Credit:      creditAmount,
			Status:      "1",
		})
	}

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
		return nil, fmt.Errorf("journal must have at least 2 splits for double-entry accounting, got %d", len(splits))
	}

	if totalDebit != totalCredit {
		return nil, fmt.Errorf("splits are not balanced: total debit %.2f != total credit %.2f", totalDebit, totalCredit)
	}

	return splits, nil
}

// PreviewJournal calculates and returns the splits that will be created
func (s *JournalService) PreviewJournal(req CreateJournalRequest, userID int) (*JournalPreviewResponse, error) {
	if len(req.Lines) == 0 {
		return nil, fmt.Errorf("journal must have at least one line")
	}

	// Calculate splits
	splitPreviews, err := s.CalculateSplitsForJournal(req, userID)
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

	return &JournalPreviewResponse{
		Journal:    req,
		Splits:     splitPreviews,
		TotalDebit: totalDebit,
		TotalCredit: totalCredit,
		IsBalanced: isBalanced,
	}, nil
}

// CreateJournal creates the journal with transaction and splits
// All operations are wrapped in a database transaction to ensure atomicity
func (s *JournalService) CreateJournal(req CreateJournalRequest, userID int) (*JournalResponse, error) {
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
	result, err := tx.Exec("INSERT INTO transactions (type, transaction_date, transaction_number, memo, status, building_id, user_id, unit_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		"journal", req.JournalDate, req.Reference, memo, transactionStatus, req.BuildingID, userID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %v", err)
	}

	transactionID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction ID: %v", err)
	}

	// Create journal
	var memoInterface interface{}
	if req.Memo != nil {
		memoInterface = *req.Memo
	} else {
		memoInterface = nil
	}

	result, err = tx.Exec("INSERT INTO journal (transaction_id, reference, journal_date, building_id, memo, total_amount) VALUES (?, ?, ?, ?, ?, ?)",
		transactionID, req.Reference, req.JournalDate, req.BuildingID, memoInterface, req.TotalAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to create journal: %v", err)
	}

	journalID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get journal ID: %v", err)
	}

	// Create journal lines
	for _, lineInput := range req.Lines {
		var unitID interface{}
		if lineInput.UnitID != nil {
			unitID = *lineInput.UnitID
		} else {
			unitID = nil
		}

		var peopleID interface{}
		if lineInput.PeopleID != nil {
			peopleID = *lineInput.PeopleID
		} else {
			peopleID = nil
		}

		var description interface{}
		if lineInput.Description != nil {
			description = *lineInput.Description
		} else {
			description = nil
		}

		var debit interface{}
		if lineInput.Debit != nil {
			debit = *lineInput.Debit
		} else {
			debit = 0
		}

		var credit interface{}
		if lineInput.Credit != nil {
			credit = *lineInput.Credit
		} else {
			credit = 0
		}

		_, err = tx.Exec("INSERT INTO journal_lines (journal_id, account_id, unit_id, people_id, description, debit, credit) VALUES (?, ?, ?, ?, ?, ?, ?)",
			journalID, lineInput.AccountID, unitID, peopleID, description, debit, credit)
		if err != nil {
			return nil, fmt.Errorf("failed to create journal line: %v", err)
		}
	}

	// Calculate and create splits
	splitPreviews, err := s.CalculateSplitsForJournal(req, userID)
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

	createdJournal, err := s.journalRepo.GetByID(int(journalID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch journal: %v", err)
	}

	createdLines, err := s.journalLineRepo.GetByJournalID(int(journalID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch journal lines: %v", err)
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

	return &JournalResponse{
		Journal:    createdJournal,
		Lines:      createdLines,
		Splits:     activeSplits,
		Transaction: createdTransaction,
	}, nil
}

// UpdateJournal updates the journal with transaction and splits
// All operations are wrapped in a database transaction to ensure atomicity
func (s *JournalService) UpdateJournal(req UpdateJournalRequest, userID int) (*JournalResponse, error) {
	// Get existing journal
	existingJournal, err := s.journalRepo.GetByID(req.ID)
	if err != nil {
		return nil, fmt.Errorf("journal not found: %v", err)
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
	_, err = tx.Exec("UPDATE transactions SET transaction_date = ?, transaction_number = ?, memo = ? WHERE id = ?",
		req.JournalDate, req.Reference, memo, existingJournal.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to update transaction: %v", err)
	}

	// Update journal
	var memoInterface interface{}
	if req.Memo != nil {
		memoInterface = *req.Memo
	} else {
		memoInterface = nil
	}

	_, err = tx.Exec("UPDATE journal SET reference = ?, journal_date = ?, memo = ?, total_amount = ? WHERE id = ?",
		req.Reference, req.JournalDate, memoInterface, req.TotalAmount, req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update journal: %v", err)
	}

	// Delete existing journal_lines (no status column)
	_, err = tx.Exec("DELETE FROM journal_lines WHERE journal_id = ?", req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete journal lines: %v", err)
	}

	// Soft delete existing splits (set status='0')
	_, err = tx.Exec("UPDATE splits SET status = '0' WHERE transaction_id = ?", existingJournal.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to soft delete splits: %v", err)
	}

	// Recreate journal lines
	for _, lineInput := range req.Lines {
		var unitID interface{}
		if lineInput.UnitID != nil {
			unitID = *lineInput.UnitID
		} else {
			unitID = nil
		}

		var peopleID interface{}
		if lineInput.PeopleID != nil {
			peopleID = *lineInput.PeopleID
		} else {
			peopleID = nil
		}

		var description interface{}
		if lineInput.Description != nil {
			description = *lineInput.Description
		} else {
			description = nil
		}

		var debit interface{}
		if lineInput.Debit != nil {
			debit = *lineInput.Debit
		} else {
			debit = 0
		}

		var credit interface{}
		if lineInput.Credit != nil {
			credit = *lineInput.Credit
		} else {
			credit = 0
		}

		_, err = tx.Exec("INSERT INTO journal_lines (journal_id, account_id, unit_id, people_id, description, debit, credit) VALUES (?, ?, ?, ?, ?, ?, ?)",
			req.ID, lineInput.AccountID, unitID, peopleID, description, debit, credit)
		if err != nil {
			return nil, fmt.Errorf("failed to create journal line: %v", err)
		}
	}

	// Calculate and recreate splits
	createReq := CreateJournalRequest{
		JournalDate: req.JournalDate,
		BuildingID:  req.BuildingID,
		Memo:        req.Memo,
		TotalAmount: req.TotalAmount,
		Lines:       req.Lines,
	}
	splitPreviews, err := s.CalculateSplitsForJournal(createReq, userID)
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
			existingJournal.TransactionID, preview.AccountID, peopleIDSplit, debit, credit, "1")
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
	updatedTransaction, err := s.transactionRepo.GetByID(existingJournal.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transaction: %v", err)
	}

	updatedJournal, err := s.journalRepo.GetByID(req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch journal: %v", err)
	}

	updatedLines, err := s.journalLineRepo.GetByJournalID(req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch journal lines: %v", err)
	}

	updatedSplits, err := s.splitRepo.GetByTransactionID(existingJournal.TransactionID)
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

	return &JournalResponse{
		Journal:    updatedJournal,
		Lines:      updatedLines,
		Splits:     activeSplits,
		Transaction: updatedTransaction,
	}, nil
}

// GetJournalDetails returns the journal with lines, splits, and transaction
func (s *JournalService) GetJournalDetails(id int) (*JournalResponse, error) {
	journal, err := s.journalRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	lines, err := s.journalLineRepo.GetByJournalID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch journal lines: %v", err)
	}

	splitsList, err := s.splitRepo.GetByTransactionID(journal.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch splits: %v", err)
	}

	transaction, err := s.transactionRepo.GetByID(journal.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transaction: %v", err)
	}

	return &JournalResponse{
		Journal:    journal,
		Lines:      lines,
		Splits:     splitsList,
		Transaction: transaction,
	}, nil
}

// GetJournals returns all journals for a building
func (s *JournalService) GetJournals(buildingID int) ([]Journal, error) {
	return s.journalRepo.GetByBuildingID(buildingID)
}

