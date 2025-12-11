package invoice_items

type InvoiceItem struct {
	ID            int     `json:"id"`
	InvoiceID     int     `json:"invoice_id"`
	ItemID        int     `json:"item_id"`
	ItemName      string  `json:"item_name"`
	PreviousValue *float64 `json:"previous_value"`
	CurrentValue  *float64 `json:"current_value"`
	Qty           *float64 `json:"qty"`
	Rate          *string `json:"rate"`
	Total         float64 `json:"total"`
	Status        int     `json:"status"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

func (ii *InvoiceItem) Validate() map[string]string {
	errors := make(map[string]string)

	if ii.InvoiceID <= 0 {
		errors["invoice_id"] = "Invoice ID must be greater than 0"
	}

	if ii.ItemID <= 0 {
		errors["item_id"] = "Item ID must be greater than 0"
	}

	if ii.ItemName == "" {
		errors["item_name"] = "Item name cannot be empty"
	}

	if ii.Total <= 0 {
		errors["total"] = "Total must be greater than 0"
	}

	if ii.Status != 0 && ii.Status != 1 {
		errors["status"] = "Status must be 0 or 1"
	}

	if len(errors) == 0 {
		return nil
	}

	return errors
}

