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

	// Create invoice payment - always use status 1 (active)
	paymentStatus := 1
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

