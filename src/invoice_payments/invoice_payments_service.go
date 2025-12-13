package invoice_payments

import (
	"database/sql"
	"fmt"

	"github.com/mysecodgit/go_accounting/src/accounts"
	"github.com/mysecodgit/go_accounting/src/invoices"
	"github.com/mysecodgit/go_accounting/src/splits"
	"github.com/mysecodgit/go_accounting/src/transactions"
)

type InvoicePaymentService struct {
	paymentRepo     InvoicePaymentRepository
	transactionRepo transactions.TransactionRepository
	splitRepo       splits.SplitRepository
	invoiceRepo     invoices.InvoiceRepository
	accountRepo     accounts.AccountRepository
	db              *sql.DB
}

func (s *InvoicePaymentService) GetPaymentRepo() InvoicePaymentRepository {
	return s.paymentRepo
}

func NewInvoicePaymentService(
	paymentRepo InvoicePaymentRepository,
	transactionRepo transactions.TransactionRepository,
	splitRepo splits.SplitRepository,
	invoiceRepo invoices.InvoiceRepository,
	accountRepo accounts.AccountRepository,
	db *sql.DB,
) *InvoicePaymentService {
	return &InvoicePaymentService{
		paymentRepo:     paymentRepo,
		transactionRepo: transactionRepo,
		splitRepo:       splitRepo,
		invoiceRepo:     invoiceRepo,
		accountRepo:     accountRepo,
		db:              db,
	}
}

// CreateInvoicePayment creates an invoice payment with transaction and splits
// Double-entry accounting:
// 1. Debit: Asset Account (cash/bank account where payment is received)
// 2. Credit: Accounts Receivable (reducing the amount owed)
func (s *InvoicePaymentService) CreateInvoicePayment(req CreateInvoicePaymentRequest, userID int) (*InvoicePaymentResponse, error) {
	// Validate invoice exists
	invoice, err := s.invoiceRepo.GetByID(req.InvoiceID)
	if err != nil {
		return nil, fmt.Errorf("invoice not found: %v", err)
	}

	// Validate invoice belongs to the building
	if invoice.BuildingID != req.BuildingID {
		return nil, fmt.Errorf("invoice does not belong to the specified building")
	}

	// Get AR account from invoice
	if invoice.ARAccountID == nil {
		return nil, fmt.Errorf("invoice does not have an A/R account configured")
	}

	arAccount, _, _, err := s.accountRepo.GetByID(*invoice.ARAccountID)
	if err != nil {
		return nil, fmt.Errorf("A/R account not found: %v", err)
	}

	// Get Asset Account from request
	assetAccount, _, _, err := s.accountRepo.GetByID(req.AccountID)
	if err != nil {
		return nil, fmt.Errorf("asset account not found: %v", err)
	}

	// Validate amount
	if req.Amount == 0 {
		return nil, fmt.Errorf("amount cannot be zero")
	}

	// Start database transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	// Create transaction record - always use status 1 (active)
	transactionStatus := 1
	var unitID interface{}
	if invoice.UnitID != nil {
		unitID = *invoice.UnitID
	} else {
		unitID = nil
	}

	result, err := tx.Exec("INSERT INTO transactions (type, transaction_date, memo, status, building_id, user_id, unit_id) VALUES (?, ?, ?, ?, ?, ?, ?)",
		"invoice payment", req.Date, fmt.Sprintf("Payment for Invoice #%d", invoice.InvoiceNo), transactionStatus, req.BuildingID, userID, unitID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %v", err)
	}

	transactionID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction ID: %v", err)
	}

	// Create invoice payment - always use status '1' (active)
	paymentStatus := "1"
	result, err = tx.Exec("INSERT INTO invoice_payments (transaction_id, date, invoice_id, user_id, account_id, amount, status) VALUES (?, ?, ?, ?, ?, ?, ?)",
		transactionID, req.Date, req.InvoiceID, userID, req.AccountID, req.Amount, paymentStatus)
	if err != nil {
		return nil, fmt.Errorf("failed to create invoice payment: %v", err)
	}

	paymentID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get payment ID: %v", err)
	}

	// Create splits for double-entry accounting
	// 1. Debit: Asset Account (cash/bank)
	debitAmount := req.Amount
	if debitAmount < 0 {
		debitAmount = -debitAmount // Use absolute value for debit
	}

	// 2. Credit: Accounts Receivable
	creditAmount := req.Amount
	if creditAmount < 0 {
		creditAmount = -creditAmount // Use absolute value for credit
	}

	var peopleID interface{}
	if invoice.PeopleID != nil {
		peopleID = *invoice.PeopleID
	} else {
		peopleID = nil
	}

	// Handle positive and negative amounts
	if req.Amount > 0 {
		// Normal payment: Debit Asset, Credit A/R
		_, err = tx.Exec("INSERT INTO splits (transaction_id, account_id, people_id, debit, credit, status) VALUES (?, ?, ?, ?, ?, ?)",
			transactionID, assetAccount.ID, peopleID, debitAmount, nil, "1")
		if err != nil {
			return nil, fmt.Errorf("failed to create asset debit split: %v", err)
		}

		_, err = tx.Exec("INSERT INTO splits (transaction_id, account_id, people_id, debit, credit, status) VALUES (?, ?, ?, ?, ?, ?)",
			transactionID, arAccount.ID, peopleID, nil, creditAmount, "1")
		if err != nil {
			return nil, fmt.Errorf("failed to create A/R credit split: %v", err)
		}
	} else {
		// Refund/reversal: Credit Asset, Debit A/R
		_, err = tx.Exec("INSERT INTO splits (transaction_id, account_id, people_id, debit, credit, status) VALUES (?, ?, ?, ?, ?, ?)",
			transactionID, assetAccount.ID, peopleID, nil, creditAmount, "1")
		if err != nil {
			return nil, fmt.Errorf("failed to create asset credit split: %v", err)
		}

		_, err = tx.Exec("INSERT INTO splits (transaction_id, account_id, people_id, debit, credit, status) VALUES (?, ?, ?, ?, ?, ?)",
			transactionID, arAccount.ID, peopleID, debitAmount, nil, "1")
		if err != nil {
			return nil, fmt.Errorf("failed to create A/R debit split: %v", err)
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

	createdPayment, err := s.paymentRepo.GetByID(int(paymentID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch invoice payment: %v", err)
	}

	createdSplits, err := s.splitRepo.GetByTransactionID(int(transactionID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch splits: %v", err)
	}

	return &InvoicePaymentResponse{
		Payment:     createdPayment,
		Splits:      createdSplits,
		Transaction: createdTransaction,
		Invoice:     invoice,
		ARAccount:   &arAccount,
	}, nil
}

// PreviewInvoicePayment calculates and returns the splits that will be created when recording a payment
func (s *InvoicePaymentService) PreviewInvoicePayment(req CreateInvoicePaymentRequest) (*InvoicePaymentPreviewResponse, error) {
	// Validate invoice exists
	invoice, err := s.invoiceRepo.GetByID(req.InvoiceID)
	if err != nil {
		return nil, fmt.Errorf("invoice not found: %v", err)
	}

	// Validate invoice belongs to the building
	if invoice.BuildingID != req.BuildingID {
		return nil, fmt.Errorf("invoice does not belong to the specified building")
	}

	// Get AR account from invoice
	if invoice.ARAccountID == nil {
		return nil, fmt.Errorf("invoice does not have an A/R account configured")
	}

	arAccount, _, _, err := s.accountRepo.GetByID(*invoice.ARAccountID)
	if err != nil {
		return nil, fmt.Errorf("A/R account not found: %v", err)
	}

	// Get Asset Account from request
	assetAccount, _, _, err := s.accountRepo.GetByID(req.AccountID)
	if err != nil {
		return nil, fmt.Errorf("asset account not found: %v", err)
	}

	// Validate amount
	if req.Amount == 0 {
		return nil, fmt.Errorf("amount cannot be zero")
	}

	// Create preview splits
	splits := []SplitPreview{}

	debitAmount := req.Amount
	if debitAmount < 0 {
		debitAmount = -debitAmount
	}

	creditAmount := req.Amount
	if creditAmount < 0 {
		creditAmount = -creditAmount
	}

	var peopleID *int
	if invoice.PeopleID != nil {
		peopleID = invoice.PeopleID
	}

	// Handle positive and negative amounts
	if req.Amount > 0 {
		// Normal payment: Debit Asset, Credit A/R
		splits = append(splits, SplitPreview{
			AccountID:   assetAccount.ID,
			AccountName: assetAccount.AccountName,
			PeopleID:     peopleID,
			Debit:        &debitAmount,
			Credit:       nil,
			Status:       "1",
		})

		splits = append(splits, SplitPreview{
			AccountID:   arAccount.ID,
			AccountName: arAccount.AccountName,
			PeopleID:     peopleID,
			Debit:        nil,
			Credit:       &creditAmount,
			Status:       "1",
		})
	} else {
		// Refund/reversal: Credit Asset, Debit A/R
		splits = append(splits, SplitPreview{
			AccountID:   assetAccount.ID,
			AccountName: assetAccount.AccountName,
			PeopleID:     peopleID,
			Debit:        nil,
			Credit:       &creditAmount,
			Status:       "1",
		})

		splits = append(splits, SplitPreview{
			AccountID:   arAccount.ID,
			AccountName: arAccount.AccountName,
			PeopleID:     peopleID,
			Debit:        &debitAmount,
			Credit:       nil,
			Status:       "1",
		})
	}

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

	return &InvoicePaymentPreviewResponse{
		Splits:      splits,
		TotalDebit:  totalDebit,
		TotalCredit: totalCredit,
		IsBalanced:  isBalanced,
	}, nil
}

// UpdateInvoicePayment updates an invoice payment and its related transaction and splits
func (s *InvoicePaymentService) UpdateInvoicePayment(paymentID int, req UpdateInvoicePaymentRequest, userID int) (*InvoicePaymentResponse, error) {
	// Get existing payment
	existingPayment, err := s.paymentRepo.GetByID(paymentID)
	if err != nil {
		return nil, fmt.Errorf("invoice payment not found: %v", err)
	}

	// Get invoice to validate building
	invoice, err := s.invoiceRepo.GetByID(existingPayment.InvoiceID)
	if err != nil {
		return nil, fmt.Errorf("invoice not found: %v", err)
	}

	if invoice.BuildingID != req.BuildingID {
		return nil, fmt.Errorf("invoice does not belong to the specified building")
	}

	// Get AR account from invoice
	if invoice.ARAccountID == nil {
		return nil, fmt.Errorf("invoice does not have an A/R account configured")
	}

	arAccount, _, _, err := s.accountRepo.GetByID(*invoice.ARAccountID)
	if err != nil {
		return nil, fmt.Errorf("A/R account not found: %v", err)
	}

	// Get Asset Account from request
	assetAccount, _, _, err := s.accountRepo.GetByID(req.AccountID)
	if err != nil {
		return nil, fmt.Errorf("asset account not found: %v", err)
	}

	// Validate amount
	if req.Amount == 0 {
		return nil, fmt.Errorf("amount cannot be zero")
	}

	// Start database transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	// Update transaction memo and date
	_, err = tx.Exec("UPDATE transactions SET transaction_date = ?, memo = ? WHERE id = ?",
		req.Date, fmt.Sprintf("Payment for Invoice #%d", invoice.InvoiceNo), existingPayment.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to update transaction: %v", err)
	}

	// Update invoice payment
	paymentStatus := "1"
	if req.Status != nil {
		if *req.Status == 1 {
			paymentStatus = "1"
		} else {
			paymentStatus = "0"
		}
	}

	updatedPayment := existingPayment
	updatedPayment.Date = req.Date
	updatedPayment.AccountID = req.AccountID
	updatedPayment.Amount = req.Amount
	// Convert string status to int for the struct
	if paymentStatus == "1" {
		updatedPayment.Status = 1
	} else {
		updatedPayment.Status = 0
	}

	_, err = tx.Exec("UPDATE invoice_payments SET date = ?, account_id = ?, amount = ?, status = ? WHERE id = ?",
		updatedPayment.Date, updatedPayment.AccountID, updatedPayment.Amount, paymentStatus, updatedPayment.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update invoice payment: %v", err)
	}

	// Soft delete existing splits
	_, err = tx.Exec("UPDATE splits SET status = '0' WHERE transaction_id = ?", existingPayment.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to soft delete existing splits: %v", err)
	}

	// Create new splits for double-entry accounting
	debitAmount := req.Amount
	if debitAmount < 0 {
		debitAmount = -debitAmount
	}

	creditAmount := req.Amount
	if creditAmount < 0 {
		creditAmount = -creditAmount
	}

	var peopleID interface{}
	if invoice.PeopleID != nil {
		peopleID = *invoice.PeopleID
	} else {
		peopleID = nil
	}

	// Handle positive and negative amounts
	if req.Amount > 0 {
		// Normal payment: Debit Asset, Credit A/R
		_, err = tx.Exec("INSERT INTO splits (transaction_id, account_id, people_id, debit, credit, status) VALUES (?, ?, ?, ?, ?, ?)",
			existingPayment.TransactionID, assetAccount.ID, peopleID, debitAmount, nil, "1")
		if err != nil {
			return nil, fmt.Errorf("failed to create asset debit split: %v", err)
		}

		_, err = tx.Exec("INSERT INTO splits (transaction_id, account_id, people_id, debit, credit, status) VALUES (?, ?, ?, ?, ?, ?)",
			existingPayment.TransactionID, arAccount.ID, peopleID, nil, creditAmount, "1")
		if err != nil {
			return nil, fmt.Errorf("failed to create A/R credit split: %v", err)
		}
	} else {
		// Refund/reversal: Credit Asset, Debit A/R
		_, err = tx.Exec("INSERT INTO splits (transaction_id, account_id, people_id, debit, credit, status) VALUES (?, ?, ?, ?, ?, ?)",
			existingPayment.TransactionID, assetAccount.ID, peopleID, nil, creditAmount, "1")
		if err != nil {
			return nil, fmt.Errorf("failed to create asset credit split: %v", err)
		}

		_, err = tx.Exec("INSERT INTO splits (transaction_id, account_id, people_id, debit, credit, status) VALUES (?, ?, ?, ?, ?, ?)",
			existingPayment.TransactionID, arAccount.ID, peopleID, debitAmount, nil, "1")
		if err != nil {
			return nil, fmt.Errorf("failed to create A/R debit split: %v", err)
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	// Fetch updated records
	updatedTransaction, err := s.transactionRepo.GetByID(existingPayment.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transaction: %v", err)
	}

	updatedPaymentRecord, err := s.paymentRepo.GetByID(paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch invoice payment: %v", err)
	}

	updatedSplits, err := s.splitRepo.GetByTransactionID(existingPayment.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch splits: %v", err)
	}

	return &InvoicePaymentResponse{
		Payment:     updatedPaymentRecord,
		Splits:      updatedSplits,
		Transaction: updatedTransaction,
		Invoice:     invoice,
		ARAccount:   &arAccount,
	}, nil
}

// GetInvoicePaymentWithDetails returns invoice payment with all related details (splits, transaction, invoice)
func (s *InvoicePaymentService) GetInvoicePaymentWithDetails(paymentID int) (*InvoicePaymentResponse, error) {
	// Get payment
	payment, err := s.paymentRepo.GetByID(paymentID)
	if err != nil {
		return nil, fmt.Errorf("invoice payment not found: %v", err)
	}

	// Get invoice
	invoice, err := s.invoiceRepo.GetByID(payment.InvoiceID)
	if err != nil {
		return nil, fmt.Errorf("invoice not found: %v", err)
	}

	// Get AR account
	var arAccount *accounts.Account
	if invoice.ARAccountID != nil {
		arAcc, _, _, err := s.accountRepo.GetByID(*invoice.ARAccountID)
		if err == nil {
			arAccount = &arAcc
		}
	}

	// Get transaction
	transaction, err := s.transactionRepo.GetByID(payment.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("transaction not found: %v", err)
	}

	// Get splits (both active and inactive)
	splits, err := s.splitRepo.GetByTransactionID(payment.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch splits: %v", err)
	}

	return &InvoicePaymentResponse{
		Payment:     payment,
		Splits:      splits,
		Transaction: transaction,
		Invoice:     invoice,
		ARAccount:   arAccount,
	}, nil
}

