package invoice_applied_credits

import (
	"github.com/mysecodgit/go_accounting/src/splits"
	"github.com/mysecodgit/go_accounting/src/transactions"
)

type CreateInvoiceAppliedCreditRequest struct {
	InvoiceID    int     `json:"invoice_id"`
	CreditMemoID int     `json:"credit_memo_id"`
	Amount       float64 `json:"amount"`
	Description  string  `json:"description"`
	Date         string  `json:"date"`
}

type InvoiceAppliedCreditResponse struct {
	InvoiceAppliedCredit InvoiceAppliedCredit      `json:"invoice_applied_credit"`
	Splits                []splits.Split            `json:"splits"`
	Transaction           transactions.Transaction   `json:"transaction"`
}

type AvailableCreditMemo struct {
	ID              int     `json:"id"`
	Date            string  `json:"date"`
	Amount          float64 `json:"amount"`
	AppliedAmount   float64 `json:"applied_amount"`   // Amount already applied to other invoices
	AvailableAmount float64 `json:"available_amount"` // Amount available to apply
	Description     string  `json:"description"`
}

type AvailableCreditsResponse struct {
	InvoiceID int                  `json:"invoice_id"`
	PeopleID  int                  `json:"people_id"`
	Credits   []AvailableCreditMemo `json:"credits"`
}

type SplitPreview struct {
	AccountID   int      `json:"account_id"`
	AccountName string   `json:"account_name"`
	PeopleID    *int     `json:"people_id"`
	Debit       *float64 `json:"debit"`
	Credit      *float64 `json:"credit"`
	Status      string   `json:"status"`
}

type InvoiceAppliedCreditPreviewResponse struct {
	AppliedCredit CreateInvoiceAppliedCreditRequest `json:"applied_credit"`
	Splits        []SplitPreview                    `json:"splits"`
	TotalDebit    float64                           `json:"total_debit"`
	TotalCredit   float64                           `json:"total_credit"`
	IsBalanced    bool                              `json:"is_balanced"`
}

