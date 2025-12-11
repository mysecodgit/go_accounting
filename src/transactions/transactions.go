package transactions

import (
	"strings"
	"time"
)

type Transaction struct {
	ID              int    `json:"id"`
	Type            string `json:"type"`
	TransactionDate string `json:"transaction_date"`
	Memo            string `json:"memo"`
	Status          int    `json:"status"`
	BuildingID      int    `json:"building_id"`
	UserID          int    `json:"user_id"`
	UnitID          *int   `json:"unit_id"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

func (t *Transaction) Validate() map[string]string {
	errors := make(map[string]string)

	validTypes := []string{"invoice", "payment", "check", "deposit", "bill", "credit memo", "sales receipt", "journal", "bill credit", "bill payment"}
	typeValid := false
	for _, validType := range validTypes {
		if t.Type == validType {
			typeValid = true
			break
		}
	}
	if !typeValid {
		errors["type"] = "Invalid transaction type"
	}

	if t.TransactionDate == "" {
		errors["transaction_date"] = "Transaction date is required"
	} else {
		_, err := time.Parse("2006-01-02", t.TransactionDate)
		if err != nil {
			errors["transaction_date"] = "Transaction date must be in YYYY-MM-DD format"
		}
	}

	if strings.TrimSpace(t.Memo) == "" {
		errors["memo"] = "Memo cannot be empty"
	}

	if t.BuildingID <= 0 {
		errors["building_id"] = "Building ID must be greater than 0"
	}

	if t.UserID <= 0 {
		errors["user_id"] = "User ID must be greater than 0"
	}

	if t.Status != 0 && t.Status != 1 {
		errors["status"] = "Status must be 0 or 1"
	}

	if len(errors) == 0 {
		return nil
	}

	return errors
}

