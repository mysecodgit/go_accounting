package credit_memo

import (
	"time"
)

type CreditMemo struct {
	ID               int     `json:"id"`
	TransactionID    int     `json:"transaction_id"`
	Reference        string  `json:"reference"`
	Date             string  `json:"date"`
	UserID           int     `json:"user_id"`
	DepositTo        int     `json:"deposit_to"`
	LiabilityAccount int     `json:"liability_account"`
	PeopleID         int     `json:"people_id"`
	BuildingID       int     `json:"building_id"`
	UnitID           int     `json:"unit_id"`
	Amount           float64 `json:"amount"`
	Description      string  `json:"description"`
	Status           string  `json:"status"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

func (c *CreditMemo) Validate() map[string]string {
	errors := make(map[string]string)

	if c.TransactionID <= 0 {
		errors["transaction_id"] = "Transaction ID must be greater than 0"
	}

	if c.Date == "" {
		errors["date"] = "Date is required"
	} else {
		_, err := time.Parse("2006-01-02", c.Date)
		if err != nil {
			errors["date"] = "Date must be in YYYY-MM-DD format"
		}
	}

	if c.UserID <= 0 {
		errors["user_id"] = "User ID must be greater than 0"
	}

	if c.DepositTo <= 0 {
		errors["deposit_to"] = "Deposit to account is required"
	}

	if c.LiabilityAccount <= 0 {
		errors["liability_account"] = "Liability account is required"
	}

	if c.PeopleID <= 0 {
		errors["people_id"] = "People ID is required"
	}

	if c.BuildingID <= 0 {
		errors["building_id"] = "Building ID must be greater than 0"
	}

	if c.UnitID <= 0 {
		errors["unit_id"] = "Unit ID is required"
	}

	if c.Amount <= 0 {
		errors["amount"] = "Amount must be greater than 0"
	}

	if c.Description == "" {
		errors["description"] = "Description is required"
	}

	if len(errors) == 0 {
		return nil
	}

	return errors
}

