package readings

import (
	"database/sql"
	"fmt"
)

type ReadingRepository interface {
	Create(reading Reading) (Reading, error)
	CreateWithTx(tx *sql.Tx, reading Reading) (Reading, error)
	Update(reading Reading) (Reading, error)
	GetByID(id int) (Reading, error)
	GetByBuildingID(buildingID int, status *string) ([]Reading, error)
	GetByUnitID(unitID int) ([]Reading, error)
	GetByLeaseID(leaseID int) ([]Reading, error)
	GetLatestByItemAndUnit(itemID, unitID int) (*Reading, error)
	Delete(id int) error
}

type readingRepo struct {
	db *sql.DB
}

func NewReadingRepository(db *sql.DB) ReadingRepository {
	return &readingRepo{db: db}
}

func (r *readingRepo) Create(reading Reading) (Reading, error) {
	result, err := r.db.Exec(
		"INSERT INTO readings (item_id, unit_id, lease_id, reading_month, reading_year, reading_date, previous_value, current_value, unit_price, total_amount, notes, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		reading.ItemID, reading.UnitID, reading.LeaseID, reading.ReadingMonth, reading.ReadingYear, reading.ReadingDate, reading.PreviousValue, reading.CurrentValue, reading.UnitPrice, reading.TotalAmount, reading.Notes, reading.Status,
	)

	if err != nil {
		return reading, err
	}

	id, _ := result.LastInsertId()
	reading.ID = int(id)

	err = r.db.QueryRow(
		"SELECT id, item_id, unit_id, lease_id, reading_month, reading_year, reading_date, previous_value, current_value, unit_price, total_amount, notes, status, created_at, updated_at FROM readings WHERE id = ?",
		reading.ID,
	).Scan(
		&reading.ID, &reading.ItemID, &reading.UnitID, &reading.LeaseID, &reading.ReadingMonth, &reading.ReadingYear, &reading.ReadingDate, &reading.PreviousValue, &reading.CurrentValue, &reading.UnitPrice, &reading.TotalAmount, &reading.Notes, &reading.Status, &reading.CreatedAt, &reading.UpdatedAt,
	)

	return reading, err
}

func (r *readingRepo) CreateWithTx(tx *sql.Tx, reading Reading) (Reading, error) {
	result, err := tx.Exec(
		"INSERT INTO readings (item_id, unit_id, lease_id, reading_month, reading_year, reading_date, previous_value, current_value, unit_price, total_amount, notes, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		reading.ItemID, reading.UnitID, reading.LeaseID, reading.ReadingMonth, reading.ReadingYear, reading.ReadingDate, reading.PreviousValue, reading.CurrentValue, reading.UnitPrice, reading.TotalAmount, reading.Notes, reading.Status,
	)

	if err != nil {
		return reading, err
	}

	id, _ := result.LastInsertId()
	reading.ID = int(id)

	err = tx.QueryRow(
		"SELECT id, item_id, unit_id, lease_id, reading_month, reading_year, reading_date, previous_value, current_value, unit_price, total_amount, notes, status, created_at, updated_at FROM readings WHERE id = ?",
		reading.ID,
	).Scan(
		&reading.ID, &reading.ItemID, &reading.UnitID, &reading.LeaseID, &reading.ReadingMonth, &reading.ReadingYear, &reading.ReadingDate, &reading.PreviousValue, &reading.CurrentValue, &reading.UnitPrice, &reading.TotalAmount, &reading.Notes, &reading.Status, &reading.CreatedAt, &reading.UpdatedAt,
	)

	return reading, err
}

func (r *readingRepo) Update(reading Reading) (Reading, error) {
	_, err := r.db.Exec(
		"UPDATE readings SET item_id = ?, unit_id = ?, lease_id = ?, reading_month = ?, reading_year = ?, reading_date = ?, previous_value = ?, current_value = ?, unit_price = ?, total_amount = ?, notes = ?, status = ? WHERE id = ?",
		reading.ItemID, reading.UnitID, reading.LeaseID, reading.ReadingMonth, reading.ReadingYear, reading.ReadingDate, reading.PreviousValue, reading.CurrentValue, reading.UnitPrice, reading.TotalAmount, reading.Notes, reading.Status, reading.ID,
	)

	if err != nil {
		return reading, err
	}

	err = r.db.QueryRow(
		"SELECT id, item_id, unit_id, lease_id, reading_month, reading_year, reading_date, previous_value, current_value, unit_price, total_amount, notes, status, created_at, updated_at FROM readings WHERE id = ?",
		reading.ID,
	).Scan(
		&reading.ID, &reading.ItemID, &reading.UnitID, &reading.LeaseID, &reading.ReadingMonth, &reading.ReadingYear, &reading.ReadingDate, &reading.PreviousValue, &reading.CurrentValue, &reading.UnitPrice, &reading.TotalAmount, &reading.Notes, &reading.Status, &reading.CreatedAt, &reading.UpdatedAt,
	)

	return reading, err
}

func (r *readingRepo) GetByID(id int) (Reading, error) {
	var reading Reading
	err := r.db.QueryRow(
		"SELECT id, item_id, unit_id, lease_id, reading_month, reading_year, reading_date, previous_value, current_value, unit_price, total_amount, notes, status, created_at, updated_at FROM readings WHERE id = ?",
		id,
	).Scan(
		&reading.ID, &reading.ItemID, &reading.UnitID, &reading.LeaseID, &reading.ReadingMonth, &reading.ReadingYear, &reading.ReadingDate, &reading.PreviousValue, &reading.CurrentValue, &reading.UnitPrice, &reading.TotalAmount, &reading.Notes, &reading.Status, &reading.CreatedAt, &reading.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return reading, fmt.Errorf("reading not found")
	}

	return reading, err
}

func (r *readingRepo) GetByBuildingID(buildingID int, status *string) ([]Reading, error) {
	// Get readings for units in this building
	query := "SELECT r.id, r.item_id, r.unit_id, r.lease_id, r.reading_month, r.reading_year, r.reading_date, r.previous_value, r.current_value, r.unit_price, r.total_amount, r.notes, r.status, r.created_at, r.updated_at FROM readings r INNER JOIN units u ON r.unit_id = u.id WHERE u.building_id = ?"
	args := []interface{}{buildingID}

	if status != nil && *status != "" {
		query += " AND r.status = ?"
		args = append(args, *status)
	}

	query += " ORDER BY r.reading_date DESC, r.id DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	readings := []Reading{}
	for rows.Next() {
		var reading Reading
		err := rows.Scan(
			&reading.ID, &reading.ItemID, &reading.UnitID, &reading.LeaseID, &reading.ReadingMonth, &reading.ReadingYear, &reading.ReadingDate, &reading.PreviousValue, &reading.CurrentValue, &reading.UnitPrice, &reading.TotalAmount, &reading.Notes, &reading.Status, &reading.CreatedAt, &reading.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		readings = append(readings, reading)
	}

	return readings, nil
}

func (r *readingRepo) GetByUnitID(unitID int) ([]Reading, error) {
	rows, err := r.db.Query(
		"SELECT id, item_id, unit_id, lease_id, reading_month, reading_year, reading_date, previous_value, current_value, unit_price, total_amount, notes, status, created_at, updated_at FROM readings WHERE unit_id = ? ORDER BY reading_date DESC, id DESC",
		unitID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	readings := []Reading{}
	for rows.Next() {
		var reading Reading
		err := rows.Scan(
			&reading.ID, &reading.ItemID, &reading.UnitID, &reading.LeaseID, &reading.ReadingMonth, &reading.ReadingYear, &reading.ReadingDate, &reading.PreviousValue, &reading.CurrentValue, &reading.UnitPrice, &reading.TotalAmount, &reading.Notes, &reading.Status, &reading.CreatedAt, &reading.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		readings = append(readings, reading)
	}

	return readings, nil
}

func (r *readingRepo) GetByLeaseID(leaseID int) ([]Reading, error) {
	rows, err := r.db.Query(
		"SELECT id, item_id, unit_id, lease_id, reading_month, reading_year, reading_date, previous_value, current_value, unit_price, total_amount, notes, status, created_at, updated_at FROM readings WHERE lease_id = ? ORDER BY reading_date DESC, id DESC",
		leaseID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	readings := []Reading{}
	for rows.Next() {
		var reading Reading
		err := rows.Scan(
			&reading.ID, &reading.ItemID, &reading.UnitID, &reading.LeaseID, &reading.ReadingMonth, &reading.ReadingYear, &reading.ReadingDate, &reading.PreviousValue, &reading.CurrentValue, &reading.UnitPrice, &reading.TotalAmount, &reading.Notes, &reading.Status, &reading.CreatedAt, &reading.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		readings = append(readings, reading)
	}

	return readings, nil
}

func (r *readingRepo) GetLatestByItemAndUnit(itemID, unitID int) (*Reading, error) {
	var reading Reading
	err := r.db.QueryRow(
		"SELECT id, item_id, unit_id, lease_id, reading_month, reading_year, reading_date, previous_value, current_value, unit_price, total_amount, notes, status, created_at, updated_at FROM readings WHERE item_id = ? AND unit_id = ? AND status = '1' ORDER BY reading_date DESC, id DESC LIMIT 1",
		itemID, unitID,
	).Scan(
		&reading.ID, &reading.ItemID, &reading.UnitID, &reading.LeaseID, &reading.ReadingMonth, &reading.ReadingYear, &reading.ReadingDate, &reading.PreviousValue, &reading.CurrentValue, &reading.UnitPrice, &reading.TotalAmount, &reading.Notes, &reading.Status, &reading.CreatedAt, &reading.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No reading found, return nil
	}

	if err != nil {
		return nil, err
	}

	return &reading, nil
}

func (r *readingRepo) Delete(id int) error {
	_, err := r.db.Exec("UPDATE readings SET status = '0' WHERE id = ?", id)
	return err
}
