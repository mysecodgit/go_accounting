package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/mysecodgit/go_accounting/config"
	_ "github.com/mysecodgit/go_accounting/handlers"
	"github.com/mysecodgit/go_accounting/src/account_types"
	"github.com/mysecodgit/go_accounting/src/accounts"
	"github.com/mysecodgit/go_accounting/src/building"
	"github.com/mysecodgit/go_accounting/src/checks"
	"github.com/mysecodgit/go_accounting/src/credit_memo"
	"github.com/mysecodgit/go_accounting/src/expense_lines"
	"github.com/mysecodgit/go_accounting/src/invoice_applied_credits"
	"github.com/mysecodgit/go_accounting/src/invoice_items"
	"github.com/mysecodgit/go_accounting/src/invoice_payments"
	"github.com/mysecodgit/go_accounting/src/invoices"
	"github.com/mysecodgit/go_accounting/src/items"
	"github.com/mysecodgit/go_accounting/src/journal"
	"github.com/mysecodgit/go_accounting/src/journal_lines"
	"github.com/mysecodgit/go_accounting/src/people"
	"github.com/mysecodgit/go_accounting/src/people_types"
	"github.com/mysecodgit/go_accounting/src/period"
	"github.com/mysecodgit/go_accounting/src/receipt_items"
	"github.com/mysecodgit/go_accounting/src/reports"
	"github.com/mysecodgit/go_accounting/src/sales_receipt"
	"github.com/mysecodgit/go_accounting/src/splits"
	"github.com/mysecodgit/go_accounting/src/transactions"
	"github.com/mysecodgit/go_accounting/src/unit"
	"github.com/mysecodgit/go_accounting/src/user"
)

func SetupRoutes(r *gin.Engine) {

	userRepo := user.NewUserRepository(config.DB)
	userService := user.NewUserService(userRepo)
	userHandler := user.NewUserHandler(userService)

	userRoutes := r.Group("/api/users")
	{
		userRoutes.GET("", userHandler.GetUsers)
		userRoutes.GET("/:id", userHandler.GetUser)
		userRoutes.POST("", userHandler.CreateUser)
		userRoutes.PUT("/:id", userHandler.UpdateUser)
		// userRoutes.DELETE("/:id", handlers.DeleteUser)
	}

	buildingRepo := building.NewBuildingRepository(config.DB)
	buildingService := building.NewBuildingService(buildingRepo)
	buildingHandler := building.NewBuildingHandler(buildingService)

	// Initialize invoice dependencies (used in both building-scoped and legacy routes)
	transactionRepo := transactions.NewTransactionRepository(config.DB)
	splitRepo := splits.NewSplitRepository(config.DB)
	invoiceItemRepo := invoice_items.NewInvoiceItemRepository(config.DB)
	itemRepoForInvoice := items.NewItemRepository(config.DB)
	accountRepoForInvoice := accounts.NewAccountRepository(config.DB)
	invoiceRepo := invoices.NewInvoiceRepository(config.DB)
	invoiceService := invoices.NewInvoiceService(invoiceRepo, transactionRepo, splitRepo, invoiceItemRepo, itemRepoForInvoice, accountRepoForInvoice, config.DB)
	invoiceHandler := invoices.NewInvoiceHandler(invoiceService)

	// Initialize sales receipt dependencies
	receiptItemRepo := receipt_items.NewReceiptItemRepository(config.DB)
	itemRepoForReceipt := items.NewItemRepository(config.DB)
	accountRepoForReceipt := accounts.NewAccountRepository(config.DB)
	receiptRepo := sales_receipt.NewSalesReceiptRepository(config.DB)
	receiptService := sales_receipt.NewSalesReceiptService(receiptRepo, transactionRepo, splitRepo, receiptItemRepo, itemRepoForReceipt, accountRepoForReceipt, config.DB)
	receiptHandler := sales_receipt.NewSalesReceiptHandler(receiptService)

	// Initialize invoice payment dependencies
	paymentRepo := invoice_payments.NewInvoicePaymentRepository(config.DB)
	paymentService := invoice_payments.NewInvoicePaymentService(paymentRepo, transactionRepo, splitRepo, invoiceRepo, accountRepoForInvoice, config.DB)
	paymentHandler := invoice_payments.NewInvoicePaymentHandler(paymentService)

	// Initialize checks dependencies
	checkRepo := checks.NewCheckRepository(config.DB)
	expenseLineRepo := expense_lines.NewExpenseLineRepository(config.DB)
	accountTypeRepoForChecks := account_types.NewAccountTypeRepository(config.DB)
	checkService := checks.NewCheckService(checkRepo, expenseLineRepo, transactionRepo, splitRepo, accountRepoForInvoice, accountTypeRepoForChecks, config.DB)
	checkHandler := checks.NewCheckHandler(checkService)

	// Initialize credit memo dependencies
	creditMemoRepo := credit_memo.NewCreditMemoRepository(config.DB)
	peopleRepoForCreditMemo := people.NewPersonRepository(config.DB)
	accountTypeRepoForCreditMemo := account_types.NewAccountTypeRepository(config.DB)
	creditMemoService := credit_memo.NewCreditMemoService(creditMemoRepo, transactionRepo, splitRepo, accountRepoForInvoice, accountTypeRepoForCreditMemo, peopleRepoForCreditMemo, config.DB)
	creditMemoHandler := credit_memo.NewCreditMemoHandler(creditMemoService)

	// Initialize invoice applied credits dependencies
	appliedCreditRepo := invoice_applied_credits.NewInvoiceAppliedCreditRepository(config.DB)
	appliedCreditService := invoice_applied_credits.NewInvoiceAppliedCreditService(appliedCreditRepo, invoiceRepo, creditMemoRepo, transactionRepo, splitRepo, accountRepoForInvoice, config.DB)
	appliedCreditHandler := invoice_applied_credits.NewInvoiceAppliedCreditHandler(appliedCreditService)

	// Initialize journal dependencies
	journalRepo := journal.NewJournalRepository(config.DB)
	journalLineRepo := journal_lines.NewJournalLineRepository(config.DB)
	accountTypeRepoForJournal := account_types.NewAccountTypeRepository(config.DB)
	journalService := journal.NewJournalService(journalRepo, journalLineRepo, transactionRepo, splitRepo, accountRepoForInvoice, accountTypeRepoForJournal, config.DB)
	journalHandler := journal.NewJournalHandler(journalService)

	// Initialize reports dependencies
	peopleRepo := people.NewPersonRepository(config.DB)
	peopleTypeRepoForReports := people_types.NewPeopleTypeRepository(config.DB)
	reportsService := reports.NewReportsService(accountRepoForInvoice, splitRepo, transactionRepo, invoiceRepo, paymentRepo, peopleRepo, peopleTypeRepoForReports, config.DB)
	reportsHandler := reports.NewReportsHandler(reportsService)

	buildingRoutes := r.Group("/api/buildings")
	{
		buildingRoutes.GET("", buildingHandler.GetBuildings)
		buildingRoutes.GET("/:id", buildingHandler.GetBuilding)
		buildingRoutes.POST("", buildingHandler.CreateBuilding)
		buildingRoutes.PUT("/:id", buildingHandler.UpdateBuilding)

		// Building-scoped routes
		unitRepo := unit.NewUnitRepository(config.DB)
		unitService := unit.NewUnitService(unitRepo)
		unitHandler := unit.NewUnitHandler(unitService)

		buildingRoutes.GET("/:id/units", unitHandler.GetUnitsByBuilding)
		buildingRoutes.POST("/:id/units", unitHandler.CreateUnit)
		buildingRoutes.GET("/:id/units/:unitId", unitHandler.GetUnit)
		buildingRoutes.PUT("/:id/units/:unitId", unitHandler.UpdateUnit)

		personRepo := people.NewPersonRepository(config.DB)
		personService := people.NewPersonService(personRepo)
		personHandler := people.NewPersonHandler(personService)

		buildingRoutes.GET("/:id/people", personHandler.GetPeopleByBuilding)
		buildingRoutes.POST("/:id/people", personHandler.CreatePerson)
		buildingRoutes.GET("/:id/people/:personId", personHandler.GetPerson)
		buildingRoutes.PUT("/:id/people/:personId", personHandler.UpdatePerson)

		periodRepo := period.NewPeriodRepository(config.DB)
		periodService := period.NewPeriodService(periodRepo)
		periodHandler := period.NewPeriodHandler(periodService)

		buildingRoutes.GET("/:id/periods", periodHandler.GetPeriodsByBuilding)
		buildingRoutes.POST("/:id/periods", periodHandler.CreatePeriod)
		buildingRoutes.GET("/:id/periods/:periodId", periodHandler.GetPeriod)
		buildingRoutes.PUT("/:id/periods/:periodId", periodHandler.UpdatePeriod)

		accountRepo := accounts.NewAccountRepository(config.DB)
		accountService := accounts.NewAccountService(accountRepo)
		accountHandler := accounts.NewAccountHandler(accountService)

		buildingRoutes.GET("/:id/accounts", accountHandler.GetAccountsByBuilding)
		buildingRoutes.POST("/:id/accounts", accountHandler.CreateAccount)
		buildingRoutes.GET("/:id/accounts/:accountId", accountHandler.GetAccount)
		buildingRoutes.PUT("/:id/accounts/:accountId", accountHandler.UpdateAccount)

		itemRepo := items.NewItemRepository(config.DB)
		itemService := items.NewItemService(itemRepo)
		itemHandler := items.NewItemHandler(itemService)

		buildingRoutes.GET("/:id/items", itemHandler.GetItemsByBuilding)
		buildingRoutes.POST("/:id/items", itemHandler.CreateItem)
		buildingRoutes.GET("/:id/items/:itemId", itemHandler.GetItem)
		buildingRoutes.PUT("/:id/items/:itemId", itemHandler.UpdateItem)

		// Invoice routes (building-scoped)
		buildingRoutes.POST("/:id/invoices/preview", invoiceHandler.PreviewInvoice)
		buildingRoutes.POST("/:id/invoices", invoiceHandler.CreateInvoice)
		buildingRoutes.GET("/:id/invoices", invoiceHandler.GetInvoices)
		// Payments route must come before single invoice route to avoid conflict (more specific route first)
		buildingRoutes.GET("/:id/invoices/:invoiceId/payments", paymentHandler.GetPaymentsByInvoice)
		// Applied credits routes (must come before single invoice route)
		buildingRoutes.GET("/:id/invoices/:invoiceId/available-credits", appliedCreditHandler.GetAvailableCredits)
		buildingRoutes.POST("/:id/invoices/:invoiceId/preview-apply-credit", appliedCreditHandler.PreviewApplyCredit)
		buildingRoutes.POST("/:id/invoices/:invoiceId/apply-credit", appliedCreditHandler.ApplyCreditToInvoice)
		buildingRoutes.GET("/:id/invoices/:invoiceId/applied-credits", appliedCreditHandler.GetAppliedCredits)
		buildingRoutes.DELETE("/:id/invoice-applied-credits/:appliedCreditId", appliedCreditHandler.DeleteAppliedCredit)
		buildingRoutes.PUT("/:id/invoices/:invoiceId", invoiceHandler.UpdateInvoice)
		buildingRoutes.GET("/:id/invoices/:invoiceId", invoiceHandler.GetInvoice)

		// Invoice Payment routes (building-scoped)
		buildingRoutes.POST("/:id/invoice-payments/preview", paymentHandler.PreviewInvoicePayment)
		buildingRoutes.POST("/:id/invoice-payments", paymentHandler.CreateInvoicePayment)
		buildingRoutes.GET("/:id/invoice-payments", paymentHandler.GetInvoicePayments)
		buildingRoutes.GET("/:id/invoice-payments/:paymentId", paymentHandler.GetInvoicePayment)
		buildingRoutes.PUT("/:id/invoice-payments/:paymentId", paymentHandler.UpdateInvoicePayment)

		// Reports routes (building-scoped)
		buildingRoutes.GET("/:id/reports/balance-sheet", reportsHandler.GetBalanceSheet)
		buildingRoutes.GET("/:id/reports/trial-balance", reportsHandler.GetTrialBalance)
		buildingRoutes.GET("/:id/reports/transaction-details-by-account", reportsHandler.GetTransactionDetailsByAccount)
		buildingRoutes.GET("/:id/reports/customer-balance-summary", reportsHandler.GetCustomerBalanceSummary)
		buildingRoutes.GET("/:id/reports/customer-balance-details", reportsHandler.GetCustomerBalanceDetails)

		// Sales Receipt routes (building-scoped)
		buildingRoutes.POST("/:id/sales-receipts/preview", receiptHandler.PreviewSalesReceipt)
		buildingRoutes.POST("/:id/sales-receipts", receiptHandler.CreateSalesReceipt)
		buildingRoutes.GET("/:id/sales-receipts", receiptHandler.GetSalesReceipts)
		buildingRoutes.PUT("/:id/sales-receipts/:receiptId", receiptHandler.UpdateSalesReceipt)
		buildingRoutes.GET("/:id/sales-receipts/:receiptId", receiptHandler.GetSalesReceipt)

		// Check routes (building-scoped)
		buildingRoutes.POST("/:id/checks/preview", checkHandler.PreviewCheck)
		buildingRoutes.POST("/:id/checks", checkHandler.CreateCheck)
		buildingRoutes.GET("/:id/checks", checkHandler.GetChecks)
		buildingRoutes.PUT("/:id/checks/:checkId", checkHandler.UpdateCheck)
		buildingRoutes.GET("/:id/checks/:checkId", checkHandler.GetCheck)

		// Credit Memo routes (building-scoped)
		buildingRoutes.POST("/:id/credit-memos/preview", creditMemoHandler.PreviewCreditMemo)
		buildingRoutes.POST("/:id/credit-memos", creditMemoHandler.CreateCreditMemo)
		buildingRoutes.GET("/:id/credit-memos", creditMemoHandler.GetCreditMemosByBuildingID)
		buildingRoutes.PUT("/:id/credit-memos/:creditMemoId", creditMemoHandler.UpdateCreditMemo)
		buildingRoutes.GET("/:id/credit-memos/:creditMemoId", creditMemoHandler.GetCreditMemoByID)

		// Journal routes (building-scoped)
		buildingRoutes.POST("/:id/journals/preview", journalHandler.PreviewJournal)
		buildingRoutes.POST("/:id/journals", journalHandler.CreateJournal)
		buildingRoutes.GET("/:id/journals", journalHandler.GetJournals)
		buildingRoutes.PUT("/:id/journals/:journalId", journalHandler.UpdateJournal)
		buildingRoutes.GET("/:id/journals/:journalId", journalHandler.GetJournal)
	}

	// Legacy routes (keeping for backward compatibility)
	unitRepo := unit.NewUnitRepository(config.DB)
	unitService := unit.NewUnitService(unitRepo)
	unitHandler := unit.NewUnitHandler(unitService)

	unitRoutes := r.Group("/api/units")
	{
		unitRoutes.GET("", unitHandler.GetUnits)
		unitRoutes.GET("/:id", unitHandler.GetUnit)
		unitRoutes.POST("", unitHandler.CreateUnit)
		unitRoutes.PUT("/:id", unitHandler.UpdateUnit)
	}

	// Legacy routes (keeping for backward compatibility)
	peopleTypeRepo := people_types.NewPeopleTypeRepository(config.DB)
	peopleTypeService := people_types.NewPeopleTypeService(peopleTypeRepo)
	peopleTypeHandler := people_types.NewPeopleTypeHandler(peopleTypeService)

	peopleTypeRoutes := r.Group("/api/people-types")
	{
		peopleTypeRoutes.GET("", peopleTypeHandler.GetPeopleTypes)
		peopleTypeRoutes.GET("/:id", peopleTypeHandler.GetPeopleType)
		peopleTypeRoutes.POST("", peopleTypeHandler.CreatePeopleType)
		peopleTypeRoutes.PUT("/:id", peopleTypeHandler.UpdatePeopleType)
	}

	personRepo := people.NewPersonRepository(config.DB)
	personService := people.NewPersonService(personRepo)
	personHandler := people.NewPersonHandler(personService)

	peopleRoutes := r.Group("/api/people")
	{
		peopleRoutes.GET("", personHandler.GetPeople)
		peopleRoutes.GET("/:id", personHandler.GetPerson)
		peopleRoutes.POST("", personHandler.CreatePerson)
		peopleRoutes.PUT("/:id", personHandler.UpdatePerson)
	}

	periodRepo := period.NewPeriodRepository(config.DB)
	periodService := period.NewPeriodService(periodRepo)
	periodHandler := period.NewPeriodHandler(periodService)

	periodRoutes := r.Group("/api/periods")
	{
		periodRoutes.GET("", periodHandler.GetPeriods)
		periodRoutes.GET("/:id", periodHandler.GetPeriod)
		periodRoutes.POST("", periodHandler.CreatePeriod)
		periodRoutes.PUT("/:id", periodHandler.UpdatePeriod)
	}

	accountTypeRepo := account_types.NewAccountTypeRepository(config.DB)
	accountTypeService := account_types.NewAccountTypeService(accountTypeRepo)
	accountTypeHandler := account_types.NewAccountTypeHandler(accountTypeService)

	accountTypeRoutes := r.Group("/api/account-types")
	{
		accountTypeRoutes.GET("", accountTypeHandler.GetAccountTypes)
		accountTypeRoutes.GET("/:id", accountTypeHandler.GetAccountType)
		accountTypeRoutes.POST("", accountTypeHandler.CreateAccountType)
		accountTypeRoutes.PUT("/:id", accountTypeHandler.UpdateAccountType)
	}

	// Legacy routes (keeping for backward compatibility)
	accountRepo := accounts.NewAccountRepository(config.DB)
	accountService := accounts.NewAccountService(accountRepo)
	accountHandler := accounts.NewAccountHandler(accountService)

	accountRoutes := r.Group("/api/accounts")
	{
		accountRoutes.GET("", accountHandler.GetAccounts)
		accountRoutes.GET("/:id", accountHandler.GetAccount)
		accountRoutes.POST("", accountHandler.CreateAccount)
		accountRoutes.PUT("/:id", accountHandler.UpdateAccount)
	}

	// Legacy routes (keeping for backward compatibility)
	itemRepo := items.NewItemRepository(config.DB)
	itemService := items.NewItemService(itemRepo)
	itemHandler := items.NewItemHandler(itemService)

	itemRoutes := r.Group("/api/items")
	{
		itemRoutes.GET("", itemHandler.GetItems)
		itemRoutes.GET("/:id", itemHandler.GetItem)
		itemRoutes.POST("", itemHandler.CreateItem)
		itemRoutes.PUT("/:id", itemHandler.UpdateItem)
	}

	// Invoice routes (legacy)
	invoiceRoutes := r.Group("/api/invoices")
	{
		invoiceRoutes.POST("/preview", invoiceHandler.PreviewInvoice)
		invoiceRoutes.POST("", invoiceHandler.CreateInvoice)
		// Payments route must come before :id route to avoid conflict
		invoiceRoutes.GET("/:id/payments", paymentHandler.GetPaymentsByInvoice)
		invoiceRoutes.PUT("/:id", invoiceHandler.UpdateInvoice)
		invoiceRoutes.GET("/:id", invoiceHandler.GetInvoice)
	}

	// Sales Receipt routes (legacy)
	receiptRoutes := r.Group("/api/sales-receipts")
	{
		receiptRoutes.POST("/preview", receiptHandler.PreviewSalesReceipt)
		receiptRoutes.POST("", receiptHandler.CreateSalesReceipt)
		receiptRoutes.PUT("/:id", receiptHandler.UpdateSalesReceipt)
		receiptRoutes.GET("/:id", receiptHandler.GetSalesReceipt)
	}

	// Invoice Payment routes (legacy)
	paymentRoutes := r.Group("/api/invoice-payments")
	{
		paymentRoutes.POST("", paymentHandler.CreateInvoicePayment)
		paymentRoutes.GET("/:id", paymentHandler.GetInvoicePayment)
	}
}
