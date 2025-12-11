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
	SectionName string          `json:"section_name"`
	Accounts    []AccountBalance `json:"accounts"`
	Total       float64         `json:"total"`
}

type BalanceSheetResponse struct {
	BuildingID int                 `json:"building_id"`
	AsOfDate   string              `json:"as_of_date"`
	Assets     BalanceSheetSection `json:"assets"`
	Liabilities BalanceSheetSection `json:"liabilities"`
	Equity     BalanceSheetSection `json:"equity"`
	TotalAssets float64            `json:"total_assets"`
	TotalLiabilitiesAndEquity float64 `json:"total_liabilities_and_equity"`
	IsBalanced bool               `json:"is_balanced"`
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
	PeopleID       int     `json:"people_id"`
	PeopleName     string  `json:"people_name"`
	PeopleType     string  `json:"people_type"`
	TotalInvoices  float64 `json:"total_invoices"`
	TotalPayments  float64 `json:"total_payments"`
	Outstanding    float64 `json:"outstanding"`
	InvoiceCount   int     `json:"invoice_count"`
	PaymentCount   int     `json:"payment_count"`
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
	Summary []CustomerVendorSummary        `json:"summary"`
	Details map[int][]CustomerVendorDetail `json:"details"` // Key: people_id
	TotalOutstanding float64               `json:"total_outstanding"`
}

// Transaction Details by Account DTOs
type TransactionDetailsRequest struct {
	BuildingID int      `json:"building_id"`
	AccountIDs []int    `json:"account_ids"` // Optional: filter by specific accounts
	StartDate  string   `json:"start_date"`
	EndDate    string   `json:"end_date"`
	TransactionType *string `json:"transaction_type"` // Optional: filter by transaction type
}

type TransactionDetail struct {
	TransactionID   int     `json:"transaction_id"`
	TransactionDate string  `json:"transaction_date"`
	TransactionType string  `json:"transaction_type"`
	AccountID       int     `json:"account_id"`
	AccountName     string  `json:"account_name"`
	PeopleID        *int    `json:"people_id"`
	PeopleName      *string `json:"people_name,omitempty"`
	Debit           *float64 `json:"debit"`
	Credit          *float64 `json:"credit"`
	Balance         float64 `json:"balance"` // Running balance
	Description     string  `json:"description"`
}

type TransactionDetailsResponse struct {
	AccountID       int               `json:"account_id,omitempty"`
	AccountName     string            `json:"account_name,omitempty"`
	StartDate       string            `json:"start_date"`
	EndDate         string            `json:"end_date"`
	OpeningBalance  float64           `json:"opening_balance"`
	ClosingBalance  float64           `json:"closing_balance"`
	TotalDebit      float64           `json:"total_debit"`
	TotalCredit     float64           `json:"total_credit"`
	Transactions    []TransactionDetail `json:"transactions"`
	ByAccount       map[int][]TransactionDetail `json:"by_account,omitempty"` // When multiple accounts
}

