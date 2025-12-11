package invoice_payments

import (
	"github.com/mysecodgit/go_accounting/src/accounts"
	"github.com/mysecodgit/go_accounting/src/invoices"
	"github.com/mysecodgit/go_accounting/src/splits"
	"github.com/mysecodgit/go_accounting/src/transactions"
)

type CreateInvoicePaymentRequest struct {
	Date      string  `json:"date"`
	InvoiceID int     `json:"invoice_id"`
	AccountID int     `json:"account_id"` // Asset account (cash/bank)
	Amount    float64 `json:"amount"`
	Status    *int    `json:"status"`
	BuildingID int    `json:"building_id"`
}

type InvoicePaymentResponse struct {
	Payment     InvoicePayment                `json:"payment"`
	Splits      []splits.Split                `json:"splits"`
	Transaction transactions.Transaction      `json:"transaction"`
	Invoice     invoices.Invoice              `json:"invoice"`
	ARAccount   *accounts.Account             `json:"ar_account,omitempty"`
}

