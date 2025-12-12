package journal_lines

import (
	"database/sql"
	"fmt"
)

type JournalLineRepository interface {
	Create(journalLine JournalLine) (JournalLine, error)
	CreateBatch(journalLines []JournalLine) error
	GetByJournalID(journalID int) ([]JournalLine, error)
	GetByID(id int) (JournalLine, error)
}

type journalLineRepo struct {
	db *sql.DB
}

func NewJournalLineRepository(db *sql.DB) JournalLineRepository {
	return &journalLineRepo{db: db}
}

func (r *journalLineRepo) Create(journalLine JournalLine) (JournalLine, error) {
	var unitID interface{}
	if journalLine.UnitID != nil {
		unitID = *journalLine.UnitID
	} else {
		unitID = nil
	}

	var peopleID interface{}
	if journalLine.PeopleID != nil {
		peopleID = *journalLine.PeopleID
	} else {
		peopleID = nil
	}

	var description interface{}
	if journalLine.Description != nil {
		description = *journalLine.Description
	} else {
		description = nil
	}

	var debit interface{}
	if journalLine.Debit != nil {
		debit = *journalLine.Debit
	} else {
		debit = 0
	}

	var credit interface{}
	if journalLine.Credit != nil {
		credit = *journalLine.Credit
	} else {
		credit = 0
	}

	result, err := r.db.Exec("INSERT INTO journal_lines (journal_id, account_id, unit_id, people_id, description, debit, credit) VALUES (?, ?, ?, ?, ?, ?, ?)",
		journalLine.JournalID, journalLine.AccountID, unitID, peopleID, description, debit, credit)

	if err != nil {
		return journalLine, err
	}

	id, _ := result.LastInsertId()
	journalLine.ID = int(id)

	err = r.db.QueryRow("SELECT id, journal_id, account_id, unit_id, people_id, description, debit, credit FROM journal_lines WHERE id = ?", journalLine.ID).
		Scan(&journalLine.ID, &journalLine.JournalID, &journalLine.AccountID, &journalLine.UnitID, &journalLine.PeopleID, &journalLine.Description, &journalLine.Debit, &journalLine.Credit)

	return journalLine, err
}

func (r *journalLineRepo) CreateBatch(journalLines []JournalLine) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO journal_lines (journal_id, account_id, unit_id, people_id, description, debit, credit) VALUES (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, journalLine := range journalLines {
		var unitID interface{}
		if journalLine.UnitID != nil {
			unitID = *journalLine.UnitID
		} else {
			unitID = nil
		}

		var peopleID interface{}
		if journalLine.PeopleID != nil {
			peopleID = *journalLine.PeopleID
		} else {
			peopleID = nil
		}

		var description interface{}
		if journalLine.Description != nil {
			description = *journalLine.Description
		} else {
			description = nil
		}

		var debit interface{}
		if journalLine.Debit != nil {
			debit = *journalLine.Debit
		} else {
			debit = 0
		}

		var credit interface{}
		if journalLine.Credit != nil {
			credit = *journalLine.Credit
		} else {
			credit = 0
		}

		_, err := stmt.Exec(journalLine.JournalID, journalLine.AccountID, unitID, peopleID, description, debit, credit)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *journalLineRepo) GetByJournalID(journalID int) ([]JournalLine, error) {
	rows, err := r.db.Query("SELECT id, journal_id, account_id, unit_id, people_id, description, debit, credit FROM journal_lines WHERE journal_id = ?", journalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	journalLines := []JournalLine{}
	for rows.Next() {
		var journalLine JournalLine
		var debit sql.NullFloat64
		var credit sql.NullFloat64
		err := rows.Scan(&journalLine.ID, &journalLine.JournalID, &journalLine.AccountID, &journalLine.UnitID, &journalLine.PeopleID, &journalLine.Description, &debit, &credit)
		if err != nil {
			return nil, err
		}
		if debit.Valid {
			journalLine.Debit = &debit.Float64
		}
		if credit.Valid {
			journalLine.Credit = &credit.Float64
		}
		journalLines = append(journalLines, journalLine)
	}

	return journalLines, nil
}

func (r *journalLineRepo) GetByID(id int) (JournalLine, error) {
	var journalLine JournalLine
	var debit sql.NullFloat64
	var credit sql.NullFloat64
	err := r.db.QueryRow("SELECT id, journal_id, account_id, unit_id, people_id, description, debit, credit FROM journal_lines WHERE id = ?", id).
		Scan(&journalLine.ID, &journalLine.JournalID, &journalLine.AccountID, &journalLine.UnitID, &journalLine.PeopleID, &journalLine.Description, &debit, &credit)

	if err == sql.ErrNoRows {
		return journalLine, fmt.Errorf("journal line not found")
	}

	if debit.Valid {
		journalLine.Debit = &debit.Float64
	}
	if credit.Valid {
		journalLine.Credit = &credit.Float64
	}

	return journalLine, err
}

