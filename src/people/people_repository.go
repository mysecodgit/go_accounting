package people

import (
	"database/sql"
	"fmt"

	"github.com/mysecodgit/go_accounting/src/building"
	"github.com/mysecodgit/go_accounting/src/people_types"
)

type PersonRepository interface {
	Create(person Person) (Person, error)
	UpdateNameAndPhone(name string, phone string, typeID int, id int) (Person, error)
	GetByID(id int) (Person, people_types.PeopleType, building.Building, error)
	GetAll() ([]Person, []people_types.PeopleType, []building.Building, error)
	GetByBuildingID(buildingID int) ([]Person, []people_types.PeopleType, []building.Building, error)
	TypeIDExists(typeID int) (bool, int, error)
	BuildingIDExists(buildingID int) (bool, error)
	PersonNameExists(name string, buildingID int, excludeID int) (bool, error)
}

type personRepo struct {
	db *sql.DB
}

func NewPersonRepository(db *sql.DB) PersonRepository {
	return &personRepo{db: db}
}

func (r *personRepo) Create(person Person) (Person, error) {
	result, err := r.db.Exec("INSERT INTO people (name, phone, type_id, building_id) VALUES (?, ?, ?, ?)",
		person.Name, person.Phone, person.TypeID, person.BuildingID)

	if err != nil {
		return person, err
	}

	id, _ := result.LastInsertId()
	person.ID = int(id)

	// Fetch the created record to get created_at and updated_at
	err = r.db.QueryRow("SELECT id, name, phone, type_id, building_id, created_at, updated_at FROM people WHERE id = ?", person.ID).
		Scan(&person.ID, &person.Name, &person.Phone, &person.TypeID, &person.BuildingID, &person.CreatedAt, &person.UpdatedAt)

	return person, err
}

func (r *personRepo) UpdateNameAndPhone(name string, phone string, typeID int, id int) (Person, error) {
	var person Person
	_, err := r.db.Exec("UPDATE people SET name=?, phone=?, type_id=?, updated_at=NOW() WHERE id=?",
		name, phone, typeID, id)

	if err != nil {
		return person, err
	}

	// Fetch the updated record to get all fields
	err = r.db.QueryRow("SELECT id, name, phone, type_id, building_id, created_at, updated_at FROM people WHERE id = ?", id).
		Scan(&person.ID, &person.Name, &person.Phone, &person.TypeID, &person.BuildingID, &person.CreatedAt, &person.UpdatedAt)

	return person, err
}

func (r *personRepo) GetByID(id int) (Person, people_types.PeopleType, building.Building, error) {
	var person Person
	var pt people_types.PeopleType
	var b building.Building
	err := r.db.QueryRow(`
		SELECT p.id, p.name, p.phone, p.type_id, p.building_id, p.created_at, p.updated_at,
		       pt.id, pt.title,
		       b.id, b.name, b.created_at, b.updated_at
		FROM people p
		INNER JOIN people_types pt ON p.type_id = pt.id
		INNER JOIN buildings b ON p.building_id = b.id
		WHERE p.id = ?`, id).
		Scan(&person.ID, &person.Name, &person.Phone, &person.TypeID, &person.BuildingID, &person.CreatedAt, &person.UpdatedAt,
			&pt.ID, &pt.Title,
			&b.ID, &b.Name, &b.CreatedAt, &b.UpdatedAt)

	if err == sql.ErrNoRows {
		return person, pt, b, fmt.Errorf("id does not exist")
	}

	return person, pt, b, err
}

func (r *personRepo) GetAll() ([]Person, []people_types.PeopleType, []building.Building, error) {
	rows, err := r.db.Query(`
		SELECT p.id, p.name, p.phone, p.type_id, p.building_id, p.created_at, p.updated_at,
		       pt.id, pt.title,
		       b.id, b.name, b.created_at, b.updated_at
		FROM people p
		INNER JOIN people_types pt ON p.type_id = pt.id
		INNER JOIN buildings b ON p.building_id = b.id`)
	if err != nil {
		return nil, nil, nil, err
	}
	defer rows.Close()

	people := []Person{}
	peopleTypes := []people_types.PeopleType{}
	buildings := []building.Building{}
	for rows.Next() {
		var p Person
		var pt people_types.PeopleType
		var b building.Building
		err := rows.Scan(&p.ID, &p.Name, &p.Phone, &p.TypeID, &p.BuildingID, &p.CreatedAt, &p.UpdatedAt,
			&pt.ID, &pt.Title,
			&b.ID, &b.Name, &b.CreatedAt, &b.UpdatedAt)
		if err != nil {
			return nil, nil, nil, err
		}
		people = append(people, p)
		peopleTypes = append(peopleTypes, pt)
		buildings = append(buildings, b)
	}
	return people, peopleTypes, buildings, nil
}

func (r *personRepo) GetByBuildingID(buildingID int) ([]Person, []people_types.PeopleType, []building.Building, error) {
	rows, err := r.db.Query(`
		SELECT p.id, p.name, p.phone, p.type_id, p.building_id, p.created_at, p.updated_at,
		       pt.id, pt.title,
		       b.id, b.name, b.created_at, b.updated_at
		FROM people p
		INNER JOIN people_types pt ON p.type_id = pt.id
		INNER JOIN buildings b ON p.building_id = b.id
		WHERE p.building_id = ?`, buildingID)
	if err != nil {
		return nil, nil, nil, err
	}
	defer rows.Close()

	people := []Person{}
	peopleTypes := []people_types.PeopleType{}
	buildings := []building.Building{}
	for rows.Next() {
		var p Person
		var pt people_types.PeopleType
		var b building.Building
		err := rows.Scan(&p.ID, &p.Name, &p.Phone, &p.TypeID, &p.BuildingID, &p.CreatedAt, &p.UpdatedAt,
			&pt.ID, &pt.Title,
			&b.ID, &b.Name, &b.CreatedAt, &b.UpdatedAt)
		if err != nil {
			return nil, nil, nil, err
		}
		people = append(people, p)
		peopleTypes = append(peopleTypes, pt)
		buildings = append(buildings, b)
	}
	return people, peopleTypes, buildings, nil
}

func (r *personRepo) TypeIDExists(typeID int) (bool, int, error) {
	var id int
	err := r.db.QueryRow("SELECT id FROM people_types WHERE id = ?", typeID).
		Scan(&id)

	if err == sql.ErrNoRows {
		return false, 0, nil
	}
	if err != nil {
		return false, 0, err
	}

	return true, 0, nil // Return 0 for buildingID since people_types are now universal
}

func (r *personRepo) BuildingIDExists(buildingID int) (bool, error) {
	var id int
	err := r.db.QueryRow("SELECT id FROM buildings WHERE id = ?", buildingID).
		Scan(&id)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *personRepo) PersonNameExists(name string, buildingID int, excludeID int) (bool, error) {
	var exists bool
	var err error

	if excludeID > 0 {
		// For update: exclude the current person ID
		query := `
            SELECT EXISTS(
                SELECT 1 FROM people
                WHERE name = ? AND building_id = ? AND id <> ?
            )
        `
		err = r.db.QueryRow(query, name, buildingID, excludeID).Scan(&exists)
	} else {
		// For create: check if any person with this name exists in the building
		query := `
            SELECT EXISTS(
                SELECT 1 FROM people
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
