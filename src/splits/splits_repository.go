package splits

import (
	"database/sql"
	"fmt"
)

type SplitRepository interface {
	Create(split Split) (Split, error)
	CreateBatch(splits []Split) error
	GetByTransactionID(transactionID int) ([]Split, error)
	GetByID(id int) (Split, error)
}

type splitRepo struct {
	db *sql.DB
}

func NewSplitRepository(db *sql.DB) SplitRepository {
	return &splitRepo{db: db}
}

func (r *splitRepo) Create(split Split) (Split, error) {
	var peopleID interface{}
	if split.PeopleID != nil {
		peopleID = *split.PeopleID
	} else {
		peopleID = nil
	}

	var debit interface{}
	if split.Debit != nil {
		debit = *split.Debit
	} else {
		debit = nil
	}

	var credit interface{}
	if split.Credit != nil {
		credit = *split.Credit
	} else {
		credit = nil
	}

	result, err := r.db.Exec("INSERT INTO splits (transaction_id, account_id, people_id, debit, credit, status) VALUES (?, ?, ?, ?, ?, ?)",
		split.TransactionID, split.AccountID, peopleID, debit, credit, split.Status)

	if err != nil {
		return split, err
	}

	id, _ := result.LastInsertId()
	split.ID = int(id)

	err = r.db.QueryRow("SELECT id, transaction_id, account_id, people_id, debit, credit, status, created_at, updated_at FROM splits WHERE id = ?", split.ID).
		Scan(&split.ID, &split.TransactionID, &split.AccountID, &split.PeopleID, &split.Debit, &split.Credit, &split.Status, &split.CreatedAt, &split.UpdatedAt)

	return split, err
}

func (r *splitRepo) CreateBatch(splits []Split) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO splits (transaction_id, account_id, people_id, debit, credit, status) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, split := range splits {
		var peopleID interface{}
		if split.PeopleID != nil {
			peopleID = *split.PeopleID
		} else {
			peopleID = nil
		}

		var debit interface{}
		if split.Debit != nil {
			debit = *split.Debit
		} else {
			debit = nil
		}

		var credit interface{}
		if split.Credit != nil {
			credit = *split.Credit
		} else {
			credit = nil
		}

		_, err := stmt.Exec(split.TransactionID, split.AccountID, peopleID, debit, credit, split.Status)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *splitRepo) GetByTransactionID(transactionID int) ([]Split, error) {
	rows, err := r.db.Query("SELECT id, transaction_id, account_id, people_id, debit, credit, status, created_at, updated_at FROM splits WHERE transaction_id = ?", transactionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	splits := []Split{}
	for rows.Next() {
		var split Split
		err := rows.Scan(&split.ID, &split.TransactionID, &split.AccountID, &split.PeopleID, &split.Debit, &split.Credit, &split.Status, &split.CreatedAt, &split.UpdatedAt)
		if err != nil {
			return nil, err
		}
		splits = append(splits, split)
	}

	return splits, nil
}

func (r *splitRepo) GetByID(id int) (Split, error) {
	var split Split
	err := r.db.QueryRow("SELECT id, transaction_id, account_id, people_id, debit, credit, status, created_at, updated_at FROM splits WHERE id = ?", id).
		Scan(&split.ID, &split.TransactionID, &split.AccountID, &split.PeopleID, &split.Debit, &split.Credit, &split.Status, &split.CreatedAt, &split.UpdatedAt)

	if err == sql.ErrNoRows {
		return split, fmt.Errorf("split not found")
	}

	return split, err
}

