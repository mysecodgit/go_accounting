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
	GetByAccountIDAndDateRange(accountID int, buildingID int, startDate string, endDate string) ([]Split, error)
	GetByAccountIDAndDateRangeWithUnit(accountID int, buildingID int, startDate string, endDate string, unitID *int) ([]Split, error)
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

func (r *splitRepo) GetByAccountIDAndDateRange(accountID int, buildingID int, startDate string, endDate string) ([]Split, error) {
	return r.GetByAccountIDAndDateRangeWithUnit(accountID, buildingID, startDate, endDate, nil)
}

func (r *splitRepo) GetByAccountIDAndDateRangeWithUnit(accountID int, buildingID int, startDate string, endDate string, unitID *int) ([]Split, error) {
	query := `
		SELECT s.id, s.transaction_id, s.account_id, s.people_id, s.debit, s.credit, s.status, s.created_at, s.updated_at
		FROM splits s
		INNER JOIN transactions t ON s.transaction_id = t.id
		WHERE s.account_id = ? AND t.building_id = ? AND s.status = '1' AND t.status = '1'
	`
	args := []interface{}{accountID, buildingID}
	
	if startDate != "" {
		query += " AND t.transaction_date >= ?"
		args = append(args, startDate)
	}
	
	if endDate != "" {
		query += " AND t.transaction_date <= ?"
		args = append(args, endDate)
	}
	
	if unitID != nil && *unitID > 0 {
		query += " AND t.unit_id = ?"
		args = append(args, *unitID)
	}
	
	query += " ORDER BY t.transaction_date, s.id"
	
	rows, err := r.db.Query(query, args...)
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

