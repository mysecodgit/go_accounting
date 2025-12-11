package sales_receipt

import (
	"database/sql"
	"fmt"
)

type SalesReceiptRepository interface {
	Create(receipt SalesReceipt) (SalesReceipt, error)
	GetByID(id int) (SalesReceipt, error)
	GetByBuildingID(buildingID int) ([]SalesReceipt, error)
	GetNextReceiptNo(buildingID int) (int, error)
	CheckDuplicateReceiptNo(buildingID int, receiptNo int, excludeID int) (bool, error)
}

type salesReceiptRepo struct {
	db *sql.DB
}

func NewSalesReceiptRepository(db *sql.DB) SalesReceiptRepository {
	return &salesReceiptRepo{db: db}
}

func (r *salesReceiptRepo) Create(receipt SalesReceipt) (SalesReceipt, error) {
	var unitID interface{}
	if receipt.UnitID != nil {
		unitID = *receipt.UnitID
	} else {
		unitID = nil
	}

	var peopleID interface{}
	if receipt.PeopleID != nil {
		peopleID = *receipt.PeopleID
	} else {
		peopleID = nil
	}

	var cancelReason interface{}
	if receipt.CancelReason != nil {
		cancelReason = *receipt.CancelReason
	} else {
		cancelReason = nil
	}

	result, err := r.db.Exec("INSERT INTO sales_receipt (receipt_no, transaction_id, receipt_date, unit_id, people_id, user_id, account_id, amount, description, cancel_reason, status, building_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		receipt.ReceiptNo, receipt.TransactionID, receipt.ReceiptDate, unitID, peopleID, receipt.UserID, receipt.AccountID, receipt.Amount, receipt.Description, cancelReason, receipt.Status, receipt.BuildingID)

	if err != nil {
		return receipt, err
	}

	id, _ := result.LastInsertId()
	receipt.ID = int(id)

	err = r.db.QueryRow("SELECT id, receipt_no, transaction_id, receipt_date, unit_id, people_id, user_id, account_id, amount, description, cancel_reason, status, building_id, createdAt, updatedAt FROM sales_receipt WHERE id = ?", receipt.ID).
		Scan(&receipt.ID, &receipt.ReceiptNo, &receipt.TransactionID, &receipt.ReceiptDate, &receipt.UnitID, &receipt.PeopleID, &receipt.UserID, &receipt.AccountID, &receipt.Amount, &receipt.Description, &receipt.CancelReason, &receipt.Status, &receipt.BuildingID, &receipt.CreatedAt, &receipt.UpdatedAt)

	return receipt, err
}

func (r *salesReceiptRepo) GetByID(id int) (SalesReceipt, error) {
	var receipt SalesReceipt
	err := r.db.QueryRow("SELECT id, receipt_no, transaction_id, receipt_date, unit_id, people_id, user_id, account_id, amount, description, cancel_reason, status, building_id, createdAt, updatedAt FROM sales_receipt WHERE id = ?", id).
		Scan(&receipt.ID, &receipt.ReceiptNo, &receipt.TransactionID, &receipt.ReceiptDate, &receipt.UnitID, &receipt.PeopleID, &receipt.UserID, &receipt.AccountID, &receipt.Amount, &receipt.Description, &receipt.CancelReason, &receipt.Status, &receipt.BuildingID, &receipt.CreatedAt, &receipt.UpdatedAt)

	if err == sql.ErrNoRows {
		return receipt, fmt.Errorf("sales receipt not found")
	}

	return receipt, err
}

func (r *salesReceiptRepo) GetByBuildingID(buildingID int) ([]SalesReceipt, error) {
	rows, err := r.db.Query("SELECT id, receipt_no, transaction_id, receipt_date, unit_id, people_id, user_id, account_id, amount, description, cancel_reason, status, building_id, createdAt, updatedAt FROM sales_receipt WHERE building_id = ? ORDER BY createdAt DESC", buildingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	receipts := []SalesReceipt{}
	for rows.Next() {
		var receipt SalesReceipt
		err := rows.Scan(&receipt.ID, &receipt.ReceiptNo, &receipt.TransactionID, &receipt.ReceiptDate, &receipt.UnitID, &receipt.PeopleID, &receipt.UserID, &receipt.AccountID, &receipt.Amount, &receipt.Description, &receipt.CancelReason, &receipt.Status, &receipt.BuildingID, &receipt.CreatedAt, &receipt.UpdatedAt)
		if err != nil {
			return nil, err
		}
		receipts = append(receipts, receipt)
	}

	return receipts, nil
}

func (r *salesReceiptRepo) GetNextReceiptNo(buildingID int) (int, error) {
	var maxNo sql.NullInt64
	err := r.db.QueryRow("SELECT MAX(receipt_no) FROM sales_receipt WHERE building_id = ?", buildingID).Scan(&maxNo)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	if maxNo.Valid {
		return int(maxNo.Int64) + 1, nil
	}

	return 1, nil
}

func (r *salesReceiptRepo) CheckDuplicateReceiptNo(buildingID int, receiptNo int, excludeID int) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM sales_receipt WHERE building_id = ? AND receipt_no = ?"
	args := []interface{}{buildingID, receiptNo}

	if excludeID > 0 {
		query += " AND id != ?"
		args = append(args, excludeID)
	}

	err := r.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
