package invoice_applied_credits

import (
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
	accountRepo       accounts.AccountRepository
}

func NewInvoiceAppliedCreditService(
	appliedCreditRepo InvoiceAppliedCreditRepository,
	invoiceRepo invoices.InvoiceRepository,
	creditMemoRepo credit_memo.CreditMemoRepository,
	accountRepo accounts.AccountRepository,
) *InvoiceAppliedCreditService {
	return &InvoiceAppliedCreditService{
		appliedCreditRepo: appliedCreditRepo,
		invoiceRepo:       invoiceRepo,
		creditMemoRepo:    creditMemoRepo,
		accountRepo:       accountRepo,
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
	invoicePeopleIDPtr := invoice.PeopleID
	splits = append(splits, SplitPreview{
		AccountID:   creditMemo.LiabilityAccount,
		AccountName: liabilityAccount.AccountName,
		PeopleID:    invoicePeopleIDPtr, // people_id is set for both splits
		Debit:       &debitAmount,
		Credit:      nil,
		Status:      "1",
	})

	// Credit: A/R account (reduces receivable)
	creditAmount := req.Amount
	splits = append(splits, SplitPreview{
		AccountID:   *invoice.ARAccountID,
		AccountName: arAccount.AccountName,
		PeopleID:    invoicePeopleIDPtr, // people_id is set for both splits
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

	// Create invoice applied credit record (no transaction or splits needed)
	appliedCreditStatus := "1"
	appliedCredit := InvoiceAppliedCredit{
		TransactionID: nil, // No transaction needed
		InvoiceID:     req.InvoiceID,
		CreditMemoID:  req.CreditMemoID,
		Amount:        req.Amount,
		Description:   req.Description,
		Date:          req.Date,
		Status:        appliedCreditStatus,
	}

	createdAppliedCredit, err := s.appliedCreditRepo.Create(appliedCredit)
	if err != nil {
		return nil, fmt.Errorf("failed to create invoice applied credit: %v", err)
	}

	return &InvoiceAppliedCreditResponse{
		InvoiceAppliedCredit: createdAppliedCredit,
		Splits:               []splits.Split{}, // No splits
		Transaction:          transactions.Transaction{}, // No transaction
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

	// Soft delete the applied credit (no transaction or splits to delete)
	appliedCredit.Status = "0"
	_, err = s.appliedCreditRepo.Update(appliedCredit)
	if err != nil {
		return fmt.Errorf("failed to delete applied credit: %v", err)
	}

	return nil
}
