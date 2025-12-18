package reports

import (
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/mysecodgit/go_accounting/src/account_types"
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

	// Categorize accounts into Assets, Liabilities, Equity, Income, and Expense
	assets := []AccountBalance{}
	liabilities := []AccountBalance{}
	equity := []AccountBalance{}
	incomeAccounts := []AccountBalance{}
	expenseAccounts := []AccountBalance{}

	for i, account := range accountsList {
		accountType := accountTypes[i]
		balance := accountBalances[account.ID]

		// Skip accounts with 0 balance
		if balance == 0 {
			continue
		}

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
		} else if typeLower == "income" {
			incomeAccounts = append(incomeAccounts, accountBalance)
		} else if typeLower == "expense" {
			expenseAccounts = append(expenseAccounts, accountBalance)
		}
	}

	// Calculate Net Income = Total Income - Total Expenses
	totalIncome := 0.0
	for _, income := range incomeAccounts {
		totalIncome += income.Balance
	}

	totalExpenses := 0.0
	for _, expense := range expenseAccounts {
		totalExpenses += expense.Balance
	}

	netIncome := totalIncome - totalExpenses

	// Add Net Income to equity section (only if it's not zero)
	if netIncome != 0 {
		equity = append(equity, AccountBalance{
			AccountID:     0, // 0 indicates this is a calculated value, not an actual account
			AccountNumber: "",
			AccountName:   "Net Income",
			AccountType:   "Net Income",
			Balance:       netIncome,
		})
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

	// Round to 2 decimals for presentation + stable comparisons (avoid floating point artifacts)
	round2 := func(v float64) float64 {
		return math.Round(v*100) / 100
	}
	for i := range assets {
		assets[i].Balance = round2(assets[i].Balance)
	}
	for i := range liabilities {
		liabilities[i].Balance = round2(liabilities[i].Balance)
	}
	for i := range equity {
		equity[i].Balance = round2(equity[i].Balance)
	}
	totalAssets = round2(totalAssets)
	totalLiabilities = round2(totalLiabilities)
	totalEquity = round2(totalEquity)
	totalLiabilitiesAndEquity = round2(totalLiabilitiesAndEquity)

	// Consider balanced if equal at 2-decimal precision
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

		// Only include accounts that have debit or credit (active splits)
		if debitBalance > 0 || creditBalance > 0 {
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

// GetTransactionDetailsByAccount generates a transaction details report grouped by account
func (s *ReportsService) GetTransactionDetailsByAccount(req TransactionDetailsByAccountRequest) (*TransactionDetailsByAccountResponse, error) {
	if req.StartDate == "" || req.EndDate == "" {
		return nil, fmt.Errorf("start date and end date are required")
	}

	// Get accounts to process
	var accountsList []accounts.Account
	var accountTypesList []account_types.AccountType
	var err error

	if req.AccountID != nil && *req.AccountID > 0 {
		// Get specific account
		account, accountType, _, err := s.accountRepo.GetByID(*req.AccountID)
		if err != nil {
			return nil, fmt.Errorf("account not found: %v", err)
		}
		accountsList = []accounts.Account{account}
		accountTypesList = []account_types.AccountType{accountType}
	} else {
		// Get all accounts for the building
		accountsList, accountTypesList, _, err = s.accountRepo.GetByBuildingID(req.BuildingID)
		if err != nil {
			return nil, fmt.Errorf("failed to get accounts: %v", err)
		}
	}

	accountDetails := []AccountTransactionDetails{}
	grandTotalDebit := 0.0
	grandTotalCredit := 0.0

	// Process each account
	for i, account := range accountsList {
		accountType := accountTypesList[i]

		// Get splits for this account within date range (with optional unit filter)
		splits, err := s.splitRepo.GetByAccountIDAndDateRangeWithUnit(account.ID, req.BuildingID, req.StartDate, req.EndDate, req.UnitID)
		if err != nil {
			return nil, fmt.Errorf("failed to get splits for account %d: %v", account.ID, err)
		}

		if len(splits) == 0 {
			continue // Skip accounts with no transactions
		}

		// Get initial balance before start date (using typeStatus)
		initialBalance, err := s.calculateAccountBalanceBeforeDate(account.ID, req.StartDate)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate initial balance for account %d: %v", account.ID, err)
		}

		// Get account type status to determine how balance changes
		typeStatusLower := strings.ToLower(accountType.TypeStatus)

		// Build transaction details for each split
		splitDetails := []TransactionDetailSplit{}
		runningBalance := initialBalance
		accountTotalDebit := 0.0
		accountTotalCredit := 0.0

		for _, split := range splits {
			// Get transaction details
			transaction, err := s.transactionRepo.GetByID(split.TransactionID)
			if err != nil {
				return nil, fmt.Errorf("failed to get transaction %d: %v", split.TransactionID, err)
			}

			// Get people name if people_id exists
			var peopleName *string
			if split.PeopleID != nil {
				person, _, _, err := s.peopleRepo.GetByID(*split.PeopleID)
				if err == nil {
					peopleName = &person.Name
				}
			}

			// Calculate running balance based on account type status
			debitAmount := 0.0
			if split.Debit != nil {
				debitAmount = *split.Debit
				accountTotalDebit += debitAmount
			}

			creditAmount := 0.0
			if split.Credit != nil {
				creditAmount = *split.Credit
				accountTotalCredit += creditAmount
			}

			// Update running balance based on typeStatus
			// Debit accounts (Assets, Expenses): increase with debits, decrease with credits
			// Credit accounts (Liabilities, Equity, Income): increase with credits, decrease with debits
			if typeStatusLower == "debit" {
				runningBalance += debitAmount
				runningBalance -= creditAmount
			} else {
				// Credit account
				runningBalance += creditAmount
				runningBalance -= debitAmount
			}

			splitDetails = append(splitDetails, TransactionDetailSplit{
				SplitID:           split.ID,
				TransactionID:     split.TransactionID,
				TransactionNumber: transaction.TransactionNumber,
				TransactionDate:   transaction.TransactionDate,
				TransactionType:   transaction.Type,
				TransactionMemo:   transaction.Memo,
				PeopleID:          split.PeopleID,
				PeopleName:        peopleName,
				Description:       nil, // Can be enhanced later if needed
				Debit:             split.Debit,
				Credit:            split.Credit,
				Balance:           runningBalance,
			})
		}

		// Calculate final balance (last running balance)
		finalBalance := runningBalance

		// Add account details
		accountDetails = append(accountDetails, AccountTransactionDetails{
			AccountID:     account.ID,
			AccountNumber: account.AccountNumber,
			AccountName:   account.AccountName,
			AccountType:   accountType.TypeName,
			Splits:        splitDetails,
			TotalDebit:    accountTotalDebit,
			TotalCredit:   accountTotalCredit,
			TotalBalance:  finalBalance,
			IsTotalRow:    false,
		})

		// Add total row for this account
		accountDetails = append(accountDetails, AccountTransactionDetails{
			AccountID:     account.ID,
			AccountNumber: account.AccountNumber,
			AccountName:   "TOTAL",
			AccountType:   "",
			Splits:        []TransactionDetailSplit{},
			TotalDebit:    accountTotalDebit,
			TotalCredit:   accountTotalCredit,
			TotalBalance:  finalBalance,
			IsTotalRow:    true,
		})

		grandTotalDebit += accountTotalDebit
		grandTotalCredit += accountTotalCredit
	}

	return &TransactionDetailsByAccountResponse{
		BuildingID:       req.BuildingID,
		StartDate:        req.StartDate,
		EndDate:          req.EndDate,
		Accounts:         accountDetails,
		GrandTotalDebit:  grandTotalDebit,
		GrandTotalCredit: grandTotalCredit,
	}, nil
}

// GetCustomerBalanceSummary generates a customer balance summary from splits
func (s *ReportsService) GetCustomerBalanceSummary(req CustomerBalanceSummaryRequest) (*CustomerBalanceSummaryResponse, error) {
	asOfDate := req.AsOfDate
	if asOfDate == "" {
		asOfDate = time.Now().Format("2006-01-02")
	}

	// Get all people types to find customer type
	peopleTypes, err := s.peopleTypeRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get people types: %v", err)
	}

	// Find the "customer" people type
	var customerTypeID *int
	for _, pt := range peopleTypes {
		if strings.ToLower(pt.Title) == "customer" {
			customerTypeID = &pt.ID
			break
		}
	}

	if customerTypeID == nil {
		return nil, fmt.Errorf("customer people type not found")
	}

	// Get all people for the building
	allPeople, peopleTypesList, _, err := s.peopleRepo.GetByBuildingID(req.BuildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get people: %v", err)
	}

	// Filter to only customers
	customers := []people.Person{}
	for i, person := range allPeople {
		if i < len(peopleTypesList) && peopleTypesList[i].ID == *customerTypeID {
			customers = append(customers, person)
		}
	}

	customerBalances := []CustomerBalance{}
	totalBalance := 0.0

	// Find Account Receivable account type
	arAccountTypeID, err := s.findAccountTypeByName("Account Receivable")
	if err != nil {
		return nil, fmt.Errorf("failed to find Account Receivable account type: %v", err)
	}

	// Calculate balance for each customer from splits on Account Receivable accounts only
	for _, customer := range customers {
		// Get all splits for this customer up to asOfDate, only for Account Receivable accounts
		query := `
			SELECT 
				COALESCE(SUM(CASE WHEN s.debit IS NOT NULL THEN s.debit ELSE 0 END), 0) as total_debit,
				COALESCE(SUM(CASE WHEN s.credit IS NOT NULL THEN s.credit ELSE 0 END), 0) as total_credit
			FROM splits s
			INNER JOIN transactions t ON s.transaction_id = t.id
			INNER JOIN accounts a ON s.account_id = a.id
			WHERE s.people_id = ?
				AND a.account_type = ?
				AND s.status = '1'
				AND t.status = '1'
				AND DATE(t.transaction_date) <= ?
		`

		var totalDebit, totalCredit sql.NullFloat64
		err := s.db.QueryRow(query, customer.ID, arAccountTypeID, asOfDate).Scan(&totalDebit, &totalCredit)
		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to calculate balance for customer %d: %v", customer.ID, err)
		}

		debitAmount := 0.0
		if totalDebit.Valid {
			debitAmount = totalDebit.Float64
		}
		creditAmount := 0.0
		if totalCredit.Valid {
			creditAmount = totalCredit.Float64
		}

		// Balance = Debit - Credit (positive means customer owes us, negative means we owe customer)
		balance := debitAmount - creditAmount

		// Only include customers with non-zero balance
		if balance != 0 {
			customerBalances = append(customerBalances, CustomerBalance{
				PeopleID:   customer.ID,
				PeopleName: customer.Name,
				Balance:    balance,
			})
			totalBalance += balance
		}
	}

	return &CustomerBalanceSummaryResponse{
		BuildingID:   req.BuildingID,
		AsOfDate:     asOfDate,
		Customers:    customerBalances,
		TotalBalance: totalBalance,
	}, nil
}

// GetCustomerBalanceDetails generates a detailed customer balance report with transaction splits
func (s *ReportsService) GetCustomerBalanceDetails(req CustomerBalanceDetailsRequest) (*CustomerBalanceDetailsResponse, error) {
	asOfDate := req.AsOfDate
	if asOfDate == "" {
		asOfDate = time.Now().Format("2006-01-02")
	}

	// Get all people types to find customer type
	peopleTypes, err := s.peopleTypeRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get people types: %v", err)
	}

	// Find the "customer" people type
	var customerTypeID *int
	for _, pt := range peopleTypes {
		if strings.ToLower(pt.Title) == "customer" {
			customerTypeID = &pt.ID
			break
		}
	}

	if customerTypeID == nil {
		return nil, fmt.Errorf("customer people type not found")
	}

	// Get all people for the building
	allPeople, peopleTypesList, _, err := s.peopleRepo.GetByBuildingID(req.BuildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get people: %v", err)
	}

	// Filter to only customers
	customers := []people.Person{}
	for i, person := range allPeople {
		if i < len(peopleTypesList) && peopleTypesList[i].ID == *customerTypeID {
			// If PeopleID filter is specified, only include that customer
			if req.PeopleID == nil || *req.PeopleID == person.ID {
				customers = append(customers, person)
			}
		}
	}

	customerDetails := []CustomerBalanceDetails{}
	grandTotalDebit := 0.0
	grandTotalCredit := 0.0
	grandTotalBalance := 0.0

	// Find Account Receivable account type
	arAccountTypeID, err := s.findAccountTypeByName("Account Receivable")
	if err != nil {
		return nil, fmt.Errorf("failed to find Account Receivable account type: %v", err)
	}

	// Process each customer
	for _, customer := range customers {
		// Get all splits for this customer up to asOfDate, only for Account Receivable accounts, ordered by transaction date
		query := `
			SELECT 
				s.id as split_id,
				s.transaction_id,
				s.account_id,
				s.debit,
				s.credit,
				t.transaction_date,
				t.type as transaction_type,
				t.transaction_number,
				t.memo as transaction_memo,
				a.account_name,
				a.account_number
			FROM splits s
			INNER JOIN transactions t ON s.transaction_id = t.id
			INNER JOIN accounts a ON s.account_id = a.id
			WHERE s.people_id = ?
				AND a.account_type = ?
				AND s.status = '1'
				AND t.status = '1'
				AND DATE(t.transaction_date) <= ?
			ORDER BY t.transaction_date, t.id, s.id
		`

		rows, err := s.db.Query(query, customer.ID, arAccountTypeID, asOfDate)
		if err != nil {
			return nil, fmt.Errorf("failed to get splits for customer %d: %v", customer.ID, err)
		}
		defer rows.Close()

		// Group splits by account
		accountMap := make(map[int]*CustomerBalanceAccount)
		accountOrder := []int{} // To maintain order

		customerTotalDebit := 0.0
		customerTotalCredit := 0.0
		customerRunningBalance := 0.0

		for rows.Next() {
			var split CustomerBalanceDetailSplit
			var accountNumber sql.NullInt64
			var debit, credit sql.NullFloat64

			err := rows.Scan(
				&split.SplitID,
				&split.TransactionID,
				&split.AccountID,
				&debit,
				&credit,
				&split.TransactionDate,
				&split.TransactionType,
				&split.TransactionNumber,
				&split.TransactionMemo,
				&split.AccountName,
				&accountNumber,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to scan split: %v", err)
			}

			if accountNumber.Valid {
				split.AccountNumber = int(accountNumber.Int64)
			}

			// Initialize account if not exists
			if _, exists := accountMap[split.AccountID]; !exists {
				accountMap[split.AccountID] = &CustomerBalanceAccount{
					AccountID:     split.AccountID,
					AccountName:   split.AccountName,
					AccountNumber: split.AccountNumber,
					Splits:        []CustomerBalanceDetailSplit{},
					TotalDebit:    0.0,
					TotalCredit:   0.0,
					TotalBalance:  0.0,
					IsTotalRow:    false,
				}
				accountOrder = append(accountOrder, split.AccountID)
			}

			account := accountMap[split.AccountID]

			if debit.Valid {
				debitValue := debit.Float64
				split.Debit = &debitValue
				account.TotalDebit += debit.Float64
				customerTotalDebit += debit.Float64
				customerRunningBalance += debit.Float64
			}

			if credit.Valid {
				creditValue := credit.Float64
				split.Credit = &creditValue
				account.TotalCredit += credit.Float64
				customerTotalCredit += credit.Float64
				customerRunningBalance -= credit.Float64
			}

			// Running balance is calculated at customer level, not account level
			split.Balance = customerRunningBalance
			account.Splits = append(account.Splits, split)
		}

		// Only include customers with transactions
		if len(accountMap) > 0 {
			// Build accounts list in order and calculate account balances
			accounts := []CustomerBalanceAccount{}
			for _, accountID := range accountOrder {
				account := accountMap[accountID]
				// Calculate account balance: debit - credit
				account.TotalBalance = account.TotalDebit - account.TotalCredit
				accounts = append(accounts, *account)
			}

			finalBalance := customerRunningBalance

			// Add customer header
			customerDetails = append(customerDetails, CustomerBalanceDetails{
				PeopleID:     customer.ID,
				PeopleName:   customer.Name,
				Accounts:     accounts,
				TotalDebit:   customerTotalDebit,
				TotalCredit:  customerTotalCredit,
				TotalBalance: finalBalance,
				IsTotalRow:   false,
				IsHeader:    true,
			})

			// Add total row for this customer
			customerDetails = append(customerDetails, CustomerBalanceDetails{
				PeopleID:     customer.ID,
				PeopleName:   "TOTAL",
				Accounts:     []CustomerBalanceAccount{},
				TotalDebit:   customerTotalDebit,
				TotalCredit:  customerTotalCredit,
				TotalBalance: finalBalance,
				IsTotalRow:   true,
				IsHeader:    false,
			})

			grandTotalDebit += customerTotalDebit
			grandTotalCredit += customerTotalCredit
			grandTotalBalance += finalBalance
		}
	}

	return &CustomerBalanceDetailsResponse{
		BuildingID:        req.BuildingID,
		AsOfDate:          asOfDate,
		Customers:         customerDetails,
		GrandTotalDebit:   grandTotalDebit,
		GrandTotalCredit:  grandTotalCredit,
		GrandTotalBalance: grandTotalBalance,
	}, nil
}

// calculateAccountBalanceForDateRange calculates the balance of an account for a specific date range
func (s *ReportsService) calculateAccountBalanceForDateRange(accountID int, startDate string, endDate string, unitID *int) (float64, error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN s.debit IS NOT NULL THEN s.debit ELSE 0 END), 0) as total_debit,
			COALESCE(SUM(CASE WHEN s.credit IS NOT NULL THEN s.credit ELSE 0 END), 0) as total_credit
		FROM splits s
		INNER JOIN transactions t ON s.transaction_id = t.id
		WHERE s.account_id = ? 
			AND s.status = '1'
			AND t.status = '1'
			AND DATE(t.transaction_date) >= ?
			AND DATE(t.transaction_date) <= ?
	`
	
	args := []interface{}{accountID, startDate, endDate}
	
	// Add unit filter if provided
	if unitID != nil && *unitID > 0 {
		query += ` AND s.unit_id = ?`
		args = append(args, *unitID)
	}

	var totalDebit, totalCredit sql.NullFloat64
	err := s.db.QueryRow(query, args...).Scan(&totalDebit, &totalCredit)
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
	typeStatusLower := strings.ToLower(accountType.TypeStatus)

	// Debit accounts: Debit - Credit (Assets, Expenses)
	// Credit accounts: Credit - Debit (Liabilities, Equity, Income)
	if typeStatusLower == "debit" {
		return debitAmount - creditAmount, nil
	} else {
		return creditAmount - debitAmount, nil
	}
}

// GetProfitAndLossStandard generates a standard profit and loss report
func (s *ReportsService) GetProfitAndLossStandard(req ProfitAndLossStandardRequest) (*ProfitAndLossStandardResponse, error) {
	if req.StartDate == "" || req.EndDate == "" {
		return nil, fmt.Errorf("start date and end date are required")
	}

	// Get all accounts for the building
	accountsList, accountTypes, _, err := s.accountRepo.GetByBuildingID(req.BuildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %v", err)
	}

	incomeAccounts := []ProfitAndLossAccount{}
	expenseAccounts := []ProfitAndLossAccount{}
	totalIncome := 0.0
	totalExpenses := 0.0

	for i, account := range accountsList {
		accountType := accountTypes[i]
		typeLower := strings.ToLower(accountType.Type)

		// Only process Income and Expense accounts
		if typeLower != "income" && typeLower != "expense" {
			continue
		}

		// Calculate balance for the date range
		balance, err := s.calculateAccountBalanceForDateRange(account.ID, req.StartDate, req.EndDate, nil)
		if err != nil {
			continue // Skip accounts with errors
		}

		// Skip accounts with 0 balance
		if balance == 0 {
			continue
		}

		accountBalance := ProfitAndLossAccount{
			AccountID:     account.ID,
			AccountNumber: account.AccountNumber,
			AccountName:   account.AccountName,
			Balance:       balance,
		}

		if typeLower == "income" {
			incomeAccounts = append(incomeAccounts, accountBalance)
			totalIncome += balance
		} else if typeLower == "expense" {
			expenseAccounts = append(expenseAccounts, accountBalance)
			totalExpenses += balance
		}
	}

	netProfitLoss := totalIncome - totalExpenses

	return &ProfitAndLossStandardResponse{
		BuildingID:    req.BuildingID,
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		Income:        ProfitAndLossSection{SectionName: "Income", Accounts: incomeAccounts, Total: totalIncome},
		Expenses:      ProfitAndLossSection{SectionName: "Expenses", Accounts: expenseAccounts, Total: totalExpenses},
		NetProfitLoss: netProfitLoss,
	}, nil
}

// GetProfitAndLossByUnit generates a profit and loss report grouped by unit (pivot table style)
func (s *ReportsService) GetProfitAndLossByUnit(req ProfitAndLossByUnitRequest) (*ProfitAndLossByUnitResponse, error) {
	if req.StartDate == "" || req.EndDate == "" {
		return nil, fmt.Errorf("start date and end date are required")
	}

	// Get all units for the building
	query := `SELECT id, name FROM units WHERE building_id = ? ORDER BY name`
	rows, err := s.db.Query(query, req.BuildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get units: %v", err)
	}
	defer rows.Close()

	type Unit struct {
		ID   int
		Name string
	}
	allUnits := []Unit{}
	for rows.Next() {
		var unit Unit
		if err := rows.Scan(&unit.ID, &unit.Name); err != nil {
			continue
		}
		allUnits = append(allUnits, unit)
	}

	// Get all accounts for the building
	accountsList, accountTypes, _, err := s.accountRepo.GetByBuildingID(req.BuildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %v", err)
	}

	// Build unit columns (only include units that have transactions)
	unitColumns := []UnitColumn{}
	unitsWithData := make(map[int]bool)

	// First pass: collect all units that have income or expense data
	for _, unit := range allUnits {
		unitID := &unit.ID
		hasData := false

		for i, account := range accountsList {
			accountType := accountTypes[i]
			typeLower := strings.ToLower(accountType.Type)
			if typeLower != "income" && typeLower != "expense" {
				continue
			}

			balance, err := s.calculateAccountBalanceForDateRange(account.ID, req.StartDate, req.EndDate, unitID)
			if err == nil && balance != 0 {
				hasData = true
				break
			}
		}

		if hasData {
			unitsWithData[unit.ID] = true
			unitColumns = append(unitColumns, UnitColumn{
				UnitID:   unit.ID,
				UnitName: unit.Name,
			})
		}
	}

	// Build income account rows
	incomeAccounts := []AccountRow{}
	incomeAccountMap := make(map[int]*AccountRow) // account_id -> AccountRow

	for i, account := range accountsList {
		accountType := accountTypes[i]
		typeLower := strings.ToLower(accountType.Type)
		if typeLower != "income" {
			continue
		}

		balances := make(map[int]float64)
		total := 0.0

		for _, unit := range unitColumns {
			unitID := &unit.UnitID
			balance, err := s.calculateAccountBalanceForDateRange(account.ID, req.StartDate, req.EndDate, unitID)
			if err == nil && balance != 0 {
				balances[unit.UnitID] = balance
				total += balance
			}
		}

		// Only include accounts that have data
		if total != 0 {
			accountRow := AccountRow{
				AccountID:     account.ID,
				AccountNumber: account.AccountNumber,
				AccountName:   account.AccountName,
				AccountType:   "income",
				Balances:      balances,
				Total:         total,
			}
			incomeAccounts = append(incomeAccounts, accountRow)
			incomeAccountMap[account.ID] = &incomeAccounts[len(incomeAccounts)-1]
		}
	}

	// Build expense account rows
	expenseAccounts := []AccountRow{}
	expenseAccountMap := make(map[int]*AccountRow) // account_id -> AccountRow

	for i, account := range accountsList {
		accountType := accountTypes[i]
		typeLower := strings.ToLower(accountType.Type)
		if typeLower != "expense" {
			continue
		}

		balances := make(map[int]float64)
		total := 0.0

		for _, unit := range unitColumns {
			unitID := &unit.UnitID
			balance, err := s.calculateAccountBalanceForDateRange(account.ID, req.StartDate, req.EndDate, unitID)
			if err == nil && balance != 0 {
				balances[unit.UnitID] = balance
				total += balance
			}
		}

		// Only include accounts that have data
		if total != 0 {
			accountRow := AccountRow{
				AccountID:     account.ID,
				AccountNumber: account.AccountNumber,
				AccountName:   account.AccountName,
				AccountType:   "expense",
				Balances:      balances,
				Total:         total,
			}
			expenseAccounts = append(expenseAccounts, accountRow)
			expenseAccountMap[account.ID] = &expenseAccounts[len(expenseAccounts)-1]
		}
	}

	// Calculate totals per unit
	totalIncome := make(map[int]float64)
	totalExpenses := make(map[int]float64)
	netProfitLoss := make(map[int]float64)

	for _, unit := range unitColumns {
		unitIncome := 0.0
		unitExpenses := 0.0

		for _, accountRow := range incomeAccounts {
			if balance, ok := accountRow.Balances[unit.UnitID]; ok {
				unitIncome += balance
			}
		}

		for _, accountRow := range expenseAccounts {
			if balance, ok := accountRow.Balances[unit.UnitID]; ok {
				unitExpenses += balance
			}
		}

		totalIncome[unit.UnitID] = unitIncome
		totalExpenses[unit.UnitID] = unitExpenses
		netProfitLoss[unit.UnitID] = unitIncome - unitExpenses
	}

	// Calculate grand totals
	grandTotalIncome := 0.0
	grandTotalExpenses := 0.0
	for _, total := range totalIncome {
		grandTotalIncome += total
	}
	for _, total := range totalExpenses {
		grandTotalExpenses += total
	}
	grandTotalNetProfitLoss := grandTotalIncome - grandTotalExpenses

	return &ProfitAndLossByUnitResponse{
		BuildingID:             req.BuildingID,
		StartDate:              req.StartDate,
		EndDate:                req.EndDate,
		Units:                  unitColumns,
		IncomeAccounts:         incomeAccounts,
		ExpenseAccounts:        expenseAccounts,
		TotalIncome:             totalIncome,
		TotalExpenses:           totalExpenses,
		NetProfitLoss:          netProfitLoss,
		GrandTotalIncome:       grandTotalIncome,
		GrandTotalExpenses:     grandTotalExpenses,
		GrandTotalNetProfitLoss: grandTotalNetProfitLoss,
	}, nil
}
