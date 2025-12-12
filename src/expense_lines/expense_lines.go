package expense_lines

type ExpenseLine struct {
	ID          int     `json:"id"`
	CheckID     int     `json:"check_id"`
	AccountID   int     `json:"account_id"`
	UnitID      *int    `json:"unit_id"`
	PeopleID    *int    `json:"people_id"`
	Description *string `json:"description"`
	Amount      float64 `json:"amount"`
}

func (e *ExpenseLine) Validate() map[string]string {
	errors := make(map[string]string)

	if e.CheckID <= 0 {
		errors["check_id"] = "Check ID must be greater than 0"
	}

	if e.AccountID <= 0 {
		errors["account_id"] = "Account is required"
	}

	if e.Amount <= 0 {
		errors["amount"] = "Amount must be greater than 0"
	}

	if len(errors) == 0 {
		return nil
	}

	return errors
}
