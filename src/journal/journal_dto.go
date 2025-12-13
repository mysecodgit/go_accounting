package journal

import (
	"github.com/mysecodgit/go_accounting/src/journal_lines"
	"github.com/mysecodgit/go_accounting/src/splits"
	"github.com/mysecodgit/go_accounting/src/transactions"
)

type JournalLineInput struct {
	AccountID   int      `json:"account_id"`
	UnitID      *int     `json:"unit_id"`
	PeopleID    *int     `json:"people_id"`
	Description *string  `json:"description"`
	Debit       *float64 `json:"debit"`
	Credit      *float64 `json:"credit"`
}

type CreateJournalRequest struct {
	Reference   string            `json:"reference"`
	JournalDate string            `json:"journal_date"`
	BuildingID    int             `json:"building_id"`
	Memo         *string          `json:"memo"`
	TotalAmount  float64          `json:"total_amount"`
	Lines        []JournalLineInput `json:"lines"`
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

type JournalPreviewResponse struct {
	Journal     CreateJournalRequest `json:"journal"`
	Splits      []SplitPreview       `json:"splits"`
	TotalDebit  float64              `json:"total_debit"`
	TotalCredit float64              `json:"total_credit"`
	IsBalanced  bool                 `json:"is_balanced"`
}

type UpdateJournalRequest struct {
	ID          int                `json:"id"`
	Reference   string             `json:"reference"`
	JournalDate string             `json:"journal_date"`
	BuildingID  int                `json:"building_id"`
	Memo        *string            `json:"memo"`
	TotalAmount float64            `json:"total_amount"`
	Lines       []JournalLineInput `json:"lines"`
}

type JournalResponse struct {
	Journal Journal                      `json:"journal"`
	Lines   []journal_lines.JournalLine `json:"lines"`
	Splits  []splits.Split               `json:"splits"`
	Transaction transactions.Transaction `json:"transaction"`
}

