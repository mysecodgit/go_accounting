package sales_receipt

import (
	"github.com/mysecodgit/go_accounting/src/receipt_items"
	"github.com/mysecodgit/go_accounting/src/splits"
	"github.com/mysecodgit/go_accounting/src/transactions"
)

type ReceiptItemInput struct {
	ItemID        int      `json:"item_id"`
	Qty           *float64 `json:"qty"`
	Rate          *string  `json:"rate"`
	PreviousValue *float64 `json:"previous_value"`
	CurrentValue  *float64 `json:"current_value"`
}

type CreateSalesReceiptRequest struct {
	ReceiptNo   string             `json:"receipt_no"`
	ReceiptDate string             `json:"receipt_date"`
	UnitID      *int               `json:"unit_id"`
	PeopleID    *int               `json:"people_id"`
	AccountID   int                `json:"account_id"` // Asset account (cash/bank)
	Amount      float64            `json:"amount"`
	Description string             `json:"description"`
	Status      *int               `json:"status"`
	BuildingID  int                `json:"building_id"`
	Items       []ReceiptItemInput `json:"items"`
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

type SalesReceiptPreviewResponse struct {
	Receipt     CreateSalesReceiptRequest `json:"receipt"`
	Splits      []SplitPreview            `json:"splits"`
	TotalDebit  float64                   `json:"total_debit"`
	TotalCredit float64                   `json:"total_credit"`
	IsBalanced  bool                      `json:"is_balanced"`
}

type UpdateSalesReceiptRequest struct {
	ID          int                `json:"id"`
	ReceiptNo   string             `json:"receipt_no"`
	ReceiptDate string             `json:"receipt_date"`
	UnitID      *int               `json:"unit_id"`
	PeopleID    *int               `json:"people_id"`
	AccountID   int                `json:"account_id"`
	Amount      float64            `json:"amount"`
	Description string             `json:"description"`
	Status      *int               `json:"status"`
	BuildingID  int                `json:"building_id"`
	Items       []ReceiptItemInput `json:"items"`
}

type SalesReceiptResponse struct {
	Receipt     SalesReceipt                `json:"receipt"`
	Items       []receipt_items.ReceiptItem `json:"items"`
	Splits      []splits.Split              `json:"splits"`
	Transaction transactions.Transaction    `json:"transaction"`
}
