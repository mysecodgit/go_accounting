package invoices

import (
	"github.com/mysecodgit/go_accounting/src/invoice_items"
	"github.com/mysecodgit/go_accounting/src/splits"
	"github.com/mysecodgit/go_accounting/src/transactions"
)

type InvoiceItemInput struct {
	ItemID        int      `json:"item_id"`
	Qty           *float64 `json:"qty"`
	Rate          *string  `json:"rate"`
	Total         *float64 `json:"total"` // Use manually edited total if provided
	PreviousValue *float64 `json:"previous_value"`
	CurrentValue  *float64 `json:"current_value"`
}

type CreateInvoiceRequest struct {
	InvoiceNo   string             `json:"invoice_no"`
	SalesDate   string             `json:"sales_date"`
	DueDate     string             `json:"due_date"`
	UnitID      *int               `json:"unit_id"`
	PeopleID    *int               `json:"people_id"`
	ARAccountID *int               `json:"ar_account_id"`
	Amount      float64            `json:"amount"`
	Description string             `json:"description"`
	Status      *int               `json:"status"` // Use pointer to distinguish between not provided (nil) and explicitly set to 0
	BuildingID  int                `json:"building_id"`
	Items       []InvoiceItemInput `json:"items"`
}

type SplitPreview struct {
	AccountID int      `json:"account_id"`
	AccountName string `json:"account_name"`
	PeopleID   *int    `json:"people_id"`
	UnitID     *int    `json:"unit_id"`
	Debit      *float64 `json:"debit"`
	Credit     *float64 `json:"credit"`
	Status     string  `json:"status"`
}

type InvoicePreviewResponse struct {
	Invoice      CreateInvoiceRequest `json:"invoice"`
	Splits       []SplitPreview       `json:"splits"`
	TotalDebit   float64              `json:"total_debit"`
	TotalCredit  float64              `json:"total_credit"`
	IsBalanced   bool                 `json:"is_balanced"`
}

type UpdateInvoiceRequest struct {
	ID           int                `json:"id"`
	InvoiceNo    string             `json:"invoice_no"`
	SalesDate    string             `json:"sales_date"`
	DueDate      string             `json:"due_date"`
	UnitID       *int               `json:"unit_id"`
	PeopleID     *int               `json:"people_id"`
	ARAccountID  *int               `json:"ar_account_id"`
	Amount       float64            `json:"amount"`
	Description  string             `json:"description"`
	Status       *int               `json:"status"` // Use pointer to distinguish between not provided (nil) and explicitly set to 0
	BuildingID   int                `json:"building_id"`
	Items        []InvoiceItemInput `json:"items"`
}

type InvoiceResponse struct {
	Invoice     Invoice                    `json:"invoice"`
	Items       []invoice_items.InvoiceItem `json:"items"`
	Splits      []splits.Split             `json:"splits"`
	Transaction transactions.Transaction  `json:"transaction"`
}

type InvoiceListItem struct {
	ID                 int     `json:"id"`
	InvoiceNo          string  `json:"invoice_no"`
	TransactionID      int     `json:"transaction_id"`
	SalesDate          string  `json:"sales_date"`
	DueDate            string  `json:"due_date"`
	ARAccountID        *int    `json:"ar_account_id"`
	UnitID             *int    `json:"unit_id"`
	PeopleID           *int    `json:"people_id"`
	UserID             int     `json:"user_id"`
	Amount             float64 `json:"amount"`
	Description        string  `json:"description"`
	CancelReason       *string `json:"cancel_reason"`
	Status             int     `json:"status"`
	BuildingID         int     `json:"building_id"`
	CreatedAt          string  `json:"created_at"`
	UpdatedAt          string  `json:"updated_at"`
	PaidAmount         float64 `json:"paid_amount"`
	AppliedCreditsTotal float64 `json:"applied_credits_total"`
}

