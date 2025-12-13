package sales_receipt

import (
	"strings"
	"time"
)

type SalesReceipt struct {
	ID            int     `json:"id"`
	ReceiptNo     string  `json:"receipt_no"`
	TransactionID int     `json:"transaction_id"`
	ReceiptDate   string  `json:"receipt_date"`
	UnitID        *int    `json:"unit_id"`
	PeopleID      *int    `json:"people_id"`
	UserID        int     `json:"user_id"`
	AccountID     int     `json:"account_id"`
	Amount        float64 `json:"amount"`
	Description   string  `json:"description"`
	CancelReason  *string `json:"cancel_reason"`
	Status        int     `json:"status"`
	BuildingID    int     `json:"building_id"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

func (sr *SalesReceipt) Validate() map[string]string {
	errors := make(map[string]string)

	if strings.TrimSpace(sr.ReceiptNo) == "" {
		errors["receipt_no"] = "Receipt number is required"
	}

	if sr.TransactionID <= 0 {
		errors["transaction_id"] = "Transaction ID must be greater than 0"
	}

	if sr.ReceiptDate == "" {
		errors["receipt_date"] = "Receipt date is required"
	} else {
		_, err := time.Parse("2006-01-02", sr.ReceiptDate)
		if err != nil {
			errors["receipt_date"] = "Receipt date must be in YYYY-MM-DD format"
		}
	}

	if sr.PeopleID == nil {
		errors["people_id"] = "People/Customer is required"
	} else if *sr.PeopleID <= 0 {
		errors["people_id"] = "People ID must be greater than 0"
	}

	if sr.UserID <= 0 {
		errors["user_id"] = "User ID must be greater than 0"
	}

	if sr.AccountID <= 0 {
		errors["account_id"] = "Account ID is required and must be greater than 0"
	}

	if sr.Amount <= 0 {
		errors["amount"] = "Amount must be greater than 0"
	}

	if strings.TrimSpace(sr.Description) == "" {
		errors["description"] = "Description cannot be empty"
	}

	if sr.BuildingID <= 0 {
		errors["building_id"] = "Building ID must be greater than 0"
	}

	if sr.Status != 0 && sr.Status != 1 {
		errors["status"] = "Status must be 0 or 1"
	}

	if len(errors) == 0 {
		return nil
	}

	return errors
}
