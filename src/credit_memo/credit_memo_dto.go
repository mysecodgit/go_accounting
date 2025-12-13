package credit_memo

import (
	"github.com/mysecodgit/go_accounting/src/splits"
	"github.com/mysecodgit/go_accounting/src/transactions"
)

type CreateCreditMemoRequest struct {
	Reference        string  `json:"reference"`
	Date             string  `json:"date"`
	DepositTo        int     `json:"deposit_to"`
	LiabilityAccount int     `json:"liability_account"`
	PeopleID         int     `json:"people_id"`
	BuildingID       int     `json:"building_id"`
	UnitID           int     `json:"unit_id"`
	Amount           float64 `json:"amount"`
	Description      string  `json:"description"`
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

type CreditMemoPreviewResponse struct {
	CreditMemo  CreateCreditMemoRequest `json:"credit_memo"`
	Splits      []SplitPreview           `json:"splits"`
	TotalDebit  float64                  `json:"total_debit"`
	TotalCredit float64                  `json:"total_credit"`
	IsBalanced  bool                     `json:"is_balanced"`
}

type UpdateCreditMemoRequest struct {
	ID               int     `json:"id"`
	Reference        string  `json:"reference"`
	Date             string  `json:"date"`
	DepositTo        int     `json:"deposit_to"`
	LiabilityAccount int     `json:"liability_account"`
	PeopleID         int     `json:"people_id"`
	BuildingID       int     `json:"building_id"`
	UnitID           int     `json:"unit_id"`
	Amount           float64 `json:"amount"`
	Description      string  `json:"description"`
}

type CreditMemoResponse struct {
	CreditMemo  CreditMemo              `json:"credit_memo"`
	Splits      []splits.Split          `json:"splits"`
	Transaction transactions.Transaction `json:"transaction"`
}

type CreditMemoListItem struct {
	CreditMemo  CreditMemo `json:"credit_memo"`
	UsedCredits float64    `json:"used_credits"`
	Balance     float64    `json:"balance"`
}

