package receipt_items

import (
	"database/sql"
)

type ReceiptItemRepository interface {
	Create(receiptItem ReceiptItem) (ReceiptItem, error)
	CreateBatch(receiptItems []ReceiptItem) error
	GetByReceiptID(receiptID int) ([]ReceiptItem, error)
	GetByID(id int) (ReceiptItem, error)
}

type receiptItemRepo struct {
	db *sql.DB
}

func NewReceiptItemRepository(db *sql.DB) ReceiptItemRepository {
	return &receiptItemRepo{db: db}
}

func (r *receiptItemRepo) Create(receiptItem ReceiptItem) (ReceiptItem, error) {
	var previousValue interface{}
	if receiptItem.PreviousValue != nil {
		previousValue = *receiptItem.PreviousValue
	} else {
		previousValue = nil
	}

	var currentValue interface{}
	if receiptItem.CurrentValue != nil {
		currentValue = *receiptItem.CurrentValue
	} else {
		currentValue = nil
	}

	var qty interface{}
	if receiptItem.Qty != nil {
		qty = *receiptItem.Qty
	} else {
		qty = nil
	}

	var rate interface{}
	if receiptItem.Rate != nil {
		rate = *receiptItem.Rate
	} else {
		rate = nil
	}

	result, err := r.db.Exec("INSERT INTO receipt_items (receipt_id, item_id, item_name, previous_value, current_value, qty, rate, total, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		receiptItem.ReceiptID, receiptItem.ItemID, receiptItem.ItemName, previousValue, currentValue, qty, rate, receiptItem.Total, receiptItem.Status)

	if err != nil {
		return receiptItem, err
	}

	id, _ := result.LastInsertId()
	receiptItem.ID = int(id)

	err = r.db.QueryRow("SELECT id, receipt_id, item_id, item_name, previous_value, current_value, qty, rate, total, status, created_at, updated_at FROM receipt_items WHERE id = ?", receiptItem.ID).
		Scan(&receiptItem.ID, &receiptItem.ReceiptID, &receiptItem.ItemID, &receiptItem.ItemName, &receiptItem.PreviousValue, &receiptItem.CurrentValue, &receiptItem.Qty, &receiptItem.Rate, &receiptItem.Total, &receiptItem.Status, &receiptItem.CreatedAt, &receiptItem.UpdatedAt)

	return receiptItem, err
}

func (r *receiptItemRepo) CreateBatch(receiptItems []ReceiptItem) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO receipt_items (receipt_id, item_id, item_name, previous_value, current_value, qty, rate, total, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, receiptItem := range receiptItems {
		var previousValue interface{}
		if receiptItem.PreviousValue != nil {
			previousValue = *receiptItem.PreviousValue
		} else {
			previousValue = nil
		}

		var currentValue interface{}
		if receiptItem.CurrentValue != nil {
			currentValue = *receiptItem.CurrentValue
		} else {
			currentValue = nil
		}

		var qty interface{}
		if receiptItem.Qty != nil {
			qty = *receiptItem.Qty
		} else {
			qty = nil
		}

		var rate interface{}
		if receiptItem.Rate != nil {
			rate = *receiptItem.Rate
		} else {
			rate = nil
		}

		_, err := stmt.Exec(receiptItem.ReceiptID, receiptItem.ItemID, receiptItem.ItemName, previousValue, currentValue, qty, rate, receiptItem.Total, receiptItem.Status)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *receiptItemRepo) GetByReceiptID(receiptID int) ([]ReceiptItem, error) {
	rows, err := r.db.Query("SELECT id, receipt_id, item_id, item_name, previous_value, current_value, qty, rate, total, status, created_at, updated_at FROM receipt_items WHERE receipt_id = ?", receiptID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	receiptItems := []ReceiptItem{}
	for rows.Next() {
		var receiptItem ReceiptItem
		err := rows.Scan(&receiptItem.ID, &receiptItem.ReceiptID, &receiptItem.ItemID, &receiptItem.ItemName, &receiptItem.PreviousValue, &receiptItem.CurrentValue, &receiptItem.Qty, &receiptItem.Rate, &receiptItem.Total, &receiptItem.Status, &receiptItem.CreatedAt, &receiptItem.UpdatedAt)
		if err != nil {
			return nil, err
		}
		receiptItems = append(receiptItems, receiptItem)
	}

	return receiptItems, nil
}

func (r *receiptItemRepo) GetByID(id int) (ReceiptItem, error) {
	var receiptItem ReceiptItem
	err := r.db.QueryRow("SELECT id, receipt_id, item_id, item_name, previous_value, current_value, qty, rate, total, status, created_at, updated_at FROM receipt_items WHERE id = ?", id).
		Scan(&receiptItem.ID, &receiptItem.ReceiptID, &receiptItem.ItemID, &receiptItem.ItemName, &receiptItem.PreviousValue, &receiptItem.CurrentValue, &receiptItem.Qty, &receiptItem.Rate, &receiptItem.Total, &receiptItem.Status, &receiptItem.CreatedAt, &receiptItem.UpdatedAt)

	if err == sql.ErrNoRows {
		return receiptItem, err
	}

	return receiptItem, err
}
