package invoice_payments

import (
	"time"
)

type InvoicePayment struct {
	ID            int     `json:"id"`
	TransactionID int     `json:"transaction_id"`
	Date          string  `json:"date"`
	InvoiceID     int     `json:"invoice_id"`
	UserID        int     `json:"user_id"`
	AccountID     int     `json:"account_id"`
	Amount        float64 `json:"amount"`
	Status        int     `json:"status"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

func (ip *InvoicePayment) Validate() map[string]string {
	errors := make(map[string]string)

	if ip.TransactionID <= 0 {
		errors["transaction_id"] = "Transaction ID must be greater than 0"
	}

	if ip.Date == "" {
		errors["date"] = "Date is required"
	} else {
		_, err := time.Parse("2006-01-02", ip.Date)
		if err != nil {
			errors["date"] = "Date must be in YYYY-MM-DD format"
		}
	}

	if ip.InvoiceID <= 0 {
		errors["invoice_id"] = "Invoice ID must be greater than 0"
	}

	if ip.UserID <= 0 {
		errors["user_id"] = "User ID must be greater than 0"
	}

	if ip.AccountID <= 0 {
		errors["account_id"] = "Account ID is required and must be greater than 0"
	}

	if ip.Amount == 0 {
		errors["amount"] = "Amount cannot be zero"
	}

	if ip.Status != 0 && ip.Status != 1 {
		errors["status"] = "Status must be 0 or 1"
	}

	if len(errors) == 0 {
		return nil
	}

	return errors
}

