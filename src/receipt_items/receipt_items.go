package receipt_items

type ReceiptItem struct {
	ID            int      `json:"id"`
	ReceiptID     int      `json:"receipt_id"`
	ItemID        int      `json:"item_id"`
	ItemName      string   `json:"item_name"`
	PreviousValue *float64 `json:"previous_value"`
	CurrentValue  *float64 `json:"current_value"`
	Qty           *float64 `json:"qty"`
	Rate          *string  `json:"rate"`
	Total         float64  `json:"total"`
	Status        string   `json:"status"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}
