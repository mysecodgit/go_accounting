package building

import (
	"database/sql"
	"fmt"
)

type BuildingRepository interface {
	Create(building Building) (Building, error)
	Update(building Building, id int) (Building, error)
	GetByID(id int) (Building, error)
	GetAll() ([]Building, error)
}

type buildingRepo struct {
	db *sql.DB
}

func NewBuildingRepository(db *sql.DB) BuildingRepository {
	return &buildingRepo{db: db}
}

func (r *buildingRepo) Create(building Building) (Building, error) {
	result, err := r.db.Exec("INSERT INTO buildings (name) VALUES (?)",
		building.Name)

	if err != nil {
		return building, err
	}

	id, _ := result.LastInsertId()
	building.ID = int(id)

	// Fetch the created record to get created_at and updated_at
	err = r.db.QueryRow("SELECT id, name, created_at, updated_at FROM buildings WHERE id = ?", building.ID).
		Scan(&building.ID, &building.Name, &building.CreatedAt, &building.UpdatedAt)

	return building, err
}

func (r *buildingRepo) Update(building Building, id int) (Building, error) {
	_, err := r.db.Exec("UPDATE buildings SET name=?, updated_at=NOW() WHERE id=?",
		building.Name, id)

	if err != nil {
		return building, err
	}

	building.ID = id

	// Fetch the updated record to get created_at and updated_at
	err = r.db.QueryRow("SELECT id, name, created_at, updated_at FROM buildings WHERE id = ?", id).
		Scan(&building.ID, &building.Name, &building.CreatedAt, &building.UpdatedAt)

	return building, err
}

func (r *buildingRepo) GetByID(id int) (Building, error) {
	var building Building
	err := r.db.QueryRow("SELECT id, name, created_at, updated_at FROM buildings WHERE id = ?", id).
		Scan(&building.ID, &building.Name, &building.CreatedAt, &building.UpdatedAt)
	
	if err == sql.ErrNoRows {
		return building, fmt.Errorf("id does not exist")
	}
	
	return building, err
}

func (r *buildingRepo) GetAll() ([]Building, error) {
	rows, err := r.db.Query("SELECT id, name, created_at, updated_at FROM buildings")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	buildings := []Building{}
	for rows.Next() {
		var b Building
		err := rows.Scan(&b.ID, &b.Name, &b.CreatedAt, &b.UpdatedAt)
		if err != nil {
			return nil, err
		}
		buildings = append(buildings, b)
	}
	return buildings, nil
}
