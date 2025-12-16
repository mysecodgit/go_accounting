package leases

import "github.com/mysecodgit/go_accounting/src/people"

type CreateLeaseRequest struct {
	PeopleID      int     `json:"people_id"`
	BuildingID    int     `json:"building_id"`
	UnitID        int     `json:"unit_id"`
	StartDate     string  `json:"start_date"`
	EndDate       *string `json:"end_date"`
	RentAmount    float64 `json:"rent_amount"`
	DepositAmount float64 `json:"deposit_amount"`
	ServiceAmount float64 `json:"service_amount"`
	LeaseTerms    string  `json:"lease_terms"`
	Status        string  `json:"status"`
}

type UpdateLeaseRequest struct {
	ID            int     `json:"id"`
	PeopleID      int     `json:"people_id"`
	BuildingID    int     `json:"building_id"`
	UnitID        int     `json:"unit_id"`
	StartDate     string  `json:"start_date"`
	EndDate       *string `json:"end_date"`
	RentAmount    float64 `json:"rent_amount"`
	DepositAmount float64 `json:"deposit_amount"`
	ServiceAmount float64 `json:"service_amount"`
	LeaseTerms    string  `json:"lease_terms"`
	Status        string  `json:"status"`
}

type LeaseResponse struct {
	Lease      Lease       `json:"lease"`
	LeaseFiles []LeaseFile `json:"lease_files"`
}

type LeaseListItem struct {
	Lease  Lease          `json:"lease"`
	People *people.Person `json:"people,omitempty"`
}
