package credit_memo

import (
	"database/sql"
	"fmt"
)

type CreditMemoRepository interface {
	Create(creditMemo CreditMemo) (CreditMemo, error)
	Update(creditMemo CreditMemo) (CreditMemo, error)
	GetByID(id int) (CreditMemo, error)
	GetByBuildingID(buildingID int) ([]CreditMemo, error)
}

type creditMemoRepo struct {
	db *sql.DB
}

func NewCreditMemoRepository(db *sql.DB) CreditMemoRepository {
	return &creditMemoRepo{db: db}
}

func (r *creditMemoRepo) Create(creditMemo CreditMemo) (CreditMemo, error) {
	result, err := r.db.Exec("INSERT INTO credit_memo (transaction_id, date, user_id, deposit_to, liability_account, people_id, building_id, unit_id, amount, description, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		creditMemo.TransactionID, creditMemo.Date, creditMemo.UserID, creditMemo.DepositTo, creditMemo.LiabilityAccount, creditMemo.PeopleID, creditMemo.BuildingID, creditMemo.UnitID, creditMemo.Amount, creditMemo.Description, creditMemo.Status)

	if err != nil {
		return creditMemo, err
	}

	id, _ := result.LastInsertId()
	creditMemo.ID = int(id)

	err = r.db.QueryRow("SELECT id, transaction_id, date, user_id, deposit_to, liability_account, people_id, building_id, unit_id, amount, description, status, created_at, updated_at FROM credit_memo WHERE id = ?", creditMemo.ID).
		Scan(&creditMemo.ID, &creditMemo.TransactionID, &creditMemo.Date, &creditMemo.UserID, &creditMemo.DepositTo, &creditMemo.LiabilityAccount, &creditMemo.PeopleID, &creditMemo.BuildingID, &creditMemo.UnitID, &creditMemo.Amount, &creditMemo.Description, &creditMemo.Status, &creditMemo.CreatedAt, &creditMemo.UpdatedAt)

	return creditMemo, err
}

func (r *creditMemoRepo) Update(creditMemo CreditMemo) (CreditMemo, error) {
	_, err := r.db.Exec("UPDATE credit_memo SET date = ?, deposit_to = ?, liability_account = ?, people_id = ?, unit_id = ?, amount = ?, description = ? WHERE id = ?",
		creditMemo.Date, creditMemo.DepositTo, creditMemo.LiabilityAccount, creditMemo.PeopleID, creditMemo.UnitID, creditMemo.Amount, creditMemo.Description, creditMemo.ID)

	if err != nil {
		return creditMemo, err
	}

	err = r.db.QueryRow("SELECT id, transaction_id, date, user_id, deposit_to, liability_account, people_id, building_id, unit_id, amount, description, status, created_at, updated_at FROM credit_memo WHERE id = ?", creditMemo.ID).
		Scan(&creditMemo.ID, &creditMemo.TransactionID, &creditMemo.Date, &creditMemo.UserID, &creditMemo.DepositTo, &creditMemo.LiabilityAccount, &creditMemo.PeopleID, &creditMemo.BuildingID, &creditMemo.UnitID, &creditMemo.Amount, &creditMemo.Description, &creditMemo.Status, &creditMemo.CreatedAt, &creditMemo.UpdatedAt)

	return creditMemo, err
}

func (r *creditMemoRepo) GetByID(id int) (CreditMemo, error) {
	var creditMemo CreditMemo
	err := r.db.QueryRow("SELECT id, transaction_id, date, user_id, deposit_to, liability_account, people_id, building_id, unit_id, amount, description, status, created_at, updated_at FROM credit_memo WHERE id = ?", id).
		Scan(&creditMemo.ID, &creditMemo.TransactionID, &creditMemo.Date, &creditMemo.UserID, &creditMemo.DepositTo, &creditMemo.LiabilityAccount, &creditMemo.PeopleID, &creditMemo.BuildingID, &creditMemo.UnitID, &creditMemo.Amount, &creditMemo.Description, &creditMemo.Status, &creditMemo.CreatedAt, &creditMemo.UpdatedAt)

	if err == sql.ErrNoRows {
		return creditMemo, fmt.Errorf("credit memo not found")
	}

	return creditMemo, err
}

func (r *creditMemoRepo) GetByBuildingID(buildingID int) ([]CreditMemo, error) {
	rows, err := r.db.Query("SELECT id, transaction_id, date, user_id, deposit_to, liability_account, people_id, building_id, unit_id, amount, description, status, created_at, updated_at FROM credit_memo WHERE building_id = ? ORDER BY created_at DESC", buildingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	creditMemos := []CreditMemo{}
	for rows.Next() {
		var creditMemo CreditMemo
		err := rows.Scan(&creditMemo.ID, &creditMemo.TransactionID, &creditMemo.Date, &creditMemo.UserID, &creditMemo.DepositTo, &creditMemo.LiabilityAccount, &creditMemo.PeopleID, &creditMemo.BuildingID, &creditMemo.UnitID, &creditMemo.Amount, &creditMemo.Description, &creditMemo.Status, &creditMemo.CreatedAt, &creditMemo.UpdatedAt)
		if err != nil {
			return nil, err
		}
		creditMemos = append(creditMemos, creditMemo)
	}

	return creditMemos, nil
}

