package items

import (
	"strings"
	"time"
)

type Item struct {
	ID              int     `json:"id"`
	Name            string  `json:"name"`
	Type            string  `json:"type"`
	Description     string  `json:"description"`
	AssetAccount    *int    `json:"asset_account"`
	IncomeAccount   *int    `json:"income_account"`
	COGSAccount     *int    `json:"cogs_account"`
	ExpenseAccount  *int    `json:"expense_account"`
	OnHand          float64 `json:"on_hand"`
	AvgCost         float64 `json:"avg_cost"`
	Date            string  `json:"date"`
	BuildingID      int     `json:"building_id"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

func (i *Item) Validate() map[string]string {
	errors := make(map[string]string)

	if strings.TrimSpace(i.Name) == "" {
		errors["name"] = "Name cannot be empty"
	}

	validTypes := []string{"inventory", "non inventory", "service", "discount", "payment"}
	typeValid := false
	for _, validType := range validTypes {
		if i.Type == validType {
			typeValid = true
			break
		}
	}
	if !typeValid {
		errors["type"] = "Type must be one of: inventory, non inventory, service, discount, payment"
	}

	if i.BuildingID <= 0 {
		errors["building_id"] = "Building ID must be greater than 0"
	}

	// Validate date format if provided
	if i.Date != "" {
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

