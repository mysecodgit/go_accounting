package reports

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/mysecodgit/go_accounting/src/accounts"
	"github.com/mysecodgit/go_accounting/src/invoice_payments"
	"github.com/mysecodgit/go_accounting/src/invoices"
	"github.com/mysecodgit/go_accounting/src/people"
	"github.com/mysecodgit/go_accounting/src/splits"
	"github.com/mysecodgit/go_accounting/src/transactions"
)

type ReportsService struct {
	accountRepo     accounts.AccountRepository
	splitRepo       splits.SplitRepository
	transactionRepo transactions.TransactionRepository
	invoiceRepo     invoices.InvoiceRepository
	paymentRepo     invoice_payments.InvoicePaymentRepository
	peopleRepo      people.PersonRepository
	db              *sql.DB
}

func NewReportsService(
	accountRepo accounts.AccountRepository,
	splitRepo splits.SplitRepository,
	transactionRepo transactions.TransactionRepository,
	invoiceRepo invoices.InvoiceRepository,
	paymentRepo invoice_payments.InvoicePaymentRepository,
	peopleRepo people.PersonRepository,
	db *sql.DB,
) *ReportsService {
	return &ReportsService{
		accountRepo:     accountRepo,
		splitRepo:       splitRepo,
		transactionRepo: transactionRepo,
		invoiceRepo:     invoiceRepo,
		paymentRepo:     paymentRepo,
		peopleRepo:      peopleRepo,
		db:              db,
	}
}

// GetBalanceSheet generates a balance sheet report
func (s *ReportsService) GetBalanceSheet(req BalanceSheetRequest) (*BalanceSheetResponse, error) {
	// Parse as of date
	asOfDate := req.AsOfDate
	if asOfDate == "" {
		asOfDate = time.Now().Format("2006-01-02")
	}

	// Get all accounts for the building
	accountsList, accountTypes, _, err := s.accountRepo.GetByBuildingID(req.BuildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %v", err)
	}

	// Calculate balances for each account up to asOfDate
	accountBalances := make(map[int]float64)
	for _, account := range accountsList {
		balance, err := s.calculateAccountBalance(account.ID, asOfDate)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate balance for account %d: %v", account.ID, err)
		}
		accountBalances[account.ID] = balance
	}

	// Categorize accounts into Assets, Liabilities, and Equity
	assets := []AccountBalance{}
	liabilities := []AccountBalance{}
	equity := []AccountBalance{}

	for i, account := range accountsList {
		accountType := accountTypes[i]
		balance := accountBalances[account.ID]

		accountBalance := AccountBalance{
			AccountID:     account.ID,
			AccountNumber: fmt.Sprintf("%d", account.AccountNumber),
			AccountName:   account.AccountName,
			AccountType:   accountType.TypeName,
			Balance:       balance,
		}

		// Categorize based on account type field (Asset, Liability, Equity, Income, Expense)
		typeLower := strings.ToLower(accountType.Type)
		if typeLower == "asset" {
			assets = append(assets, accountBalance)
		} else if typeLower == "liability" {
			liabilities = append(liabilities, accountBalance)
		} else if typeLower == "equity" {
			equity = append(equity, accountBalance)
		} else if typeLower == "income" || typeLower == "expense" {
			// Income and Expense accounts are typically shown in Income Statement, not Balance Sheet
			// But for now, we'll include them in Equity section (Income increases equity, Expense decreases)
			// You might want to create a separate Income Statement report later
			equity = append(equity, accountBalance)
		}
	}

	// Calculate totals
	totalAssets := 0.0
	for _, asset := range assets {
		totalAssets += asset.Balance
	}

	totalLiabilities := 0.0
	for _, liability := range liabilities {
		totalLiabilities += liability.Balance
	}

	totalEquity := 0.0
	for _, eq := range equity {
		totalEquity += eq.Balance
	}

	totalLiabilitiesAndEquity := totalLiabilities + totalEquity
	isBalanced := totalAssets == totalLiabilitiesAndEquity

	return &BalanceSheetResponse{
		BuildingID:                req.BuildingID,
		AsOfDate:                  asOfDate,
		Assets:                    BalanceSheetSection{SectionName: "Assets", Accounts: assets, Total: totalAssets},
		Liabilities:               BalanceSheetSection{SectionName: "Liabilities", Accounts: liabilities, Total: totalLiabilities},
		Equity:                    BalanceSheetSection{SectionName: "Equity", Accounts: equity, Total: totalEquity},
		TotalAssets:               totalAssets,
		TotalLiabilitiesAndEquity: totalLiabilitiesAndEquity,
		IsBalanced:                isBalanced,
	}, nil
}

// calculateAccountBalance calculates the balance of an account up to a specific date
func (s *ReportsService) calculateAccountBalance(accountID int, asOfDate string) (float64, error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN s.debit IS NOT NULL THEN s.debit ELSE 0 END), 0) as total_debit,
			COALESCE(SUM(CASE WHEN s.credit IS NOT NULL THEN s.credit ELSE 0 END), 0) as total_credit
		FROM splits s
		INNER JOIN transactions t ON s.transaction_id = t.id
		WHERE s.account_id = ? 
			AND s.status = '1'
			AND t.status = '1'
			AND DATE(t.transaction_date) <= ?
	`

	var totalDebit, totalCredit sql.NullFloat64
	err := s.db.QueryRow(query, accountID, asOfDate).Scan(&totalDebit, &totalCredit)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	// Convert NullFloat64 to float64
	debitAmount := 0.0
	if totalDebit.Valid {
		debitAmount = totalDebit.Float64
	}
	creditAmount := 0.0
	if totalCredit.Valid {
		creditAmount = totalCredit.Float64
	}

	// Get account type to determine if it's a debit or credit account
	_, accountType, _, err := s.accountRepo.GetByID(accountID)
	if err != nil {
		return 0, err
	}

	// Use typeStatus field to determine balance calculation
	// typeStatus can be "debit" or "credit"
	typeStatusLower := strings.ToLower(accountType.TypeStatus)

	// Debit accounts: Debit - Credit (Assets, Expenses)
	// Credit accounts: Credit - Debit (Liabilities, Equity, Income)
	if typeStatusLower == "debit" {
		return debitAmount - creditAmount, nil
	} else {
		return creditAmount - debitAmount, nil
	}
}

// calculateAccountBalanceBeforeDate calculates the balance of an account before a specific date (exclusive)
func (s *ReportsService) calculateAccountBalanceBeforeDate(accountID int, beforeDate string) (float64, error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN s.debit IS NOT NULL THEN s.debit ELSE 0 END), 0) as total_debit,
			COALESCE(SUM(CASE WHEN s.credit IS NOT NULL THEN s.credit ELSE 0 END), 0) as total_credit
		FROM splits s
		INNER JOIN transactions t ON s.transaction_id = t.id
		WHERE s.account_id = ? 
			AND s.status = '1'
			AND t.status = '1'
			AND DATE(t.transaction_date) < ?
	`

	var totalDebit, totalCredit sql.NullFloat64
	err := s.db.QueryRow(query, accountID, beforeDate).Scan(&totalDebit, &totalCredit)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	// Convert NullFloat64 to float64
	debitAmount := 0.0
	if totalDebit.Valid {
		debitAmount = totalDebit.Float64
	}
	creditAmount := 0.0
	if totalCredit.Valid {
		creditAmount = totalCredit.Float64
	}

	// Get account type to determine if it's a debit or credit account
	_, accountType, _, err := s.accountRepo.GetByID(accountID)
	if err != nil {
		return 0, err
	}

	// Use typeStatus field to determine balance calculation
	// typeStatus can be "debit" or "credit"
	typeStatusLower := strings.ToLower(accountType.TypeStatus)

	// Debit accounts: Debit - Credit (Assets, Expenses)
	// Credit accounts: Credit - Debit (Liabilities, Equity, Income)
	if typeStatusLower == "debit" {
		return debitAmount - creditAmount, nil
	} else {
		return creditAmount - debitAmount, nil
	}
}

// GetCustomerVendorReport generates customer or vendor report
func (s *ReportsService) GetCustomerVendorReport(req CustomerVendorReportRequest) (*CustomerVendorReportResponse, error) {
	// Get people based on filters
	var peopleList []people.Person
	var err error

	if req.PeopleID != nil {
		person, _, _, err := s.peopleRepo.GetByID(*req.PeopleID)
		if err != nil {
			return nil, fmt.Errorf("person not found: %v", err)
		}
		peopleList = []people.Person{person}
	} else {
		peopleList, _, _, err = s.peopleRepo.GetByBuildingID(req.BuildingID)
		if err != nil {
			return nil, fmt.Errorf("failed to get people: %v", err)
		}
	}

	summary := []CustomerVendorSummary{}
	details := make(map[int][]CustomerVendorDetail)
	totalOutstanding := 0.0

	for _, person := range peopleList {
		// Get invoices for this person within date range
		invoices, err := s.getInvoicesForPerson(person.ID, req.StartDate, req.EndDate)
		if err != nil {
			continue
		}

		// Get payments for this person within date range
		payments, err := s.getPaymentsForPerson(person.ID, req.StartDate, req.EndDate)
		if err != nil {
			continue
		}

		// Calculate totals
		totalInvoices := 0.0
		totalPayments := 0.0
		invoiceCount := len(invoices)
		paymentCount := len(payments)

		personDetails := []CustomerVendorDetail{}

		// Process invoices
		for _, invoice := range invoices {
			totalInvoices += invoice.Amount
			personDetails = append(personDetails, CustomerVendorDetail{
				TransactionID:   invoice.TransactionID,
				TransactionDate: invoice.SalesDate,
				TransactionType: "Invoice",
				InvoiceNo:       &invoice.InvoiceNo,
				Description:     invoice.Description,
				Amount:          invoice.Amount,
			})
		}

		// Process payments
		for _, payment := range payments {
			totalPayments += payment.Amount
			personDetails = append(personDetails, CustomerVendorDetail{
				TransactionID:   payment.TransactionID,
				TransactionDate: payment.Date,
				TransactionType: "Payment",
				Description:     fmt.Sprintf("Payment for Invoice"),
				Amount:          payment.Amount,
			})
		}

		outstanding := totalInvoices - totalPayments
		totalOutstanding += outstanding

		// Get people type name
		peopleTypeName := "Customer" // Default
		// You might want to fetch people type details here

		summary = append(summary, CustomerVendorSummary{
			PeopleID:      person.ID,
			PeopleName:    person.Name,
			PeopleType:    peopleTypeName,
			TotalInvoices: totalInvoices,
			TotalPayments: totalPayments,
			Outstanding:   outstanding,
			InvoiceCount:  invoiceCount,
			PaymentCount:  paymentCount,
		})

		details[person.ID] = personDetails
	}

	return &CustomerVendorReportResponse{
		Summary:          summary,
		Details:          details,
		TotalOutstanding: totalOutstanding,
	}, nil
}

// Helper function to get invoices for a person
func (s *ReportsService) getInvoicesForPerson(peopleID int, startDate, endDate string) ([]invoices.Invoice, error) {
	query := `
		SELECT id, invoice_no, transaction_id, sales_date, due_date, ar_account_id, unit_id, people_id, user_id, amount, description, refrence, cancel_reason, status, building_id, createdAt, updatedAt
		FROM invoices
		WHERE people_id = ? 
			AND status = 1
			AND DATE(sales_date) >= ?
			AND DATE(sales_date) <= ?
		ORDER BY sales_date DESC
	`

	rows, err := s.db.Query(query, peopleID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	invoiceList := []invoices.Invoice{}
	for rows.Next() {
		var invoice invoices.Invoice
		err := rows.Scan(&invoice.ID, &invoice.InvoiceNo, &invoice.TransactionID, &invoice.SalesDate, &invoice.DueDate, &invoice.ARAccountID, &invoice.UnitID, &invoice.PeopleID, &invoice.UserID, &invoice.Amount, &invoice.Description, &invoice.Reference, &invoice.CancelReason, &invoice.Status, &invoice.BuildingID, &invoice.CreatedAt, &invoice.UpdatedAt)
		if err != nil {
			continue
		}
		invoiceList = append(invoiceList, invoice)
	}

	return invoiceList, nil
}

// Helper function to get payments for a person
func (s *ReportsService) getPaymentsForPerson(peopleID int, startDate, endDate string) ([]invoice_payments.InvoicePayment, error) {
	// Get payments through invoices
	query := `
		SELECT ip.id, ip.transaction_id, ip.date, ip.invoice_id, ip.user_id, ip.account_id, ip.amount, ip.status, ip.createdAt, ip.updatedAt
		FROM invoice_payments ip
		INNER JOIN invoices i ON ip.invoice_id = i.id
		WHERE i.people_id = ?
			AND ip.status = 1
			AND DATE(ip.date) >= ?
			AND DATE(ip.date) <= ?
		ORDER BY ip.date DESC
	`

	rows, err := s.db.Query(query, peopleID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	paymentList := []invoice_payments.InvoicePayment{}
	for rows.Next() {
		var payment invoice_payments.InvoicePayment
		err := rows.Scan(&payment.ID, &payment.TransactionID, &payment.Date, &payment.InvoiceID, &payment.UserID, &payment.AccountID, &payment.Amount, &payment.Status, &payment.CreatedAt, &payment.UpdatedAt)
		if err != nil {
			continue
		}
		paymentList = append(paymentList, payment)
	}

	return paymentList, nil
}

// GetTransactionDetailsByAccount generates transaction details report filtered by account
func (s *ReportsService) GetTransactionDetailsByAccount(req TransactionDetailsRequest) (*TransactionDetailsResponse, error) {
	// Build query
	query := `
		SELECT 
			s.id, s.transaction_id, s.account_id, s.people_id, s.debit, s.credit, s.status,
			t.transaction_date, t.type as transaction_type, t.memo as description
		FROM splits s
		INNER JOIN transactions t ON s.transaction_id = t.id
		WHERE t.building_id = ?
			AND s.status = '1'
			AND t.status = 1
			AND DATE(t.transaction_date) >= ?
			AND DATE(t.transaction_date) <= ?
	`

	args := []interface{}{req.BuildingID, req.StartDate, req.EndDate}

	// Add account filter if provided
	if len(req.AccountIDs) > 0 {
		placeholders := strings.Repeat("?,", len(req.AccountIDs))
		placeholders = placeholders[:len(placeholders)-1]
		query += fmt.Sprintf(" AND s.account_id IN (%s)", placeholders)
		for _, accountID := range req.AccountIDs {
			args = append(args, accountID)
		}
	}

	// Add transaction type filter if provided
	if req.TransactionType != nil && *req.TransactionType != "" {
		query += " AND t.type = ?"
		args = append(args, *req.TransactionType)
	}

	query += " ORDER BY t.transaction_date, s.id"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %v", err)
	}
	defer rows.Close()

	transactions := []TransactionDetail{}
	accountBalances := make(map[int]float64) // Track running balance per account
	accountTotals := make(map[int]struct {
		debit  float64
		credit float64
	})

	// Get opening balances (before start date)
	openingBalances := make(map[int]float64)
	if len(req.AccountIDs) > 0 {
		for _, accountID := range req.AccountIDs {
			// Calculate balance up to (but not including) start date
			balance, err := s.calculateAccountBalanceBeforeDate(accountID, req.StartDate)
			if err == nil {
				openingBalances[accountID] = balance
			}
		}
	}

	for rows.Next() {
		var split splits.Split
		var transactionDate, transactionType, description string
		err := rows.Scan(&split.ID, &split.TransactionID, &split.AccountID, &split.PeopleID, &split.Debit, &split.Credit, &split.Status, &transactionDate, &transactionType, &description)
		if err != nil {
			continue
		}

		// Get account name
		account, _, _, err := s.accountRepo.GetByID(split.AccountID)
		if err != nil {
			continue
		}

		// Get people name if available
		var peopleName *string
		if split.PeopleID != nil {
			person, _, _, err := s.peopleRepo.GetByID(*split.PeopleID)
			if err == nil {
				peopleName = &person.Name
			}
		}

		// Calculate running balance
		debitAmount := 0.0
		if split.Debit != nil {
			debitAmount = *split.Debit
		}
		creditAmount := 0.0
		if split.Credit != nil {
			creditAmount = *split.Credit
		}

		// Update totals
		if _, exists := accountTotals[split.AccountID]; !exists {
			accountTotals[split.AccountID] = struct {
				debit  float64
				credit float64
			}{debit: 0, credit: 0}
		}
		totals := accountTotals[split.AccountID]
		totals.debit += debitAmount
		totals.credit += creditAmount
		accountTotals[split.AccountID] = totals

		// Calculate running balance
		_, accountType, _, err := s.accountRepo.GetByID(split.AccountID)
		if err == nil {
			// Initialize balance if not exists
			if _, exists := accountBalances[split.AccountID]; !exists {
				accountBalances[split.AccountID] = openingBalances[split.AccountID]
			}
			// Update balance based on account typeStatus
			typeStatusLower := strings.ToLower(accountType.TypeStatus)
			if typeStatusLower == "debit" {
				// Debit accounts: Debit increases, Credit decreases
				accountBalances[split.AccountID] += debitAmount - creditAmount
			} else {
				// Credit accounts: Credit increases, Debit decreases
				accountBalances[split.AccountID] += creditAmount - debitAmount
			}
		}

		transactions = append(transactions, TransactionDetail{
			TransactionID:   split.TransactionID,
			TransactionDate: transactionDate,
			TransactionType: transactionType,
			AccountID:       split.AccountID,
			AccountName:     account.AccountName,
			PeopleID:        split.PeopleID,
			PeopleName:      peopleName,
			Debit:           split.Debit,
			Credit:          split.Credit,
			Balance:         accountBalances[split.AccountID],
			Description:     description,
		})
	}

	// Calculate totals
	totalDebit := 0.0
	totalCredit := 0.0
	for _, totals := range accountTotals {
		totalDebit += totals.debit
		totalCredit += totals.credit
	}

	// Calculate closing balance
	closingBalance := 0.0
	if len(req.AccountIDs) == 1 {
		// Single account - return single response
		accountID := req.AccountIDs[0]
		closingBalance = accountBalances[accountID]
		account, _, _, _ := s.accountRepo.GetByID(accountID)
		return &TransactionDetailsResponse{
			AccountID:      accountID,
			AccountName:    account.AccountName,
			StartDate:      req.StartDate,
			EndDate:        req.EndDate,
			OpeningBalance: openingBalances[accountID],
			ClosingBalance: closingBalance,
			TotalDebit:     totalDebit,
			TotalCredit:    totalCredit,
			Transactions:   transactions,
		}, nil
	} else {
		// Multiple accounts - group by account
		byAccount := make(map[int][]TransactionDetail)
		for _, txn := range transactions {
			byAccount[txn.AccountID] = append(byAccount[txn.AccountID], txn)
		}
		return &TransactionDetailsResponse{
			StartDate:    req.StartDate,
			EndDate:      req.EndDate,
			TotalDebit:   totalDebit,
			TotalCredit:  totalCredit,
			Transactions: transactions,
			ByAccount:    byAccount,
		}, nil
	}
}
