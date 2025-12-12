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

// Customer/Vendor Report DTOs
type CustomerVendorReportRequest struct {
	BuildingID int    `json:"building_id"`
	PeopleID   *int   `json:"people_id"` // Optional: filter by specific customer/vendor
	TypeID     *int   `json:"type_id"`   // Optional: filter by people type (customer/vendor)
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date"`
}

type CustomerVendorSummary struct {
	PeopleID      int     `json:"people_id"`
	PeopleName    string  `json:"people_name"`
	PeopleType    string  `json:"people_type"`
	TotalInvoices float64 `json:"total_invoices"`
	TotalPayments float64 `json:"total_payments"`
	Outstanding   float64 `json:"outstanding"`
	InvoiceCount  int     `json:"invoice_count"`
	PaymentCount  int     `json:"payment_count"`
}

type CustomerVendorDetail struct {
	TransactionID   int     `json:"transaction_id"`
	TransactionDate string  `json:"transaction_date"`
	TransactionType string  `json:"transaction_type"`
	InvoiceNo       *int    `json:"invoice_no,omitempty"`
	Description     string  `json:"description"`
	Amount          float64 `json:"amount"`
	Balance         float64 `json:"balance"`
}

type CustomerVendorReportResponse struct {
	Summary          []CustomerVendorSummary        `json:"summary"`
	Details          map[int][]CustomerVendorDetail `json:"details"` // Key: people_id
	TotalOutstanding float64                        `json:"total_outstanding"`
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
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date"`
}

type TransactionDetailSplit struct {
	SplitID         int      `json:"split_id"`
	TransactionID   int      `json:"transaction_id"`
	TransactionDate string   `json:"transaction_date"`
	TransactionType string   `json:"transaction_type"`
	TransactionMemo string   `json:"transaction_memo"`
	PeopleID        *int     `json:"people_id"`
	PeopleName      *string  `json:"people_name,omitempty"`
	Description     *string  `json:"description,omitempty"`
	Debit           *float64 `json:"debit"`
	Credit          *float64 `json:"credit"`
	Balance         float64  `json:"balance"` // Running balance for this account
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
