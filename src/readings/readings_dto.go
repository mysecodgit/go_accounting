package readings

import (
	"github.com/mysecodgit/go_accounting/src/items"
	"github.com/mysecodgit/go_accounting/src/leases"
	"github.com/mysecodgit/go_accounting/src/people"
	"github.com/mysecodgit/go_accounting/src/unit"
)

type LeaseWithPeople struct {
	Lease  leases.Lease   `json:"lease"`
	People *people.Person `json:"people,omitempty"`
}

type CreateReadingRequest struct {
	ItemID        int      `json:"item_id"`
	UnitID        int      `json:"unit_id"`
	LeaseID       *int     `json:"lease_id"`
	ReadingMonth  *string  `json:"reading_month"`
	ReadingYear   *string  `json:"reading_year"`
	ReadingDate   string   `json:"reading_date"`
	PreviousValue *float64 `json:"previous_value"`
	CurrentValue  *float64 `json:"current_value"`
	UnitPrice     *float64 `json:"unit_price"`
	TotalAmount   *float64 `json:"total_amount"`
	Notes         *string  `json:"notes"`
	Status        string   `json:"status"`
}

type UpdateReadingRequest struct {
	ID            int      `json:"id"`
	ItemID        int      `json:"item_id"`
	UnitID        int      `json:"unit_id"`
	LeaseID       *int     `json:"lease_id"`
	ReadingMonth  *string  `json:"reading_month"`
	ReadingYear   *string  `json:"reading_year"`
	ReadingDate   string   `json:"reading_date"`
	PreviousValue *float64 `json:"previous_value"`
	CurrentValue  *float64 `json:"current_value"`
	UnitPrice     *float64 `json:"unit_price"`
	TotalAmount   *float64 `json:"total_amount"`
	Notes         *string  `json:"notes"`
	Status        string   `json:"status"`
}

type ReadingResponse struct {
	Reading Reading       `json:"reading"`
	Item    items.Item    `json:"item"`
	Unit    unit.Unit     `json:"unit"`
	Lease   *leases.Lease `json:"lease,omitempty"`
}

type ReadingListItem struct {
	Reading Reading          `json:"reading"`
	Item    items.Item       `json:"item"`
	Unit    unit.Unit        `json:"unit"`
	Lease   *LeaseWithPeople `json:"lease,omitempty"`
}

type BulkImportReadingRequest struct {
	ItemID        int      `json:"item_id"`
	UnitID        int      `json:"unit_id"`
	LeaseID       *int     `json:"lease_id"`
	ReadingMonth  *string  `json:"reading_month"`
	ReadingYear   *string  `json:"reading_year"`
	ReadingDate   string   `json:"reading_date"`
	PreviousValue *float64 `json:"previous_value"`
	CurrentValue  *float64 `json:"current_value"`
	UnitPrice     *float64 `json:"unit_price"`
	TotalAmount   *float64 `json:"total_amount"`
	Notes         *string  `json:"notes"`
	Status        string   `json:"status"`
}

type BulkImportReadingsRequest struct {
	Readings []BulkImportReadingRequest `json:"readings"`
}

type BulkImportReadingsResponse struct {
	SuccessCount int      `json:"success_count"`
	FailedCount  int      `json:"failed_count"`
	Errors       []string `json:"errors,omitempty"`
}
