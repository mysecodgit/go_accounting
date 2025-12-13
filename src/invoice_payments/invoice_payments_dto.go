package invoice_payments

import (
	"github.com/mysecodgit/go_accounting/src/accounts"
	"github.com/mysecodgit/go_accounting/src/invoices"
	"github.com/mysecodgit/go_accounting/src/splits"
	"github.com/mysecodgit/go_accounting/src/transactions"
)

type CreateInvoicePaymentRequest struct {
	Reference string  `json:"reference"`
	Date      string  `json:"date"`
	InvoiceID int     `json:"invoice_id"`
	AccountID int     `json:"account_id"` // Asset account (cash/bank)
	Amount    float64 `json:"amount"`
	Status    *int    `json:"status"`
	BuildingID int    `json:"building_id"`
}

type UpdateInvoicePaymentRequest struct {
	Reference string  `json:"reference"`
	Date      string  `json:"date"`
	AccountID int     `json:"account_id"` // Asset account (cash/bank)
	Amount    float64 `json:"amount"`
	Status    *int    `json:"status"`
	BuildingID int    `json:"building_id"`
}

type SplitPreview struct {
	AccountID   int      `json:"account_id"`
	AccountName string   `json:"account_name"`
	PeopleID    *int     `json:"people_id"`
	Debit       *float64 `json:"debit"`
	Credit      *float64 `json:"credit"`
	Status      string   `json:"status"`
}

type InvoicePaymentPreviewResponse struct {
	Splits      []SplitPreview `json:"splits"`
	TotalDebit  float64        `json:"total_debit"`
	TotalCredit float64        `json:"total_credit"`
	IsBalanced bool           `json:"is_balanced"`
}

type InvoicePaymentResponse struct {
	Payment     InvoicePayment                `json:"payment"`
	Splits      []splits.Split                `json:"splits"`
	Transaction transactions.Transaction      `json:"transaction"`
	Invoice     invoices.Invoice              `json:"invoice"`
	ARAccount   *accounts.Account             `json:"ar_account,omitempty"`
}

