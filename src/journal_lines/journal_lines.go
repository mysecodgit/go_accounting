package journal_lines

type JournalLine struct {
	ID          int      `json:"id"`
	JournalID   int      `json:"journal_id"`
	AccountID   int      `json:"account_id"`
	UnitID      *int     `json:"unit_id"`
	PeopleID    *int     `json:"people_id"`
	Description *string  `json:"description"`
	Debit       *float64 `json:"debit"`
	Credit      *float64 `json:"credit"`
}

func (j *JournalLine) Validate() map[string]string {
	errors := make(map[string]string)

	if j.JournalID <= 0 {
		errors["journal_id"] = "Journal ID must be greater than 0"
	}

	if j.AccountID <= 0 {
		errors["account_id"] = "Account is required"
	}

	if j.Debit == nil && j.Credit == nil {
		errors["amount"] = "Either debit or credit must be specified"
	}

	if j.Debit != nil && j.Credit != nil && *j.Debit > 0 && *j.Credit > 0 {
		errors["amount"] = "Cannot have both debit and credit"
	}

	if j.Debit != nil && *j.Debit < 0 {
		errors["debit"] = "Debit cannot be negative"
	}

	if j.Credit != nil && *j.Credit < 0 {
		errors["credit"] = "Credit cannot be negative"
	}

	if len(errors) == 0 {
		return nil
	}

	return errors
}

