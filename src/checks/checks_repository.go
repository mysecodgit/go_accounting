package checks

import (
	"database/sql"
	"fmt"
)

type CheckRepository interface {
	Create(check Check) (Check, error)
	Update(check Check) (Check, error)
	GetByID(id int) (Check, error)
	GetByBuildingID(buildingID int) ([]Check, error)
}

type checkRepo struct {
	db *sql.DB
}

func NewCheckRepository(db *sql.DB) CheckRepository {
	return &checkRepo{db: db}
}

func (r *checkRepo) Create(check Check) (Check, error) {
	var referenceNumber interface{}
	if check.ReferenceNumber != nil {
		referenceNumber = *check.ReferenceNumber
	} else {
		referenceNumber = nil
	}

	var memo interface{}
	if check.Memo != nil {
		memo = *check.Memo
	} else {
		memo = nil
	}

	result, err := r.db.Exec("INSERT INTO checks (transaction_id, check_date, reference_number, payment_account_id, building_id, memo, total_amount) VALUES (?, ?, ?, ?, ?, ?, ?)",
		check.TransactionID, check.CheckDate, referenceNumber, check.PaymentAccountID, check.BuildingID, memo, check.TotalAmount)

	if err != nil {
		return check, err
	}

	id, _ := result.LastInsertId()
	check.ID = int(id)

	err = r.db.QueryRow("SELECT id, transaction_id, check_date, reference_number, payment_account_id, building_id, memo, total_amount, created_at FROM checks WHERE id = ?", check.ID).
		Scan(&check.ID, &check.TransactionID, &check.CheckDate, &check.ReferenceNumber, &check.PaymentAccountID, &check.BuildingID, &check.Memo, &check.TotalAmount, &check.CreatedAt)

	return check, err
}

func (r *checkRepo) Update(check Check) (Check, error) {
	var referenceNumber interface{}
	if check.ReferenceNumber != nil {
		referenceNumber = *check.ReferenceNumber
	} else {
		referenceNumber = nil
	}

	var memo interface{}
	if check.Memo != nil {
		memo = *check.Memo
	} else {
		memo = nil
	}

	_, err := r.db.Exec("UPDATE checks SET check_date = ?, reference_number = ?, payment_account_id = ?, memo = ?, total_amount = ? WHERE id = ?",
		check.CheckDate, referenceNumber, check.PaymentAccountID, memo, check.TotalAmount, check.ID)

	if err != nil {
		return check, err
	}

	err = r.db.QueryRow("SELECT id, transaction_id, check_date, reference_number, payment_account_id, building_id, memo, total_amount, created_at FROM checks WHERE id = ?", check.ID).
		Scan(&check.ID, &check.TransactionID, &check.CheckDate, &check.ReferenceNumber, &check.PaymentAccountID, &check.BuildingID, &check.Memo, &check.TotalAmount, &check.CreatedAt)

	return check, err
}

func (r *checkRepo) GetByID(id int) (Check, error) {
	var check Check
	err := r.db.QueryRow("SELECT id, transaction_id, check_date, reference_number, payment_account_id, building_id, memo, total_amount, created_at FROM checks WHERE id = ?", id).
		Scan(&check.ID, &check.TransactionID, &check.CheckDate, &check.ReferenceNumber, &check.PaymentAccountID, &check.BuildingID, &check.Memo, &check.TotalAmount, &check.CreatedAt)

	if err == sql.ErrNoRows {
		return check, fmt.Errorf("check not found")
	}

	return check, err
}

func (r *checkRepo) GetByBuildingID(buildingID int) ([]Check, error) {
	rows, err := r.db.Query("SELECT id, transaction_id, check_date, reference_number, payment_account_id, building_id, memo, total_amount, created_at FROM checks WHERE building_id = ? ORDER BY created_at DESC", buildingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	checks := []Check{}
	for rows.Next() {
		var check Check
		err := rows.Scan(&check.ID, &check.TransactionID, &check.CheckDate, &check.ReferenceNumber, &check.PaymentAccountID, &check.BuildingID, &check.Memo, &check.TotalAmount, &check.CreatedAt)
		if err != nil {
			return nil, err
		}
		checks = append(checks, check)
	}

	return checks, nil
}
