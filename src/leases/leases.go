package leases

import (
	"time"
)

type Lease struct {
	ID            int     `json:"id"`
	PeopleID      int     `json:"people_id"`
	BuildingID    int     `json:"building_id"`
	UnitID        int     `json:"unit_id"`
	StartDate     string  `json:"start_date"`
	EndDate       *string `json:"end_date"`
	RentAmount    float64 `json:"rent_amount"`
	DepositAmount float64 `json:"deposit_amount"`
	ServiceAmount float64 `json:"service_amount"`
	LeaseTerms    string  `json:"lease_terms"`
	Status        string  `json:"status"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

func (l *Lease) Validate() map[string]string {
	errors := make(map[string]string)

	if l.PeopleID <= 0 {
		errors["people_id"] = "People ID must be greater than 0"
	}

	if l.BuildingID <= 0 {
		errors["building_id"] = "Building ID must be greater than 0"
	}

	if l.UnitID <= 0 {
		errors["unit_id"] = "Unit ID must be greater than 0"
	}

	if l.StartDate == "" {
		errors["start_date"] = "Start date is required"
	} else {
		_, err := time.Parse("2006-01-02", l.StartDate)
		if err != nil {
			errors["start_date"] = "Start date must be in YYYY-MM-DD format"
		}
	}

	if l.EndDate != nil && *l.EndDate != "" {
		_, err := time.Parse("2006-01-02", *l.EndDate)
		if err != nil {
			errors["end_date"] = "End date must be in YYYY-MM-DD format"
		}
	}

	if l.RentAmount < 0 {
		errors["rent_amount"] = "Rent amount cannot be negative"
	}

	if l.DepositAmount < 0 {
		errors["deposit_amount"] = "Deposit amount cannot be negative"
	}

	if l.ServiceAmount < 0 {
		errors["service_amount"] = "Service amount cannot be negative"
	}

	// Lease terms is optional, no validation needed

	if len(errors) == 0 {
		return nil
	}

	return errors
}

