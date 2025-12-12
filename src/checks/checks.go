package checks

import (
	"time"
)

type Check struct {
	ID               int     `json:"id"`
	TransactionID    int     `json:"transaction_id"`
	CheckDate        string  `json:"check_date"`
	ReferenceNumber  *string `json:"reference_number"`
	PaymentAccountID int     `json:"payment_account_id"`
	BuildingID       int     `json:"building_id"`
	Memo             *string `json:"memo"`
	TotalAmount      float64 `json:"total_amount"`
	CreatedAt        string  `json:"created_at"`
}

func (c *Check) Validate() map[string]string {
	errors := make(map[string]string)

	if c.TransactionID <= 0 {
		errors["transaction_id"] = "Transaction ID must be greater than 0"
	}

	if c.CheckDate == "" {
		errors["check_date"] = "Check date is required"
	} else {
		_, err := time.Parse("2006-01-02", c.CheckDate)
		if err != nil {
			errors["check_date"] = "Check date must be in YYYY-MM-DD format"
		}
	}

	if c.PaymentAccountID <= 0 {
		errors["payment_account_id"] = "Payment account is required"
	}

	if c.BuildingID <= 0 {
		errors["building_id"] = "Building ID must be greater than 0"
	}

	if c.TotalAmount <= 0 {
		errors["total_amount"] = "Total amount must be greater than 0"
	}

	if len(errors) == 0 {
		return nil
	}

	return errors
}
