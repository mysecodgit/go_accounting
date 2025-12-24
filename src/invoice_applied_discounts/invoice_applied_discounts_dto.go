package invoice_applied_discounts

import (
	"github.com/mysecodgit/go_accounting/src/splits"
	"github.com/mysecodgit/go_accounting/src/transactions"
)

type CreateInvoiceAppliedDiscountRequest struct {
	InvoiceID     int     `json:"invoice_id"`
	TransactionID int     `json:"transaction_id"`
	ARAccount     int     `json:"ar_account"`
	IncomeAccount int     `json:"income_account"`
	Amount        float64 `json:"amount"`
	Description   string  `json:"description"`
	Date          string  `json:"date"`
	Reference     string  `json:"reference"`
}

type InvoiceAppliedDiscountResponse struct {
	InvoiceAppliedDiscount InvoiceAppliedDiscount    `json:"invoice_applied_discount"`
	Splits                  []splits.Split           `json:"splits"`
	Transaction             transactions.Transaction `json:"transaction"`
}

type SplitPreview struct {
	AccountID   int      `json:"account_id"`
	AccountName string   `json:"account_name"`
	PeopleID    *int     `json:"people_id"`
	UnitID      *int     `json:"unit_id"`
	Debit       *float64 `json:"debit"`
	Credit      *float64 `json:"credit"`
	Status      string   `json:"status"`
}

type InvoiceAppliedDiscountPreviewResponse struct {
	AppliedDiscount CreateInvoiceAppliedDiscountRequest `json:"applied_discount"`
	Splits          []SplitPreview                      `json:"splits"`
	TotalDebit      float64                             `json:"total_debit"`
	TotalCredit     float64                             `json:"total_credit"`
	IsBalanced      bool                                `json:"is_balanced"`
}

