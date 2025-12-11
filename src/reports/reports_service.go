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
	"github.com/mysecodgit/go_accounting/src/people_types"
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
	peopleTypeRepo  people_types.PeopleTypeRepository
	db              *sql.DB
}

func NewReportsService(
	accountRepo accounts.AccountRepository,
	splitRepo splits.SplitRepository,
	transactionRepo transactions.TransactionRepository,
	invoiceRepo invoices.InvoiceRepository,
	paymentRepo invoice_payments.InvoicePaymentRepository,
	peopleRepo people.PersonRepository,
	peopleTypeRepo people_types.PeopleTypeRepository,
	db *sql.DB,
) *ReportsService {
	return &ReportsService{
		accountRepo:     accountRepo,
		splitRepo:       splitRepo,
		transactionRepo: transactionRepo,
		invoiceRepo:     invoiceRepo,
		paymentRepo:     paymentRepo,
		peopleRepo:      peopleRepo,
		peopleTypeRepo:  peopleTypeRepo,
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
	var peopleTypes []people_types.PeopleType
	var err error

	if req.PeopleID != nil {
		person, peopleType, _, err := s.peopleRepo.GetByID(*req.PeopleID)
		if err != nil {
			return nil, fmt.Errorf("person not found: %v", err)
		}
		peopleList = []people.Person{person}
		peopleTypes = []people_types.PeopleType{peopleType}
	} else {
		peopleList, peopleTypes, _, err = s.peopleRepo.GetByBuildingID(req.BuildingID)
		if err != nil {
			return nil, fmt.Errorf("failed to get people: %v", err)
		}
	}

	// Filter by type_id if provided
	if req.TypeID != nil {
		filteredPeople := []people.Person{}
		filteredTypes := []people_types.PeopleType{}
		for i, person := range peopleList {
			if person.TypeID == *req.TypeID {
				filteredPeople = append(filteredPeople, person)
				filteredTypes = append(filteredTypes, peopleTypes[i])
			}
		}
		peopleList = filteredPeople
		peopleTypes = filteredTypes
	}

	// Find Account Receivable and Account Payable account types
	arAccountTypeID, err := s.findAccountTypeByName("Account Receivable")
	if err != nil {
		return nil, fmt.Errorf("failed to find Account Receivable account type: %v", err)
	}
	apAccountTypeID, err := s.findAccountTypeByName("Account Payable")
	if err != nil {
		return nil, fmt.Errorf("failed to find Account Payable account type: %v", err)
	}

	// Get all accounts for the building
	accountsList, accountTypesList, _, err := s.accountRepo.GetByBuildingID(req.BuildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %v", err)
	}

	// Find AR and AP accounts
	arAccountIDs := []int{}
	apAccountIDs := []int{}
	for i, account := range accountsList {
		if i < len(accountTypesList) && accountTypesList[i].ID == arAccountTypeID {
			arAccountIDs = append(arAccountIDs, account.ID)
		}
		if i < len(accountTypesList) && accountTypesList[i].ID == apAccountTypeID {
			apAccountIDs = append(apAccountIDs, account.ID)
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
				Description:     "Payment for Invoice",
				Amount:          payment.Amount,
			})
		}

		// Get people type name from the peopleTypes array
		peopleTypeName := "Customer" // Default
		for i, p := range peopleList {
			if p.ID == person.ID && i < len(peopleTypes) {
				peopleTypeName = peopleTypes[i].Title
				break
			}
		}

		// Determine which account type to use based on people type
		var accountIDsToUse []int
		isCustomer := strings.Contains(strings.ToLower(peopleTypeName), "customer")

		if isCustomer {
			// For customers: use Account Receivable
			accountIDsToUse = arAccountIDs
		} else {
			// For vendors: use Account Payable
			accountIDsToUse = apAccountIDs
		}

		// Calculate balance from splits up to end date
		balance, err := s.calculatePersonBalanceFromSplits(person.ID, accountIDsToUse, req.EndDate)
		if err != nil {
			// Fallback to invoice - payment calculation if split calculation fails
			balance = totalInvoices - totalPayments
		}

		outstanding := balance
		totalOutstanding += outstanding

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
			AND status = '1'
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
			AND ip.status = '1'
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

// FindPeopleTypeByTitle finds a people type by title (case-insensitive)
func (s *ReportsService) FindPeopleTypeByTitle(title string) (int, error) {
	allTypes, err := s.peopleTypeRepo.GetAll()
	if err != nil {
		return 0, err
	}

	titleLower := strings.ToLower(title)
	for _, pt := range allTypes {
		if strings.ToLower(pt.Title) == titleLower {
			return pt.ID, nil
		}
	}

	return 0, fmt.Errorf("people type '%s' not found", title)
}

// findAccountTypeByName finds an account type by typeName (case-insensitive)
func (s *ReportsService) findAccountTypeByName(typeName string) (int, error) {
	query := `
		SELECT id FROM account_types 
		WHERE LOWER(typeName) LIKE LOWER(?)
		LIMIT 1
	`

	var accountTypeID int
	err := s.db.QueryRow(query, "%"+typeName+"%").Scan(&accountTypeID)
	if err != nil {
		return 0, fmt.Errorf("account type '%s' not found: %v", typeName, err)
	}

	return accountTypeID, nil
}

// GetTrialBalance generates a trial balance report
func (s *ReportsService) GetTrialBalance(req TrialBalanceRequest) (*TrialBalanceResponse, error) {
	// Get all accounts for the building
	accountsList, accountTypesList, _, err := s.accountRepo.GetByBuildingID(req.BuildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %v", err)
	}

	trialBalanceAccounts := []TrialBalanceAccount{}
	totalDebit := 0.0
	totalCredit := 0.0

	for i, account := range accountsList {
		if i >= len(accountTypesList) {
			continue
		}
		accountType := accountTypesList[i]

		// Calculate account balance up to as_of_date
		balance, err := s.calculateAccountBalance(account.ID, req.AsOfDate)
		if err != nil {
			continue
		}

		// Determine debit and credit balances based on account type
		var debitBalance, creditBalance float64
		typeStatusLower := strings.ToLower(accountType.TypeStatus)

		if typeStatusLower == "debit" {
			// Debit accounts: positive balance = debit, negative = credit
			if balance >= 0 {
				debitBalance = balance
				creditBalance = 0
			} else {
				debitBalance = 0
				creditBalance = -balance // Make it positive for display
			}
		} else {
			// Credit accounts: positive balance = credit, negative = debit
			if balance >= 0 {
				debitBalance = 0
				creditBalance = balance
			} else {
				debitBalance = -balance // Make it positive for display
				creditBalance = 0
			}
		}

		// Include all accounts (even with zero balance)
		trialBalanceAccounts = append(trialBalanceAccounts, TrialBalanceAccount{
			AccountID:     account.ID,
			AccountNumber: account.AccountNumber,
			AccountName:   account.AccountName,
			AccountType:   accountType.TypeName,
			DebitBalance:  debitBalance,
			CreditBalance: creditBalance,
		})

		totalDebit += debitBalance
		totalCredit += creditBalance
	}

	// Check if balanced (allowing for small rounding differences)
	isBalanced := (totalDebit-totalCredit) < 0.01 && (totalDebit-totalCredit) > -0.01

	// Add total row
	trialBalanceAccounts = append(trialBalanceAccounts, TrialBalanceAccount{
		AccountID:     0,
		AccountNumber: 0,
		AccountName:   "TOTAL",
		AccountType:   "",
		DebitBalance:  totalDebit,
		CreditBalance: totalCredit,
		IsTotalRow:    true,
	})

	return &TrialBalanceResponse{
		BuildingID:  req.BuildingID,
		AsOfDate:    req.AsOfDate,
		Accounts:    trialBalanceAccounts,
		TotalDebit:  totalDebit,
		TotalCredit: totalCredit,
		IsBalanced:  isBalanced,
	}, nil
}

// calculatePersonBalanceFromSplits calculates the balance for a person from splits
// For Account Receivable (debit account): Debit - Credit
// For Account Payable (credit account): Credit - Debit
func (s *ReportsService) calculatePersonBalanceFromSplits(peopleID int, accountIDs []int, asOfDate string) (float64, error) {
	if len(accountIDs) == 0 {
		return 0.0, nil
	}

	// Build placeholders for account IDs
	placeholders := strings.Repeat("?,", len(accountIDs))
	placeholders = placeholders[:len(placeholders)-1]

	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN s.debit IS NOT NULL THEN s.debit ELSE 0 END), 0) as total_debit,
			COALESCE(SUM(CASE WHEN s.credit IS NOT NULL THEN s.credit ELSE 0 END), 0) as total_credit
		FROM splits s
		INNER JOIN transactions t ON s.transaction_id = t.id
		INNER JOIN accounts a ON s.account_id = a.id
		INNER JOIN account_types at ON a.account_type = at.id
		WHERE s.people_id = ?
			AND s.account_id IN (` + placeholders + `)
			AND s.status = '1'
			AND t.status = '1'
			AND DATE(t.transaction_date) <= ?
	`

	args := []interface{}{peopleID}
	for _, accountID := range accountIDs {
		args = append(args, accountID)
	}
	args = append(args, asOfDate)

	var totalDebit, totalCredit sql.NullFloat64
	err := s.db.QueryRow(query, args...).Scan(&totalDebit, &totalCredit)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	debitAmount := 0.0
	if totalDebit.Valid {
		debitAmount = totalDebit.Float64
	}
	creditAmount := 0.0
	if totalCredit.Valid {
		creditAmount = totalCredit.Float64
	}

	// Determine if it's a debit or credit account by checking the account type
	// Account Receivable is typically a debit account (Asset)
	// Account Payable is typically a credit account (Liability)
	// We'll check the first account's type to determine the calculation
	if len(accountIDs) > 0 {
		_, accountType, _, err := s.accountRepo.GetByID(accountIDs[0])
		if err == nil {
			typeStatusLower := strings.ToLower(accountType.TypeStatus)
			typeNameLower := strings.ToLower(accountType.TypeName)

			// Account Receivable: Debit - Credit (debit account)
			// Account Payable: Credit - Debit (credit account)
			if strings.Contains(typeNameLower, "receivable") || typeStatusLower == "debit" {
				return debitAmount - creditAmount, nil
			} else if strings.Contains(typeNameLower, "payable") || typeStatusLower == "credit" {
				return creditAmount - debitAmount, nil
			}
		}
	}

	// Default: assume debit account (Account Receivable)
	return debitAmount - creditAmount, nil
}
