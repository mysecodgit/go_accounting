package transactions

import (
	"database/sql"
	"fmt"
)

type TransactionRepository interface {
	Create(transaction Transaction) (Transaction, error)
	Update(transaction Transaction) (Transaction, error)
	GetByID(id int) (Transaction, error)
	GetByBuildingID(buildingID int) ([]Transaction, error)
}

type transactionRepo struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) TransactionRepository {
	return &transactionRepo{db: db}
}

func (r *transactionRepo) Create(transaction Transaction) (Transaction, error) {
	var unitID interface{}
	if transaction.UnitID != nil {
		unitID = *transaction.UnitID
	} else {
		unitID = nil
	}

	result, err := r.db.Exec("INSERT INTO transactions (type, transaction_date, memo, status, building_id, user_id, unit_id) VALUES (?, ?, ?, ?, ?, ?, ?)",
		transaction.Type, transaction.TransactionDate, transaction.Memo, transaction.Status, transaction.BuildingID, transaction.UserID, unitID)

	if err != nil {
		return transaction, err
	}

	id, _ := result.LastInsertId()
	transaction.ID = int(id)

	err = r.db.QueryRow("SELECT id, type, transaction_date, memo, status, building_id, user_id, unit_id, created_at, updated_at FROM transactions WHERE id = ?", transaction.ID).
		Scan(&transaction.ID, &transaction.Type, &transaction.TransactionDate, &transaction.Memo, &transaction.Status, &transaction.BuildingID, &transaction.UserID, &transaction.UnitID, &transaction.CreatedAt, &transaction.UpdatedAt)

	return transaction, err
}

func (r *transactionRepo) Update(transaction Transaction) (Transaction, error) {
	_, err := r.db.Exec("UPDATE transactions SET type = ?, transaction_date = ?, memo = ? WHERE id = ?",
		transaction.Type, transaction.TransactionDate, transaction.Memo, transaction.ID)

	if err != nil {
		return transaction, err
	}

	err = r.db.QueryRow("SELECT id, type, transaction_date, memo, status, building_id, user_id, unit_id, created_at, updated_at FROM transactions WHERE id = ?", transaction.ID).
		Scan(&transaction.ID, &transaction.Type, &transaction.TransactionDate, &transaction.Memo, &transaction.Status, &transaction.BuildingID, &transaction.UserID, &transaction.UnitID, &transaction.CreatedAt, &transaction.UpdatedAt)

	return transaction, err
}

func (r *transactionRepo) GetByID(id int) (Transaction, error) {
	var transaction Transaction
	err := r.db.QueryRow("SELECT id, type, transaction_date, memo, status, building_id, user_id, unit_id, created_at, updated_at FROM transactions WHERE id = ?", id).
		Scan(&transaction.ID, &transaction.Type, &transaction.TransactionDate, &transaction.Memo, &transaction.Status, &transaction.BuildingID, &transaction.UserID, &transaction.UnitID, &transaction.CreatedAt, &transaction.UpdatedAt)

	if err == sql.ErrNoRows {
		return transaction, fmt.Errorf("transaction not found")
	}

	return transaction, err
}

func (r *transactionRepo) GetByBuildingID(buildingID int) ([]Transaction, error) {
	rows, err := r.db.Query("SELECT id, type, transaction_date, memo, status, building_id, user_id, unit_id, created_at, updated_at FROM transactions WHERE building_id = ? ORDER BY created_at DESC", buildingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transactions := []Transaction{}
	for rows.Next() {
		var transaction Transaction
		err := rows.Scan(&transaction.ID, &transaction.Type, &transaction.TransactionDate, &transaction.Memo, &transaction.Status, &transaction.BuildingID, &transaction.UserID, &transaction.UnitID, &transaction.CreatedAt, &transaction.UpdatedAt)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

