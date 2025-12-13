package invoice_applied_credits

import (
	"database/sql"
	"fmt"

	"github.com/mysecodgit/go_accounting/src/accounts"
	"github.com/mysecodgit/go_accounting/src/credit_memo"
	"github.com/mysecodgit/go_accounting/src/invoices"
	"github.com/mysecodgit/go_accounting/src/splits"
	"github.com/mysecodgit/go_accounting/src/transactions"
)

type InvoiceAppliedCreditService struct {
	appliedCreditRepo InvoiceAppliedCreditRepository
	invoiceRepo       invoices.InvoiceRepository
	creditMemoRepo    credit_memo.CreditMemoRepository
	transactionRepo   transactions.TransactionRepository
	splitRepo         splits.SplitRepository
	accountRepo       accounts.AccountRepository
	db                *sql.DB
}

func NewInvoiceAppliedCreditService(
	appliedCreditRepo InvoiceAppliedCreditRepository,
	invoiceRepo invoices.InvoiceRepository,
	creditMemoRepo credit_memo.CreditMemoRepository,
	transactionRepo transactions.TransactionRepository,
	splitRepo splits.SplitRepository,
	accountRepo accounts.AccountRepository,
	db *sql.DB,
) *InvoiceAppliedCreditService {
	return &InvoiceAppliedCreditService{
		appliedCreditRepo: appliedCreditRepo,
		invoiceRepo:       invoiceRepo,
		creditMemoRepo:    creditMemoRepo,
		transactionRepo:   transactionRepo,
		splitRepo:         splitRepo,
		accountRepo:       accountRepo,
		db:                db,
	}
}

// GetAvailableCreditsForInvoice gets all available credit memos for an invoice (matching people_id)
func (s *InvoiceAppliedCreditService) GetAvailableCreditsForInvoice(invoiceID int) (*AvailableCreditsResponse, error) {
	// Get invoice to find people_id
	invoice, err := s.invoiceRepo.GetByID(invoiceID)
	if err != nil {
		return nil, fmt.Errorf("invoice not found: %v", err)
	}

	if invoice.PeopleID == nil {
		return &AvailableCreditsResponse{
			InvoiceID: invoiceID,
			PeopleID:  0,
			Credits:   []AvailableCreditMemo{},
		}, nil
	}

	peopleID := *invoice.PeopleID

	// Get all credit memos for this people_id
	allCreditMemos, err := s.creditMemoRepo.GetByBuildingID(invoice.BuildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch credit memos: %v", err)
	}

	// Filter credit memos by people_id and status = '1'
	availableCredits := []AvailableCreditMemo{}
	for _, creditMemo := range allCreditMemos {
		if creditMemo.PeopleID == peopleID && creditMemo.Status == "1" {
			// Get applied amount for this credit memo
			appliedAmount, err := s.appliedCreditRepo.GetAppliedAmountByCreditMemoID(creditMemo.ID)
			if err != nil {
				continue
			}

			// Calculate available amount and round to 2 decimal places to avoid floating-point precision issues
			availableAmount := creditMemo.Amount - appliedAmount
			// Round to 2 decimal places
			availableAmount = float64(int(availableAmount*100+0.5)) / 100

			if availableAmount > 0 {
				availableCredits = append(availableCredits, AvailableCreditMemo{
					ID:              creditMemo.ID,
					Date:            creditMemo.Date,
					Amount:          creditMemo.Amount,
					AppliedAmount:   appliedAmount,
					AvailableAmount: availableAmount,
					Description:     creditMemo.Description,
				})
			}
		}
	}

	return &AvailableCreditsResponse{
		InvoiceID: invoiceID,
		PeopleID:  peopleID,
		Credits:   availableCredits,
	}, nil
}

// PreviewApplyCredit previews the splits that will be created when applying a credit
func (s *InvoiceAppliedCreditService) PreviewApplyCredit(req CreateInvoiceAppliedCreditRequest) (*InvoiceAppliedCreditPreviewResponse, error) {
	// Validate amount
	if req.Amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than 0")
	}

	// Get invoice
	invoice, err := s.invoiceRepo.GetByID(req.InvoiceID)
	if err != nil {
		return nil, fmt.Errorf("invoice not found: %v", err)
	}

	if invoice.PeopleID == nil {
		return nil, fmt.Errorf("invoice must have a people_id")
	}

	if invoice.ARAccountID == nil {
		return nil, fmt.Errorf("invoice must have an A/R account")
	}

	// Get credit memo
	creditMemo, err := s.creditMemoRepo.GetByID(req.CreditMemoID)
	if err != nil {
		return nil, fmt.Errorf("credit memo not found: %v", err)
	}

	// Validate people_id matches
	if creditMemo.PeopleID != *invoice.PeopleID {
		return nil, fmt.Errorf("credit memo people_id does not match invoice people_id")
	}

	// Check available amount
	appliedAmount, err := s.appliedCreditRepo.GetAppliedAmountByCreditMemoID(req.CreditMemoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get applied amount: %v", err)
	}

	availableAmount := creditMemo.Amount - appliedAmount
	if req.Amount > availableAmount {
		return nil, fmt.Errorf("amount exceeds available credit. Available: %.2f, Requested: %.2f", availableAmount, req.Amount)
	}

	// Get accounts for splits
	arAccount, _, _, err := s.accountRepo.GetByID(*invoice.ARAccountID)
	if err != nil {
		return nil, fmt.Errorf("A/R account not found: %v", err)
	}

	liabilityAccount, _, _, err := s.accountRepo.GetByID(creditMemo.LiabilityAccount)
	if err != nil {
		return nil, fmt.Errorf("liability account not found: %v", err)
	}

	// Create preview splits: Debit liability account, Credit A/R account
	splits := []SplitPreview{}

	// Debit: Liability account (reduces liability)
	debitAmount := req.Amount
	peopleIDPtr := &creditMemo.PeopleID
	splits = append(splits, SplitPreview{
		AccountID:   creditMemo.LiabilityAccount,
		AccountName: liabilityAccount.AccountName,
		PeopleID:    peopleIDPtr,
		Debit:       &debitAmount,
		Credit:      nil,
		Status:      "1",
	})

	// Credit: A/R account (reduces receivable)
	creditAmount := req.Amount
	invoicePeopleIDPtr := invoice.PeopleID
	splits = append(splits, SplitPreview{
		AccountID:   *invoice.ARAccountID,
		AccountName: arAccount.AccountName,
		PeopleID:    invoicePeopleIDPtr,
		Debit:       nil,
		Credit:      &creditAmount,
		Status:      "1",
	})

	// Calculate totals
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

	isBalanced := totalDebit == totalCredit

	return &InvoiceAppliedCreditPreviewResponse{
		AppliedCredit: req,
		Splits:        splits,
		TotalDebit:    totalDebit,
		TotalCredit:   totalCredit,
		IsBalanced:    isBalanced,
	}, nil
}

// ApplyCreditToInvoice applies a credit memo to an invoice with double-entry accounting
// Splits: Debit liability account (from credit memo), Credit A/R account (from invoice)
func (s *InvoiceAppliedCreditService) ApplyCreditToInvoice(req CreateInvoiceAppliedCreditRequest, userID int) (*InvoiceAppliedCreditResponse, error) {
	// Validate amount
	if req.Amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than 0")
	}

	// Get invoice
	invoice, err := s.invoiceRepo.GetByID(req.InvoiceID)
	if err != nil {
		return nil, fmt.Errorf("invoice not found: %v", err)
	}

	if invoice.PeopleID == nil {
		return nil, fmt.Errorf("invoice must have a people_id")
	}

	if invoice.ARAccountID == nil {
		return nil, fmt.Errorf("invoice must have an A/R account")
	}

	// Get credit memo
	creditMemo, err := s.creditMemoRepo.GetByID(req.CreditMemoID)
	if err != nil {
		return nil, fmt.Errorf("credit memo not found: %v", err)
	}

	// Validate people_id matches
	if creditMemo.PeopleID != *invoice.PeopleID {
		return nil, fmt.Errorf("credit memo people_id does not match invoice people_id")
	}

	// Check available amount
	appliedAmount, err := s.appliedCreditRepo.GetAppliedAmountByCreditMemoID(req.CreditMemoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get applied amount: %v", err)
	}

	availableAmount := creditMemo.Amount - appliedAmount
	if req.Amount > availableAmount {
		return nil, fmt.Errorf("amount exceeds available credit. Available: %.2f, Requested: %.2f", availableAmount, req.Amount)
	}

	// Start database transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %v", err)
	}

	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()

	// Create transaction record
	transactionStatus := "1"
	result, err := tx.Exec("INSERT INTO transactions (type, transaction_date, memo, status, building_id, user_id, unit_id) VALUES (?, ?, ?, ?, ?, ?, ?)",
		"credit applied", req.Date, req.Description, transactionStatus, invoice.BuildingID, userID, invoice.UnitID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %v", err)
	}

	transactionID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction ID: %v", err)
	}

	// Create invoice applied credit record (must include transaction_id)
	appliedCreditStatus := "1"
	result, err = tx.Exec("INSERT INTO invoice_applied_credits (transaction_id, invoice_id, credit_memo_id, amount, description, date, status) VALUES (?, ?, ?, ?, ?, ?, ?)",
		transactionID, req.InvoiceID, req.CreditMemoID, req.Amount, req.Description, req.Date, appliedCreditStatus)
	if err != nil {
		return nil, fmt.Errorf("failed to create invoice applied credit: %v", err)
	}

	appliedCreditID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get applied credit ID: %v", err)
	}

	// Validate accounts exist
	_, _, _, err = s.accountRepo.GetByID(*invoice.ARAccountID)
	if err != nil {
		return nil, fmt.Errorf("A/R account not found: %v", err)
	}

	_, _, _, err = s.accountRepo.GetByID(creditMemo.LiabilityAccount)
	if err != nil {
		return nil, fmt.Errorf("liability account not found: %v", err)
	}

	// Create splits: Debit liability account, Credit A/R account
	// Debit: Liability account (reduces liability)
	debitAmount := req.Amount
	peopleIDPtr := &creditMemo.PeopleID
	_, err = tx.Exec("INSERT INTO splits (transaction_id, account_id, people_id, debit, credit, status) VALUES (?, ?, ?, ?, ?, ?)",
		transactionID, creditMemo.LiabilityAccount, peopleIDPtr, debitAmount, nil, "1")
	if err != nil {
		return nil, fmt.Errorf("failed to create debit split: %v", err)
	}

	// Credit: A/R account (reduces receivable)
	creditAmount := req.Amount
	invoicePeopleIDPtr := invoice.PeopleID
	_, err = tx.Exec("INSERT INTO splits (transaction_id, account_id, people_id, debit, credit, status) VALUES (?, ?, ?, ?, ?, ?)",
		transactionID, *invoice.ARAccountID, invoicePeopleIDPtr, nil, creditAmount, "1")
	if err != nil {
		return nil, fmt.Errorf("failed to create credit split: %v", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}
	committed = true

	// Fetch created records
	createdTransaction, err := s.transactionRepo.GetByID(int(transactionID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transaction: %v", err)
	}

	createdAppliedCredit, err := s.appliedCreditRepo.GetByID(int(appliedCreditID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch applied credit: %v", err)
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

	return &InvoiceAppliedCreditResponse{
		InvoiceAppliedCredit: createdAppliedCredit,
		Splits:               activeSplits,
		Transaction:          createdTransaction,
	}, nil
}

// GetAppliedCreditsByInvoiceID gets all applied credits for an invoice
func (s *InvoiceAppliedCreditService) GetAppliedCreditsByInvoiceID(invoiceID int) ([]InvoiceAppliedCredit, error) {
	return s.appliedCreditRepo.GetByInvoiceID(invoiceID)
}

// DeleteAppliedCredit soft deletes an applied credit (sets status to '0')
func (s *InvoiceAppliedCreditService) DeleteAppliedCredit(appliedCreditID int) error {
	// Get the applied credit
	appliedCredit, err := s.appliedCreditRepo.GetByID(appliedCreditID)
	if err != nil {
		return fmt.Errorf("applied credit not found: %v", err)
	}

	// Check if already deleted
	if appliedCredit.Status == "0" {
		return fmt.Errorf("applied credit is already deleted")
	}

	// Start transaction for soft deletion
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}

	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()

	// Soft delete the applied credit
	appliedCredit.Status = "0"
	_, err = tx.Exec("UPDATE invoice_applied_credits SET status = ? WHERE id = ?", appliedCredit.Status, appliedCredit.ID)
	if err != nil {
		return fmt.Errorf("failed to delete applied credit: %v", err)
	}

	// Soft delete the splits (set status to '0')
	_, err = tx.Exec("UPDATE splits SET status = '0' WHERE transaction_id = ?", appliedCredit.TransactionID)
	if err != nil {
		return fmt.Errorf("failed to delete splits: %v", err)
	}

	// Soft delete the transaction
	_, err = tx.Exec("UPDATE transactions SET status = '0' WHERE id = ?", appliedCredit.TransactionID)
	if err != nil {
		return fmt.Errorf("failed to delete transaction: %v", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}
	committed = true

	return nil
}
