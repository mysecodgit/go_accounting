package invoice_applied_discounts

import (
	"time"
)

type InvoiceAppliedDiscount struct {
	ID            int     `json:"id"`
	Reference     string  `json:"reference"`
	InvoiceID     int     `json:"invoice_id"`
	TransactionID int     `json:"transaction_id"`
	ARAccount     int     `json:"ar_account"`
	IncomeAccount int     `json:"income_account"`
	Amount        float64 `json:"amount"`
	Description   string  `json:"description"`
	Date          string  `json:"date"`
	Status        string  `json:"status"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

func (i *InvoiceAppliedDiscount) Validate() map[string]string {
	errors := make(map[string]string)

	if i.InvoiceID <= 0 {
		errors["invoice_id"] = "Invoice ID must be greater than 0"
	}

	if i.TransactionID <= 0 {
		errors["transaction_id"] = "Transaction ID must be greater than 0"
	}

	if i.ARAccount <= 0 {
		errors["ar_account"] = "A/R Account must be greater than 0"
	}

	if i.IncomeAccount <= 0 {
		errors["income_account"] = "Income Account must be greater than 0"
	}

	if i.Amount <= 0 {
		errors["amount"] = "Amount must be greater than 0"
	}

	if i.Description == "" {
		errors["description"] = "Description is required"
	}

	if i.Date == "" {
		errors["date"] = "Date is required"
	} else {
		_, err := time.Parse("2006-01-02", i.Date)
		if err != nil {
			errors["date"] = "Date must be in YYYY-MM-DD format"
		}
	}

	if len(errors) == 0 {
		return nil
	}

	return errors
}

