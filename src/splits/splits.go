package splits

type Split struct {
	ID            int     `json:"id"`
	TransactionID int     `json:"transaction_id"`
	AccountID     int     `json:"account_id"`
	PeopleID      *int    `json:"people_id"`
	UnitID        *int    `json:"unit_id"`
	Debit         *float64 `json:"debit"`
	Credit        *float64 `json:"credit"`
	Status        string  `json:"status"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

func (s *Split) Validate() map[string]string {
	errors := make(map[string]string)

	if s.TransactionID <= 0 {
		errors["transaction_id"] = "Transaction ID must be greater than 0"
	}

	if s.AccountID <= 0 {
		errors["account_id"] = "Account ID must be greater than 0"
	}

	if s.Debit == nil && s.Credit == nil {
		errors["amount"] = "Either debit or credit must be specified"
	}

	if s.Debit != nil && s.Credit != nil && *s.Debit > 0 && *s.Credit > 0 {
		errors["amount"] = "Cannot have both debit and credit"
	}

	if s.Status == "" {
		errors["status"] = "Status is required"
	}

	if len(errors) == 0 {
		return nil
	}

	return errors
}

