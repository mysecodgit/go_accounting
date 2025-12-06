package people_types

import (
	"database/sql"
	"fmt"
	"github.com/mysecodgit/go_accounting/src/building"
)

type PeopleTypeRepository interface {
	Create(peopleType PeopleType) (PeopleType, error)
	UpdateTitle(title string, id int) (PeopleType, error)
	GetByID(id int) (PeopleType, building.Building, error)
	GetAll() ([]PeopleType, []building.Building, error)
	TitleExistsInBuilding(title string, buildingID int) (bool, error)
	TitleExistsInBuildingExcludingID(title string, buildingID int, excludeID int) (bool, error)
}

type peopleTypeRepo struct {
	db *sql.DB
}

func NewPeopleTypeRepository(db *sql.DB) PeopleTypeRepository {
	return &peopleTypeRepo{db: db}
}

func (r *peopleTypeRepo) Create(peopleType PeopleType) (PeopleType, error) {
	result, err := r.db.Exec("INSERT INTO people_types (title, building_id) VALUES (?, ?)",
		peopleType.Title, peopleType.BuildingID)

	if err != nil {
		return peopleType, err
	}

	id, _ := result.LastInsertId()
	peopleType.ID = int(id)

	return peopleType, err
}

func (r *peopleTypeRepo) UpdateTitle(title string, id int) (PeopleType, error) {
	var peopleType PeopleType
	_, err := r.db.Exec("UPDATE people_types SET title=? WHERE id=?",
		title, id)

	if err != nil {
		return peopleType, err
	}

	// Fetch the updated record
	err = r.db.QueryRow("SELECT id, title, building_id FROM people_types WHERE id = ?", id).
		Scan(&peopleType.ID, &peopleType.Title, &peopleType.BuildingID)

	return peopleType, err
}

func (r *peopleTypeRepo) GetByID(id int) (PeopleType, building.Building, error) {
	var peopleType PeopleType
	var b building.Building
	err := r.db.QueryRow(`
		SELECT pt.id, pt.title, pt.building_id,
		       b.id, b.name, b.created_at, b.updated_at
		FROM people_types pt
		INNER JOIN buildings b ON pt.building_id = b.id
		WHERE pt.id = ?`, id).
		Scan(&peopleType.ID, &peopleType.Title, &peopleType.BuildingID,
			&b.ID, &b.Name, &b.CreatedAt, &b.UpdatedAt)
	
	if err == sql.ErrNoRows {
		return peopleType, b, fmt.Errorf("id does not exist")
	}
	
	return peopleType, b, err
}

func (r *peopleTypeRepo) GetAll() ([]PeopleType, []building.Building, error) {
	rows, err := r.db.Query(`
		SELECT pt.id, pt.title, pt.building_id,
		       b.id, b.name, b.created_at, b.updated_at
		FROM people_types pt
		INNER JOIN buildings b ON pt.building_id = b.id`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	peopleTypes := []PeopleType{}
	buildings := []building.Building{}
	for rows.Next() {
		var p PeopleType
		var b building.Building
		err := rows.Scan(&p.ID, &p.Title, &p.BuildingID,
			&b.ID, &b.Name, &b.CreatedAt, &b.UpdatedAt)
		if err != nil {
			return nil, nil, err
		}
		peopleTypes = append(peopleTypes, p)
		buildings = append(buildings, b)
	}
	return peopleTypes, buildings, nil
}

func (r *peopleTypeRepo) TitleExistsInBuilding(title string, buildingID int) (bool, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM people_types WHERE title = ? AND building_id = ?", title, buildingID).
		Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *peopleTypeRepo) TitleExistsInBuildingExcludingID(title string, buildingID int, excludeID int) (bool, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM people_types WHERE title = ? AND building_id = ? AND id != ?", title, buildingID, excludeID).
		Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
