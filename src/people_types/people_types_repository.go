package people_types

import (
	"database/sql"
	"fmt"
)

type PeopleTypeRepository interface {
	Create(peopleType PeopleType) (PeopleType, error)
	UpdateTitle(title string, id int) (PeopleType, error)
	GetByID(id int) (PeopleType, error)
	GetAll() ([]PeopleType, error)
	TitleExists(title string) (bool, error)
	TitleExistsExcludingID(title string, excludeID int) (bool, error)
}

type peopleTypeRepo struct {
	db *sql.DB
}

func NewPeopleTypeRepository(db *sql.DB) PeopleTypeRepository {
	return &peopleTypeRepo{db: db}
}

func (r *peopleTypeRepo) Create(peopleType PeopleType) (PeopleType, error) {
	result, err := r.db.Exec("INSERT INTO people_types (title) VALUES (?)",
		peopleType.Title)

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
	err = r.db.QueryRow("SELECT id, title FROM people_types WHERE id = ?", id).
		Scan(&peopleType.ID, &peopleType.Title)

	return peopleType, err
}

func (r *peopleTypeRepo) GetByID(id int) (PeopleType, error) {
	var peopleType PeopleType
	err := r.db.QueryRow("SELECT id, title FROM people_types WHERE id = ?", id).
		Scan(&peopleType.ID, &peopleType.Title)
	
	if err == sql.ErrNoRows {
		return peopleType, fmt.Errorf("id does not exist")
	}
	
	return peopleType, err
}

func (r *peopleTypeRepo) GetAll() ([]PeopleType, error) {
	rows, err := r.db.Query("SELECT id, title FROM people_types ORDER BY title")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	peopleTypes := []PeopleType{}
	for rows.Next() {
		var p PeopleType
		err := rows.Scan(&p.ID, &p.Title)
		if err != nil {
			return nil, err
		}
		peopleTypes = append(peopleTypes, p)
	}
	return peopleTypes, nil
}

func (r *peopleTypeRepo) TitleExists(title string) (bool, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM people_types WHERE title = ?", title).
		Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *peopleTypeRepo) TitleExistsExcludingID(title string, excludeID int) (bool, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM people_types WHERE title = ? AND id != ?", title, excludeID).
		Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
