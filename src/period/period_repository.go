package period

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/mysecodgit/go_accounting/src/building"
)

type PeriodRepository interface {
	Create(period Period) (Period, error)
	Update(period Period, id int) (Period, error)
	GetByID(id int) (Period, building.Building, error)
	GetAll() ([]Period, []building.Building, error)
	BuildingIDExists(buildingID int) (bool, error)
	CheckDuplicatePeriod(buildingID int, start string, end string, excludeID int) (bool, error)
}

type periodRepo struct {
	db *sql.DB
}

func NewPeriodRepository(db *sql.DB) PeriodRepository {
	return &periodRepo{db: db}
}

func (r *periodRepo) Create(period Period) (Period, error) {
	// Check for duplicate period before inserting
	exists, err := r.CheckDuplicatePeriod(period.BuildingID, period.Start, period.End, 0)
	if err != nil {
		return period, err
	}
	if exists {
		return period, fmt.Errorf("building cannot have duplicate period with the same start and end date")
	}

	result, err := r.db.Exec("INSERT INTO periods (period_name, `start`, `end`, building_id, is_closed) VALUES (?, ?, ?, ?, ?)",
		period.PeriodName, period.Start, period.End, period.BuildingID, period.IsClosed)

	if err != nil {
		// Check if it's a duplicate key error
		if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "UNIQUE constraint") {
			return period, fmt.Errorf("building cannot have duplicate period with the same start and end date")
		}
		return period, err
	}

	id, _ := result.LastInsertId()
	period.ID = int(id)

	// Fetch the created record to get created_at and updated_at
	err = r.db.QueryRow("SELECT id, period_name, `start`, `end`, building_id, is_closed, created_at, updated_at FROM periods WHERE id = ?", period.ID).
		Scan(&period.ID, &period.PeriodName, &period.Start, &period.End, &period.BuildingID, &period.IsClosed, &period.CreatedAt, &period.UpdatedAt)

	return period, err
}

func (r *periodRepo) Update(period Period, id int) (Period, error) {
	// Check for duplicate period before updating (excluding current period)
	exists, err := r.CheckDuplicatePeriod(period.BuildingID, period.Start, period.End, id)
	if err != nil {
		return period, err
	}
	if exists {
		return period, fmt.Errorf("building cannot have duplicate period with the same start and end date")
	}

	_, err = r.db.Exec("UPDATE periods SET period_name=?, `start`=?, `end`=?, building_id=?, is_closed=?, updated_at=NOW() WHERE id=?",
		period.PeriodName, period.Start, period.End, period.BuildingID, period.IsClosed, id)

	if err != nil {
		// Check if it's a duplicate key error
		if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "UNIQUE constraint") {
			return period, fmt.Errorf("building cannot have duplicate period with the same start and end date")
		}
		return period, err
	}

	period.ID = id

	// Fetch the updated record to get created_at and updated_at
	err = r.db.QueryRow("SELECT id, period_name, `start`, `end`, building_id, is_closed, created_at, updated_at FROM periods WHERE id = ?", id).
		Scan(&period.ID, &period.PeriodName, &period.Start, &period.End, &period.BuildingID, &period.IsClosed, &period.CreatedAt, &period.UpdatedAt)

	return period, err
}

func (r *periodRepo) GetByID(id int) (Period, building.Building, error) {
	var period Period
	var b building.Building
	err := r.db.QueryRow("SELECT p.id, p.period_name, p.`start`, p.`end`, p.building_id, p.is_closed, p.created_at, p.updated_at, "+
		"b.id, b.name, b.created_at, b.updated_at "+
		"FROM periods p "+
		"INNER JOIN buildings b ON p.building_id = b.id "+
		"WHERE p.id = ?", id).
		Scan(&period.ID, &period.PeriodName, &period.Start, &period.End, &period.BuildingID, &period.IsClosed, &period.CreatedAt, &period.UpdatedAt,
			&b.ID, &b.Name, &b.CreatedAt, &b.UpdatedAt)

	if err == sql.ErrNoRows {
		return period, b, fmt.Errorf("id does not exist")
	}

	return period, b, err
}

func (r *periodRepo) GetAll() ([]Period, []building.Building, error) {
	rows, err := r.db.Query("SELECT p.id, p.period_name, p.`start`, p.`end`, p.building_id, p.is_closed, p.created_at, p.updated_at, " +
		"b.id, b.name, b.created_at, b.updated_at " +
		"FROM periods p " +
		"INNER JOIN buildings b ON p.building_id = b.id " +
		"ORDER BY p.created_at DESC")
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	periods := []Period{}
	buildings := []building.Building{}
	for rows.Next() {
		var p Period
		var b building.Building
		err := rows.Scan(&p.ID, &p.PeriodName, &p.Start, &p.End, &p.BuildingID, &p.IsClosed, &p.CreatedAt, &p.UpdatedAt,
			&b.ID, &b.Name, &b.CreatedAt, &b.UpdatedAt)
		if err != nil {
			return nil, nil, err
		}
		periods = append(periods, p)
		buildings = append(buildings, b)
	}
	return periods, buildings, nil
}

func (r *periodRepo) BuildingIDExists(buildingID int) (bool, error) {
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM buildings WHERE id = ?)", buildingID).Scan(&exists)
	return exists, err
}

func (r *periodRepo) CheckDuplicatePeriod(buildingID int, start string, end string, excludeID int) (bool, error) {
	var count int
	var err error

	if excludeID > 0 {
		err = r.db.QueryRow(
			"SELECT COUNT(*) FROM periods WHERE building_id = ? AND `start` = ? AND `end` = ? AND id != ?",
			buildingID, start, end, excludeID).Scan(&count)
	} else {
		err = r.db.QueryRow(
			"SELECT COUNT(*) FROM periods WHERE building_id = ? AND `start` = ? AND `end` = ?",
			buildingID, start, end).Scan(&count)
	}

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
