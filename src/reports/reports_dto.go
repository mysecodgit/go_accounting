package reports

// Balance Sheet DTOs
type BalanceSheetRequest struct {
	BuildingID int    `json:"building_id"`
	AsOfDate   string `json:"as_of_date"` // Date to calculate balance sheet as of
}

type AccountBalance struct {
	AccountID     int     `json:"account_id"`
	AccountNumber string  `json:"account_number"`
	AccountName   string  `json:"account_name"`
	AccountType   string  `json:"account_type"`
	Balance       float64 `json:"balance"`
}

type BalanceSheetSection struct {
	SectionName string           `json:"section_name"`
	Accounts    []AccountBalance `json:"accounts"`
	Total       float64          `json:"total"`
}

type BalanceSheetResponse struct {
	BuildingID                int                 `json:"building_id"`
	AsOfDate                  string              `json:"as_of_date"`
	Assets                    BalanceSheetSection `json:"assets"`
	Liabilities               BalanceSheetSection `json:"liabilities"`
	Equity                    BalanceSheetSection `json:"equity"`
	TotalAssets               float64             `json:"total_assets"`
	TotalLiabilitiesAndEquity float64             `json:"total_liabilities_and_equity"`
	IsBalanced                bool                `json:"is_balanced"`
}

// Trial Balance DTOs
type TrialBalanceRequest struct {
	BuildingID int    `json:"building_id"`
	AsOfDate   string `json:"as_of_date"` // Date to calculate trial balance as of
}

type TrialBalanceAccount struct {
	AccountID     int     `json:"account_id"`
	AccountNumber int     `json:"account_number"`
	AccountName   string  `json:"account_name"`
	AccountType   string  `json:"account_type"`
	DebitBalance  float64 `json:"debit_balance"`          // Debit balance (0 if credit account)
	CreditBalance float64 `json:"credit_balance"`         // Credit balance (0 if debit account)
	IsTotalRow    bool    `json:"is_total_row,omitempty"` // Flag to indicate this is a total row
}

type TrialBalanceResponse struct {
	BuildingID  int                   `json:"building_id"`
	AsOfDate    string                `json:"as_of_date"`
	Accounts    []TrialBalanceAccount `json:"accounts"`
	TotalDebit  float64               `json:"total_debit"`
	TotalCredit float64               `json:"total_credit"`
	IsBalanced  bool                  `json:"is_balanced"`
}

// Transaction Details by Account DTOs
type TransactionDetailsByAccountRequest struct {
	BuildingID int    `json:"building_id"`
	AccountID  *int   `json:"account_id"` // Optional: filter by specific account
	UnitID     *int   `json:"unit_id"`    // Optional: filter by specific unit
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date"`
}

type TransactionDetailSplit struct {
	SplitID           int      `json:"split_id"`
	TransactionID     int      `json:"transaction_id"`
	TransactionNumber string   `json:"transaction_number"`
	TransactionDate   string   `json:"transaction_date"`
	TransactionType   string   `json:"transaction_type"`
	TransactionMemo   string   `json:"transaction_memo"`
	PeopleID          *int     `json:"people_id"`
	PeopleName        *string  `json:"people_name,omitempty"`
	Description       *string  `json:"description,omitempty"`
	Debit             *float64 `json:"debit"`
	Credit            *float64 `json:"credit"`
	Balance           float64  `json:"balance"` // Running balance for this account
}

type AccountTransactionDetails struct {
	AccountID     int                      `json:"account_id"`
	AccountNumber int                      `json:"account_number"`
	AccountName   string                   `json:"account_name"`
	AccountType   string                   `json:"account_type"`
	Splits        []TransactionDetailSplit `json:"splits"`
	TotalDebit    float64                  `json:"total_debit"`
	TotalCredit   float64                  `json:"total_credit"`
	TotalBalance  float64                  `json:"total_balance"`          // Final balance for the account
	IsTotalRow    bool                     `json:"is_total_row,omitempty"` // Flag for total row
}

type TransactionDetailsByAccountResponse struct {
	BuildingID       int                         `json:"building_id"`
	StartDate        string                      `json:"start_date"`
	EndDate          string                      `json:"end_date"`
	Accounts         []AccountTransactionDetails `json:"accounts"`
	GrandTotalDebit  float64                     `json:"grand_total_debit"`
	GrandTotalCredit float64                     `json:"grand_total_credit"`
}

// Customer Balance Summary DTOs
type CustomerBalanceSummaryRequest struct {
	BuildingID int    `json:"building_id"`
	AsOfDate   string `json:"as_of_date"` // Date to calculate balances as of
}

type CustomerBalance struct {
	PeopleID   int     `json:"people_id"`
	PeopleName string  `json:"people_name"`
	Balance    float64 `json:"balance"` // Total balance from all splits
}

type CustomerBalanceSummaryResponse struct {
	BuildingID    int                `json:"building_id"`
	AsOfDate      string             `json:"as_of_date"`
	Customers     []CustomerBalance  `json:"customers"`
	TotalBalance  float64            `json:"total_balance"`
}

// Customer Balance Details DTOs
type CustomerBalanceDetailsRequest struct {
	BuildingID int    `json:"building_id"`
	AsOfDate   string `json:"as_of_date"` // Date to calculate balances as of
	PeopleID   *int   `json:"people_id"`  // Optional: filter by specific customer
}

type CustomerBalanceDetailSplit struct {
	SplitID           int      `json:"split_id"`
	TransactionID     int      `json:"transaction_id"`
	TransactionNumber string   `json:"transaction_number"`
	TransactionDate   string   `json:"transaction_date"`
	TransactionType   string   `json:"transaction_type"`
	TransactionMemo   string   `json:"transaction_memo"`
	AccountID         int      `json:"account_id"`
	AccountName       string   `json:"account_name"`
	AccountNumber     int      `json:"account_number"`
	Debit             *float64 `json:"debit"`
	Credit            *float64 `json:"credit"`
	Balance           float64  `json:"balance"` // Running balance for this customer
}

type CustomerBalanceAccount struct {
	AccountID     int                        `json:"account_id"`
	AccountName   string                     `json:"account_name"`
	AccountNumber int                        `json:"account_number"`
	Splits        []CustomerBalanceDetailSplit `json:"splits"`
	TotalDebit    float64                    `json:"total_debit"`
	TotalCredit   float64                    `json:"total_credit"`
	TotalBalance  float64                    `json:"total_balance"`
	IsTotalRow    bool                       `json:"is_total_row,omitempty"` // Flag for account total row
}

type CustomerBalanceDetails struct {
	PeopleID     int                        `json:"people_id"`
	PeopleName   string                     `json:"people_name"`
	Accounts     []CustomerBalanceAccount    `json:"accounts"`
	TotalDebit   float64                    `json:"total_debit"`
	TotalCredit  float64                    `json:"total_credit"`
	TotalBalance float64                    `json:"total_balance"` // Final balance for the customer
	IsTotalRow   bool                       `json:"is_total_row,omitempty"` // Flag for customer total row
	IsHeader     bool                       `json:"is_header,omitempty"` // Flag for customer header row
}

type CustomerBalanceDetailsResponse struct {
	BuildingID       int                      `json:"building_id"`
	AsOfDate         string                   `json:"as_of_date"`
	Customers        []CustomerBalanceDetails `json:"customers"`
	GrandTotalDebit  float64                 `json:"grand_total_debit"`
	GrandTotalCredit float64                 `json:"grand_total_credit"`
	GrandTotalBalance float64                `json:"grand_total_balance"`
}
