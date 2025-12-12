package expense_lines

import (
	"database/sql"
	"fmt"
)

type ExpenseLineRepository interface {
	Create(expenseLine ExpenseLine) (ExpenseLine, error)
	CreateBatch(expenseLines []ExpenseLine) error
	GetByCheckID(checkID int) ([]ExpenseLine, error)
	GetByID(id int) (ExpenseLine, error)
}

type expenseLineRepo struct {
	db *sql.DB
}

func NewExpenseLineRepository(db *sql.DB) ExpenseLineRepository {
	return &expenseLineRepo{db: db}
}

func (r *expenseLineRepo) Create(expenseLine ExpenseLine) (ExpenseLine, error) {
	var unitID interface{}
	if expenseLine.UnitID != nil {
		unitID = *expenseLine.UnitID
	} else {
		unitID = nil
	}

	var peopleID interface{}
	if expenseLine.PeopleID != nil {
		peopleID = *expenseLine.PeopleID
	} else {
		peopleID = nil
	}

	var description interface{}
	if expenseLine.Description != nil {
		description = *expenseLine.Description
	} else {
		description = nil
	}

	result, err := r.db.Exec("INSERT INTO expense_lines (check_id, account_id, unit_id, people_id, description, amount) VALUES (?, ?, ?, ?, ?, ?)",
		expenseLine.CheckID, expenseLine.AccountID, unitID, peopleID, description, expenseLine.Amount)

	if err != nil {
		return expenseLine, err
	}

	id, _ := result.LastInsertId()
	expenseLine.ID = int(id)

	err = r.db.QueryRow("SELECT id, check_id, account_id, unit_id, people_id, description, amount FROM expense_lines WHERE id = ?", expenseLine.ID).
		Scan(&expenseLine.ID, &expenseLine.CheckID, &expenseLine.AccountID, &expenseLine.UnitID, &expenseLine.PeopleID, &expenseLine.Description, &expenseLine.Amount)

	return expenseLine, err
}

func (r *expenseLineRepo) CreateBatch(expenseLines []ExpenseLine) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO expense_lines (check_id, account_id, unit_id, people_id, description, amount) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, expenseLine := range expenseLines {
		var unitID interface{}
		if expenseLine.UnitID != nil {
			unitID = *expenseLine.UnitID
		} else {
			unitID = nil
		}

		var peopleID interface{}
		if expenseLine.PeopleID != nil {
			peopleID = *expenseLine.PeopleID
		} else {
			peopleID = nil
		}

		var description interface{}
		if expenseLine.Description != nil {
			description = *expenseLine.Description
		} else {
			description = nil
		}

		_, err := stmt.Exec(expenseLine.CheckID, expenseLine.AccountID, unitID, peopleID, description, expenseLine.Amount)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *expenseLineRepo) GetByCheckID(checkID int) ([]ExpenseLine, error) {
	rows, err := r.db.Query("SELECT id, check_id, account_id, unit_id, people_id, description, amount FROM expense_lines WHERE check_id = ?", checkID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	expenseLines := []ExpenseLine{}
	for rows.Next() {
		var expenseLine ExpenseLine
		err := rows.Scan(&expenseLine.ID, &expenseLine.CheckID, &expenseLine.AccountID, &expenseLine.UnitID, &expenseLine.PeopleID, &expenseLine.Description, &expenseLine.Amount)
		if err != nil {
			return nil, err
		}
		expenseLines = append(expenseLines, expenseLine)
	}

	return expenseLines, nil
}

func (r *expenseLineRepo) GetByID(id int) (ExpenseLine, error) {
	var expenseLine ExpenseLine
	err := r.db.QueryRow("SELECT id, check_id, account_id, unit_id, people_id, description, amount FROM expense_lines WHERE id = ?", id).
		Scan(&expenseLine.ID, &expenseLine.CheckID, &expenseLine.AccountID, &expenseLine.UnitID, &expenseLine.PeopleID, &expenseLine.Description, &expenseLine.Amount)

	if err == sql.ErrNoRows {
		return expenseLine, fmt.Errorf("expense line not found")
	}

	return expenseLine, err
}
