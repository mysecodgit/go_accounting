package period

import (
	"strings"
	"time"
)

type Period struct {
	ID         int    `json:"id"`
	PeriodName string `json:"period_name"`
	Start      string `json:"start"`
	End        string `json:"end"`
	BuildingID int    `json:"building_id"`
	IsClosed   int    `json:"is_closed"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

func (p *Period) Validate() map[string]string {
	errors := make(map[string]string)

	if strings.TrimSpace(p.PeriodName) == "" {
		errors["period_name"] = "Period name cannot be empty"
	}

	if strings.TrimSpace(p.Start) == "" {
		errors["start"] = "Start date cannot be empty"
	} else {
		_, err := time.Parse("2006-01-02", p.Start)
		if err != nil {
			errors["start"] = "Start date must be in YYYY-MM-DD format"
		}
	}

	if strings.TrimSpace(p.End) == "" {
		errors["end"] = "End date cannot be empty"
	} else {
		_, err := time.Parse("2006-01-02", p.End)
		if err != nil {
			errors["end"] = "End date must be in YYYY-MM-DD format"
		}
	}

	// Validate that end date is after start date
	if p.Start != "" && p.End != "" {
		startDate, err1 := time.Parse("2006-01-02", p.Start)
		endDate, err2 := time.Parse("2006-01-02", p.End)
		if err1 == nil && err2 == nil && !endDate.After(startDate) && !endDate.Equal(startDate) {
			errors["end"] = "End date must be after or equal to start date"
		}
	}

	if p.BuildingID <= 0 {
		errors["building_id"] = "Building ID must be greater than 0"
	}

	if p.IsClosed != 0 && p.IsClosed != 1 {
		errors["is_closed"] = "Is closed must be 0 or 1"
	}

	if len(errors) == 0 {
		return nil
	}

	return errors
}

