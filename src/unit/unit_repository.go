package unit

import (
	"database/sql"
	"fmt"

	"github.com/mysecodgit/go_accounting/src/building"
)

type UnitRepository interface {
	Create(unit Unit) (Unit, error)
	Update(unit Unit, id int) (Unit, error)
	GetByID(id int) (Unit, building.Building, error)
	GetAll() ([]Unit, []building.Building, error)
	GetByBuildingID(buildingID int) ([]Unit, []building.Building, error)
	UnitNameExists(name string, buildingId int, excludeId int) (bool, error)
}

type unitRepo struct {
	db *sql.DB
}

func NewUnitRepository(db *sql.DB) UnitRepository {
	return &unitRepo{db: db}
}

func (r *unitRepo) UnitNameExists(name string, buildingID int, excludeID int) (bool, error) {
	var exists bool
	var err error

	if excludeID > 0 {
		// For update: exclude the current unit ID
		query := `
            SELECT EXISTS(
                SELECT 1 FROM units
                WHERE name = ? AND building_id = ? AND id <> ?
            )
        `
		err = r.db.QueryRow(query, name, buildingID, excludeID).Scan(&exists)
	} else {
		// For create: check if any unit with this name exists in the building
		query := `
            SELECT EXISTS(
                SELECT 1 FROM units
                WHERE name = ? AND building_id = ?
            )
        `
		err = r.db.QueryRow(query, name, buildingID).Scan(&exists)
	}

	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *unitRepo) Create(unit Unit) (Unit, error) {
	result, err := r.db.Exec("INSERT INTO units (name, building_id) VALUES (?, ?)",
		unit.Name, unit.BuildingID)

	if err != nil {
		return unit, err
	}

	id, _ := result.LastInsertId()
	unit.ID = int(id)

	// Fetch the created record to get created_at and updated_at
	err = r.db.QueryRow("SELECT id, name, building_id, created_at, updated_at FROM units WHERE id = ?", unit.ID).
		Scan(&unit.ID, &unit.Name, &unit.BuildingID, &unit.CreatedAt, &unit.UpdatedAt)

	return unit, err
}

func (r *unitRepo) Update(unit Unit, id int) (Unit, error) {
	_, err := r.db.Exec("UPDATE units SET name=?, building_id=?, updated_at=NOW() WHERE id=?",
		unit.Name, unit.BuildingID, id)

	if err != nil {
		return unit, err
	}

	unit.ID = id

	// Fetch the updated record to get created_at and updated_at
	err = r.db.QueryRow("SELECT id, name, building_id, created_at, updated_at FROM units WHERE id = ?", id).
		Scan(&unit.ID, &unit.Name, &unit.BuildingID, &unit.CreatedAt, &unit.UpdatedAt)

	return unit, err
}

func (r *unitRepo) GetByID(id int) (Unit, building.Building, error) {
	var unit Unit
	var b building.Building
	err := r.db.QueryRow(`
		SELECT u.id, u.name, u.building_id, u.created_at, u.updated_at,
		       b.id, b.name, b.created_at, b.updated_at
		FROM units u
		INNER JOIN buildings b ON u.building_id = b.id
		WHERE u.id = ?`, id).
		Scan(&unit.ID, &unit.Name, &unit.BuildingID, &unit.CreatedAt, &unit.UpdatedAt,
			&b.ID, &b.Name, &b.CreatedAt, &b.UpdatedAt)

	if err == sql.ErrNoRows {
		return unit, b, fmt.Errorf("id does not exist")
	}

	return unit, b, err
}

func (r *unitRepo) GetAll() ([]Unit, []building.Building, error) {
	rows, err := r.db.Query(`
		SELECT u.id, u.name, u.building_id, u.created_at, u.updated_at,
		       b.id, b.name, b.created_at, b.updated_at
		FROM units u
		INNER JOIN buildings b ON u.building_id = b.id`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	units := []Unit{}
	buildings := []building.Building{}
	for rows.Next() {
		var u Unit
		var b building.Building
		err := rows.Scan(&u.ID, &u.Name, &u.BuildingID, &u.CreatedAt, &u.UpdatedAt,
			&b.ID, &b.Name, &b.CreatedAt, &b.UpdatedAt)
		if err != nil {
			return nil, nil, err
		}
		units = append(units, u)
		buildings = append(buildings, b)
	}
	return units, buildings, nil
}

func (r *unitRepo) GetByBuildingID(buildingID int) ([]Unit, []building.Building, error) {
	rows, err := r.db.Query(`
		SELECT u.id, u.name, u.building_id, u.created_at, u.updated_at,
		       b.id, b.name, b.created_at, b.updated_at
		FROM units u
		INNER JOIN buildings b ON u.building_id = b.id
		WHERE u.building_id = ?`, buildingID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	units := []Unit{}
	buildings := []building.Building{}
	for rows.Next() {
		var u Unit
		var b building.Building
		err := rows.Scan(&u.ID, &u.Name, &u.BuildingID, &u.CreatedAt, &u.UpdatedAt,
			&b.ID, &b.Name, &b.CreatedAt, &b.UpdatedAt)
		if err != nil {
			return nil, nil, err
		}
		units = append(units, u)
		buildings = append(buildings, b)
	}
	return units, buildings, nil
}
