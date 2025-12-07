package accounts

import "strings"

type Account struct {
	ID            int    `json:"id"`
	AccountNumber int    `json:"account_number"`
	AccountName   string `json:"account_name"`
	AccountType   int    `json:"account_type"`
	BuildingID    int    `json:"building_id"`
	IsDefault     int    `json:"isDefault"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

func (a *Account) Validate() map[string]string {
	errors := make(map[string]string)

	if a.AccountNumber <= 0 {
		errors["account_number"] = "Account number must be a valid positive number"
	}

	if strings.TrimSpace(a.AccountName) == "" {
		errors["account_name"] = "Account name cannot be empty"
	}

	if a.AccountType <= 0 {
		errors["account_type"] = "Account type ID must be greater than 0"
	}

	if a.BuildingID <= 0 {
		errors["building_id"] = "Building ID must be greater than 0"
	}

	if a.IsDefault != 0 && a.IsDefault != 1 {
		errors["isDefault"] = "IsDefault must be 0 or 1"
	}

	if len(errors) == 0 {
		return nil
	}

	return errors
}
