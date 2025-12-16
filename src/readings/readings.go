package readings

import (
	"strings"
	"time"
)

type Reading struct {
	ID            int     `json:"id"`
	ItemID        int     `json:"item_id"`
	UnitID        int     `json:"unit_id"`
	LeaseID       *int    `json:"lease_id"`
	ReadingMonth  *string `json:"reading_month"`
	ReadingYear   *string `json:"reading_year"`
	ReadingDate   string  `json:"reading_date"`
	PreviousValue *float64 `json:"previous_value"`
	CurrentValue  *float64 `json:"current_value"`
	UnitPrice     *float64 `json:"unit_price"`
	TotalAmount   *float64 `json:"total_amount"`
	Notes         *string `json:"notes"`
	Status        string  `json:"status"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

func (r *Reading) Validate() map[string]string {
	errors := make(map[string]string)

	if r.ItemID <= 0 {
		errors["item_id"] = "Item ID must be greater than 0"
	}

	if r.UnitID <= 0 {
		errors["unit_id"] = "Unit ID must be greater than 0"
	}

	if r.ReadingDate == "" {
		errors["reading_date"] = "Reading date is required"
	} else {
		_, err := time.Parse("2006-01-02", r.ReadingDate)
		if err != nil {
			errors["reading_date"] = "Reading date must be in YYYY-MM-DD format"
		}
	}

	if r.ReadingMonth != nil && *r.ReadingMonth != "" {
		month := strings.TrimSpace(*r.ReadingMonth)
		if len(month) > 10 {
			errors["reading_month"] = "Reading month must be 10 characters or less"
		}
	}

	if r.ReadingYear != nil && *r.ReadingYear != "" {
		year := strings.TrimSpace(*r.ReadingYear)
		if len(year) > 5 {
			errors["reading_year"] = "Reading year must be 5 characters or less"
		}
	}

	if r.PreviousValue != nil && *r.PreviousValue < 0 {
		errors["previous_value"] = "Previous value cannot be negative"
	}

	if r.CurrentValue != nil && *r.CurrentValue < 0 {
		errors["current_value"] = "Current value cannot be negative"
	}

	if r.UnitPrice != nil && *r.UnitPrice < 0 {
		errors["unit_price"] = "Unit price cannot be negative"
	}

	if r.TotalAmount != nil && *r.TotalAmount < 0 {
		errors["total_amount"] = "Total amount cannot be negative"
	}

	if len(errors) == 0 {
		return nil
	}

	return errors
}

