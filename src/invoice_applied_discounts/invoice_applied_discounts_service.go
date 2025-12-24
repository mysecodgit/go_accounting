package invoice_applied_discounts

import (
	"database/sql"
	"fmt"

	"github.com/mysecodgit/go_accounting/src/accounts"
	"github.com/mysecodgit/go_accounting/src/invoices"
	"github.com/mysecodgit/go_accounting/src/splits"
	"github.com/mysecodgit/go_accounting/src/transactions"
)

type InvoiceAppliedDiscountService struct {
	appliedDiscountRepo InvoiceAppliedDiscountRepository
	invoiceRepo         invoices.InvoiceRepository
	accountRepo         accounts.AccountRepository
	transactionRepo     transactions.TransactionRepository
	splitRepo           splits.SplitRepository
	db                  *sql.DB
}

func NewInvoiceAppliedDiscountService(
	appliedDiscountRepo InvoiceAppliedDiscountRepository,
	invoiceRepo invoices.InvoiceRepository,
	accountRepo accounts.AccountRepository,
	transactionRepo transactions.TransactionRepository,
	splitRepo splits.SplitRepository,
	db *sql.DB,
) *InvoiceAppliedDiscountService {
	return &InvoiceAppliedDiscountService{
		appliedDiscountRepo: appliedDiscountRepo,
		invoiceRepo:         invoiceRepo,
		accountRepo:         accountRepo,
		transactionRepo:     transactionRepo,
		splitRepo:           splitRepo,
		db:                  db,
	}
}

// PreviewApplyDiscount previews the splits that will be created when applying a discount
func (s *InvoiceAppliedDiscountService) PreviewApplyDiscount(req CreateInvoiceAppliedDiscountRequest) (*InvoiceAppliedDiscountPreviewResponse, error) {
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

	// Validate A/R account matches
	if *invoice.ARAccountID != req.ARAccount {
		return nil, fmt.Errorf("A/R account does not match invoice A/R account")
	}

	// Get accounts for splits
	arAccount, _, _, err := s.accountRepo.GetByID(req.ARAccount)
	if err != nil {
		return nil, fmt.Errorf("A/R account not found: %v", err)
	}

	incomeAccount, _, _, err := s.accountRepo.GetByID(req.IncomeAccount)
	if err != nil {
		return nil, fmt.Errorf("income account not found: %v", err)
	}

	// Create preview splits: Debit Income Account, Credit A/R Account
	splitPreviews := []SplitPreview{}

	// Debit: Income Account (reduces income)
	debitAmount := req.Amount
	invoicePeopleIDPtr := invoice.PeopleID
	invoiceUnitIDPtr := invoice.UnitID
	splitPreviews = append(splitPreviews, SplitPreview{
		AccountID:   req.IncomeAccount,
		AccountName: incomeAccount.AccountName,
		PeopleID:    invoicePeopleIDPtr,
		UnitID:      invoiceUnitIDPtr,
		Debit:       &debitAmount,
		Credit:      nil,
		Status:      "1",
	})

	// Credit: A/R Account (reduces receivable)
	creditAmount := req.Amount
	splitPreviews = append(splitPreviews, SplitPreview{
		AccountID:   req.ARAccount,
		AccountName: arAccount.AccountName,
		PeopleID:    invoicePeopleIDPtr,
		UnitID:      invoiceUnitIDPtr,
		Debit:       nil,
		Credit:      &creditAmount,
		Status:      "1",
	})

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

	return &InvoiceAppliedDiscountPreviewResponse{
		AppliedDiscount: req,
		Splits:          splitPreviews,
		TotalDebit:      totalDebit,
		TotalCredit:     totalCredit,
		IsBalanced:      isBalanced,
	}, nil
}

// ApplyDiscountToInvoice applies a discount to an invoice with double-entry accounting
// Splits: Debit Income Account, Credit A/R Account
func (s *InvoiceAppliedDiscountService) ApplyDiscountToInvoice(req CreateInvoiceAppliedDiscountRequest, userID int) (*InvoiceAppliedDiscountResponse, error) {
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

	// Validate A/R account matches
	if *invoice.ARAccountID != req.ARAccount {
		return nil, fmt.Errorf("A/R account does not match invoice A/R account")
	}

	// Validate accounts exist
	_, _, _, err = s.accountRepo.GetByID(req.ARAccount)
	if err != nil {
		return nil, fmt.Errorf("A/R account not found: %v", err)
	}

	_, _, _, err = s.accountRepo.GetByID(req.IncomeAccount)
	if err != nil {
		return nil, fmt.Errorf("income account not found: %v", err)
	}

	// Start database transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	// Create transaction record
	transactionStatus := "1"
	var unitID interface{}
	if invoice.UnitID != nil {
		unitID = *invoice.UnitID
	} else {
		unitID = nil
	}

	transactionMemo := fmt.Sprintf("Discount applied to Invoice #%s: %s", invoice.InvoiceNo, req.Description)
	result, err := tx.Exec("INSERT INTO transactions (type, transaction_date, transaction_number, memo, status, building_id, user_id, unit_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		"payment", req.Date, "", transactionMemo, transactionStatus, invoice.BuildingID, userID, unitID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %v", err)
	}

	transactionID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction ID: %v", err)
	}

	// Create splits for double-entry accounting
	// 1. Debit: Income Account (reduces income)
	debitAmount := req.Amount
	// 2. Credit: A/R Account (reduces receivable)
	creditAmount := req.Amount

	var peopleID interface{}
	if invoice.PeopleID != nil {
		peopleID = *invoice.PeopleID
	} else {
		peopleID = nil
	}

	// Debit Income Account
	_, err = tx.Exec("INSERT INTO splits (transaction_id, account_id, people_id, unit_id, debit, credit, status) VALUES (?, ?, ?, ?, ?, ?, ?)",
		transactionID, req.IncomeAccount, peopleID, unitID, debitAmount, nil, "1")
	if err != nil {
		return nil, fmt.Errorf("failed to create income debit split: %v", err)
	}

	// Credit A/R Account
	_, err = tx.Exec("INSERT INTO splits (transaction_id, account_id, people_id, unit_id, debit, credit, status) VALUES (?, ?, ?, ?, ?, ?, ?)",
		transactionID, req.ARAccount, peopleID, unitID, nil, creditAmount, "1")
	if err != nil {
		return nil, fmt.Errorf("failed to create A/R credit split: %v", err)
	}

	// Create invoice applied discount record
	appliedDiscountStatus := "1"
	appliedDiscount := InvoiceAppliedDiscount{
		InvoiceID:     req.InvoiceID,
		TransactionID: int(transactionID),
		ARAccount:     req.ARAccount,
		IncomeAccount: req.IncomeAccount,
		Amount:        req.Amount,
		Description:   req.Description,
		Date:          req.Date,
		Status:        appliedDiscountStatus,
		Reference:     req.Reference,
	}

	// Insert applied discount
	_, err = tx.Exec("INSERT INTO invoice_applied_discounts (invoice_id, transaction_id, ar_account, income_account, amount, description, date, status, reference) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		appliedDiscount.InvoiceID, appliedDiscount.TransactionID, appliedDiscount.ARAccount, appliedDiscount.IncomeAccount, appliedDiscount.Amount, appliedDiscount.Description, appliedDiscount.Date, appliedDiscount.Status, appliedDiscount.Reference)
	if err != nil {
		return nil, fmt.Errorf("failed to create invoice applied discount: %v", err)
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

	createdSplits, err := s.splitRepo.GetByTransactionID(int(transactionID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch splits: %v", err)
	}

	// Fetch the created applied discount
	createdAppliedDiscount, err := s.appliedDiscountRepo.GetByTransactionID(int(transactionID))
	if err != nil || len(createdAppliedDiscount) == 0 {
		return nil, fmt.Errorf("failed to fetch invoice applied discount: %v", err)
	}

	return &InvoiceAppliedDiscountResponse{
		InvoiceAppliedDiscount: createdAppliedDiscount[0],
		Splits:                 createdSplits,
		Transaction:            createdTransaction,
	}, nil
}

// GetAppliedDiscountsByInvoiceID gets all applied discounts for an invoice
func (s *InvoiceAppliedDiscountService) GetAppliedDiscountsByInvoiceID(invoiceID int) ([]InvoiceAppliedDiscount, error) {
	return s.appliedDiscountRepo.GetByInvoiceID(invoiceID)
}

// DeleteAppliedDiscount soft deletes an applied discount (sets status to '0')
func (s *InvoiceAppliedDiscountService) DeleteAppliedDiscount(appliedDiscountID int) error {
	// Get the applied discount
	appliedDiscount, err := s.appliedDiscountRepo.GetByID(appliedDiscountID)
	if err != nil {
		return fmt.Errorf("applied discount not found: %v", err)
	}

	// Check if already deleted
	if appliedDiscount.Status == "0" {
		return fmt.Errorf("applied discount is already deleted")
	}

	// Soft delete the applied discount
	appliedDiscount.Status = "0"
	_, err = s.appliedDiscountRepo.Update(appliedDiscount)
	if err != nil {
		return fmt.Errorf("failed to delete applied discount: %v", err)
	}

	// Also soft delete the transaction and splits
	transaction, err := s.transactionRepo.GetByID(appliedDiscount.TransactionID)
	if err == nil {
		// Update transaction status to '0'
		_, err = s.db.Exec("UPDATE transactions SET status = '0' WHERE id = ?", transaction.ID)
		if err != nil {
			return fmt.Errorf("failed to delete transaction: %v", err)
		}

		// Update splits status to '0'
		_, err = s.db.Exec("UPDATE splits SET status = '0' WHERE transaction_id = ?", transaction.ID)
		if err != nil {
			return fmt.Errorf("failed to delete splits: %v", err)
		}
	}

	return nil
}
