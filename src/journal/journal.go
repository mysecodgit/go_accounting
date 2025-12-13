package journal

import (
	"time"
)

type Journal struct {
	ID            int     `json:"id"`
	TransactionID int     `json:"transaction_id"`
	Reference     string  `json:"reference"`
	JournalDate   string  `json:"journal_date"`
	BuildingID    int     `json:"building_id"`
	Memo          *string `json:"memo"`
	TotalAmount   float64 `json:"total_amount"`
	CreatedAt     string  `json:"created_at"`
}

func (j *Journal) Validate() map[string]string {
	errors := make(map[string]string)

	if j.TransactionID <= 0 {
		errors["transaction_id"] = "Transaction ID must be greater than 0"
	}

	if j.JournalDate == "" {
		errors["journal_date"] = "Journal date is required"
	} else {
		_, err := time.Parse("2006-01-02", j.JournalDate)
		if err != nil {
			errors["journal_date"] = "Journal date must be in YYYY-MM-DD format"
		}
	}

	if j.BuildingID <= 0 {
		errors["building_id"] = "Building ID must be greater than 0"
	}

	if j.TotalAmount < 0 {
		errors["total_amount"] = "Total amount cannot be negative"
	}

	if len(errors) == 0 {
		return nil
	}

	return errors
}

