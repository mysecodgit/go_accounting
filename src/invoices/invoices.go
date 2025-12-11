package invoices

import (
	"strings"
	"time"
)

type Invoice struct {
	ID            int     `json:"id"`
	InvoiceNo     int     `json:"invoice_no"`
	TransactionID int     `json:"transaction_id"`
	SalesDate     string  `json:"sales_date"`
	DueDate       string  `json:"due_date"`
	ARAccountID   *int    `json:"ar_account_id"`
	UnitID        *int    `json:"unit_id"`
	PeopleID      *int    `json:"people_id"`
	UserID        int     `json:"user_id"`
	Amount        float64 `json:"amount"`
	Description   string  `json:"description"`
	Reference     string  `json:"refrence"`
	CancelReason  *string `json:"cancel_reason"`
	Status        int     `json:"status"`
	BuildingID    int     `json:"building_id"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

func (i *Invoice) Validate() map[string]string {
	errors := make(map[string]string)

	if i.InvoiceNo <= 0 {
		errors["invoice_no"] = "Invoice number must be greater than 0"
	}

	if i.TransactionID <= 0 {
		errors["transaction_id"] = "Transaction ID must be greater than 0"
	}

	if i.SalesDate == "" {
		errors["sales_date"] = "Sales date is required"
	} else {
		_, err := time.Parse("2006-01-02", i.SalesDate)
		if err != nil {
			errors["sales_date"] = "Sales date must be in YYYY-MM-DD format"
		}
	}

	if i.DueDate == "" {
		errors["due_date"] = "Due date is required"
	} else {
		_, err := time.Parse("2006-01-02", i.DueDate)
		if err != nil {
			errors["due_date"] = "Due date must be in YYYY-MM-DD format"
		}
	}

	if i.UnitID == nil {
		errors["unit_id"] = "Unit is required"
	} else if *i.UnitID <= 0 {
		errors["unit_id"] = "Unit ID must be greater than 0"
	}

	if i.PeopleID == nil {
		errors["people_id"] = "People/Customer is required"
	} else if *i.PeopleID <= 0 {
		errors["people_id"] = "People ID must be greater than 0"
	}

	if i.UserID <= 0 {
		errors["user_id"] = "User ID must be greater than 0"
	}

	if i.Amount <= 0 {
		errors["amount"] = "Amount must be greater than 0"
	}

	if strings.TrimSpace(i.Description) == "" {
		errors["description"] = "Description cannot be empty"
	}

	if strings.TrimSpace(i.Reference) == "" {
		errors["refrence"] = "Reference cannot be empty"
	}

	if i.BuildingID <= 0 {
		errors["building_id"] = "Building ID must be greater than 0"
	}

	if i.Status != 0 && i.Status != 1 {
		errors["status"] = "Status must be 0 or 1"
	}

	if len(errors) == 0 {
		return nil
	}

	return errors
}
