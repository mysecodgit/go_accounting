package checks

import (
	"github.com/mysecodgit/go_accounting/src/expense_lines"
	"github.com/mysecodgit/go_accounting/src/splits"
	"github.com/mysecodgit/go_accounting/src/transactions"
)

type ExpenseLineInput struct {
	AccountID   int     `json:"account_id"`
	UnitID      *int    `json:"unit_id"`
	PeopleID    *int    `json:"people_id"`
	Description *string `json:"description"`
	Amount      float64 `json:"amount"`
}

type CreateCheckRequest struct {
	CheckDate        string             `json:"check_date"`
	ReferenceNumber  *string            `json:"reference_number"`
	PaymentAccountID int                `json:"payment_account_id"`
	BuildingID       int                `json:"building_id"`
	Memo             *string            `json:"memo"`
	TotalAmount      float64            `json:"total_amount"`
	ExpenseLines     []ExpenseLineInput `json:"expense_lines"`
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

type CheckPreviewResponse struct {
	Check       CreateCheckRequest `json:"check"`
	Splits      []SplitPreview     `json:"splits"`
	TotalDebit  float64            `json:"total_debit"`
	TotalCredit float64            `json:"total_credit"`
	IsBalanced  bool               `json:"is_balanced"`
}

type UpdateCheckRequest struct {
	ID               int                `json:"id"`
	CheckDate        string             `json:"check_date"`
	ReferenceNumber  *string            `json:"reference_number"`
	PaymentAccountID int                `json:"payment_account_id"`
	BuildingID       int                `json:"building_id"`
	Memo             *string            `json:"memo"`
	TotalAmount      float64            `json:"total_amount"`
	ExpenseLines     []ExpenseLineInput `json:"expense_lines"`
}

type CheckResponse struct {
	Check        Check                       `json:"check"`
	ExpenseLines []expense_lines.ExpenseLine `json:"expense_lines"`
	Splits       []splits.Split              `json:"splits"`
	Transaction  transactions.Transaction    `json:"transaction"`
}
