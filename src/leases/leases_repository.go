package leases

import (
	"database/sql"
	"fmt"
)

type LeaseRepository interface {
	Create(lease Lease) (Lease, error)
	Update(lease Lease) (Lease, error)
	GetByID(id int) (Lease, error)
	GetByBuildingID(buildingID int) ([]Lease, error)
	GetByUnitID(unitID int) ([]Lease, error)
	Delete(id int) error
}

type leaseRepo struct {
	db *sql.DB
}

func NewLeaseRepository(db *sql.DB) LeaseRepository {
	return &leaseRepo{db: db}
}

func (r *leaseRepo) Create(lease Lease) (Lease, error) {
	result, err := r.db.Exec(
		"INSERT INTO leases (people_id, building_id, unit_id, start_date, end_date, rent_amount, deposit_amount, service_amount, lease_terms, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		lease.PeopleID, lease.BuildingID, lease.UnitID, lease.StartDate, lease.EndDate, lease.RentAmount, lease.DepositAmount, lease.ServiceAmount, lease.LeaseTerms, lease.Status,
	)

	if err != nil {
		return lease, err
	}

	id, _ := result.LastInsertId()
	lease.ID = int(id)

	err = r.db.QueryRow(
		"SELECT id, people_id, building_id, unit_id, start_date, end_date, rent_amount, deposit_amount, service_amount, lease_terms, status FROM leases WHERE id = ?",
		lease.ID,
	).Scan(
		&lease.ID, &lease.PeopleID, &lease.BuildingID, &lease.UnitID, &lease.StartDate, &lease.EndDate, &lease.RentAmount, &lease.DepositAmount, &lease.ServiceAmount, &lease.LeaseTerms, &lease.Status,
	)

	return lease, err
}

func (r *leaseRepo) Update(lease Lease) (Lease, error) {
	_, err := r.db.Exec(
		"UPDATE leases SET people_id = ?, building_id = ?, unit_id = ?, start_date = ?, end_date = ?, rent_amount = ?, deposit_amount = ?, service_amount = ?, lease_terms = ?, status = ? WHERE id = ?",
		lease.PeopleID, lease.BuildingID, lease.UnitID, lease.StartDate, lease.EndDate, lease.RentAmount, lease.DepositAmount, lease.ServiceAmount, lease.LeaseTerms, lease.Status, lease.ID,
	)

	if err != nil {
		return lease, err
	}

	err = r.db.QueryRow(
		"SELECT id, people_id, building_id, unit_id, start_date, end_date, rent_amount, deposit_amount, service_amount, lease_terms, status FROM leases WHERE id = ?",
		lease.ID,
	).Scan(
		&lease.ID, &lease.PeopleID, &lease.BuildingID, &lease.UnitID, &lease.StartDate, &lease.EndDate, &lease.RentAmount, &lease.DepositAmount, &lease.ServiceAmount, &lease.LeaseTerms, &lease.Status,
	)

	return lease, err
}

func (r *leaseRepo) GetByID(id int) (Lease, error) {
	var lease Lease
	err := r.db.QueryRow(
		"SELECT id, people_id, building_id, unit_id, start_date, end_date, rent_amount, deposit_amount, service_amount, lease_terms, status FROM leases WHERE id = ?",
		id,
	).Scan(
		&lease.ID, &lease.PeopleID, &lease.BuildingID, &lease.UnitID, &lease.StartDate, &lease.EndDate, &lease.RentAmount, &lease.DepositAmount, &lease.ServiceAmount, &lease.LeaseTerms, &lease.Status,
	)

	if err == sql.ErrNoRows {
		return lease, fmt.Errorf("lease not found")
	}

	return lease, err
}

func (r *leaseRepo) GetByBuildingID(buildingID int) ([]Lease, error) {
	rows, err := r.db.Query(
		"SELECT id, people_id, building_id, unit_id, start_date, end_date, rent_amount, deposit_amount, service_amount, lease_terms, status FROM leases WHERE building_id = ? ORDER BY id DESC",
		buildingID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	leases := []Lease{}
	for rows.Next() {
		var lease Lease
		err := rows.Scan(
			&lease.ID, &lease.PeopleID, &lease.BuildingID, &lease.UnitID, &lease.StartDate, &lease.EndDate, &lease.RentAmount, &lease.DepositAmount, &lease.ServiceAmount, &lease.LeaseTerms, &lease.Status,
		)
		if err != nil {
			return nil, err
		}
		leases = append(leases, lease)
	}

	return leases, nil
}

func (r *leaseRepo) GetByUnitID(unitID int) ([]Lease, error) {
	rows, err := r.db.Query(
		"SELECT id, people_id, building_id, unit_id, start_date, end_date, rent_amount, deposit_amount, service_amount, lease_terms, status FROM leases WHERE unit_id = ? AND status = '1' ORDER BY id DESC",
		unitID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	leases := []Lease{}
	for rows.Next() {
		var lease Lease
		err := rows.Scan(
			&lease.ID, &lease.PeopleID, &lease.BuildingID, &lease.UnitID, &lease.StartDate, &lease.EndDate, &lease.RentAmount, &lease.DepositAmount, &lease.ServiceAmount, &lease.LeaseTerms, &lease.Status,
		)
		if err != nil {
			return nil, err
		}
		leases = append(leases, lease)
	}

	return leases, nil
}

func (r *leaseRepo) Delete(id int) error {
	_, err := r.db.Exec("UPDATE leases SET status = '0' WHERE id = ?", id)
	return err
}
