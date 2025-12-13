package journal

import (
	"database/sql"
	"fmt"
)

type JournalRepository interface {
	Create(journal Journal) (Journal, error)
	Update(journal Journal) (Journal, error)
	GetByID(id int) (Journal, error)
	GetByBuildingID(buildingID int) ([]Journal, error)
}

type journalRepo struct {
	db *sql.DB
}

func NewJournalRepository(db *sql.DB) JournalRepository {
	return &journalRepo{db: db}
}

func (r *journalRepo) Create(journal Journal) (Journal, error) {
	var memo interface{}
	if journal.Memo != nil {
		memo = *journal.Memo
	} else {
		memo = nil
	}

	result, err := r.db.Exec("INSERT INTO journal (transaction_id, reference, journal_date, building_id, memo, total_amount) VALUES (?, ?, ?, ?, ?, ?)",
		journal.TransactionID, journal.Reference, journal.JournalDate, journal.BuildingID, memo, journal.TotalAmount)

	if err != nil {
		return journal, err
	}

	id, _ := result.LastInsertId()
	journal.ID = int(id)

	err = r.db.QueryRow("SELECT id, transaction_id, reference, journal_date, building_id, memo, total_amount, created_at FROM journal WHERE id = ?", journal.ID).
		Scan(&journal.ID, &journal.TransactionID, &journal.Reference, &journal.JournalDate, &journal.BuildingID, &journal.Memo, &journal.TotalAmount, &journal.CreatedAt)

	return journal, err
}

func (r *journalRepo) Update(journal Journal) (Journal, error) {
	var memo interface{}
	if journal.Memo != nil {
		memo = *journal.Memo
	} else {
		memo = nil
	}

	_, err := r.db.Exec("UPDATE journal SET reference = ?, journal_date = ?, memo = ?, total_amount = ? WHERE id = ?",
		journal.Reference, journal.JournalDate, memo, journal.TotalAmount, journal.ID)

	if err != nil {
		return journal, err
	}

	err = r.db.QueryRow("SELECT id, transaction_id, reference, journal_date, building_id, memo, total_amount, created_at FROM journal WHERE id = ?", journal.ID).
		Scan(&journal.ID, &journal.TransactionID, &journal.Reference, &journal.JournalDate, &journal.BuildingID, &journal.Memo, &journal.TotalAmount, &journal.CreatedAt)

	return journal, err
}

func (r *journalRepo) GetByID(id int) (Journal, error) {
	var journal Journal
	err := r.db.QueryRow("SELECT id, transaction_id, journal_date, building_id, memo, total_amount, created_at FROM journal WHERE id = ?", id).
		Scan(&journal.ID, &journal.TransactionID, &journal.JournalDate, &journal.BuildingID, &journal.Memo, &journal.TotalAmount, &journal.CreatedAt)

	if err == sql.ErrNoRows {
		return journal, fmt.Errorf("journal not found")
	}

	return journal, err
}

func (r *journalRepo) GetByBuildingID(buildingID int) ([]Journal, error) {
	rows, err := r.db.Query("SELECT id, transaction_id, reference, journal_date, building_id, memo, total_amount, created_at FROM journal WHERE building_id = ? ORDER BY created_at DESC", buildingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	journals := []Journal{}
	for rows.Next() {
		var journal Journal
		err := rows.Scan(&journal.ID, &journal.TransactionID, &journal.Reference, &journal.JournalDate, &journal.BuildingID, &journal.Memo, &journal.TotalAmount, &journal.CreatedAt)
		if err != nil {
			return nil, err
		}
		journals = append(journals, journal)
	}

	return journals, nil
}

