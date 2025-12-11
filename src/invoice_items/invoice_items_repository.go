package invoice_items

import (
	"database/sql"
	"fmt"
)

type InvoiceItemRepository interface {
	Create(invoiceItem InvoiceItem) (InvoiceItem, error)
	CreateBatch(invoiceItems []InvoiceItem) error
	GetByInvoiceID(invoiceID int) ([]InvoiceItem, error)
	GetByID(id int) (InvoiceItem, error)
}

type invoiceItemRepo struct {
	db *sql.DB
}

func NewInvoiceItemRepository(db *sql.DB) InvoiceItemRepository {
	return &invoiceItemRepo{db: db}
}

func (r *invoiceItemRepo) Create(invoiceItem InvoiceItem) (InvoiceItem, error) {
	var previousValue interface{}
	if invoiceItem.PreviousValue != nil {
		previousValue = *invoiceItem.PreviousValue
	} else {
		previousValue = nil
	}

	var currentValue interface{}
	if invoiceItem.CurrentValue != nil {
		currentValue = *invoiceItem.CurrentValue
	} else {
		currentValue = nil
	}

	var qty interface{}
	if invoiceItem.Qty != nil {
		qty = *invoiceItem.Qty
	} else {
		qty = nil
	}

	var rate interface{}
	if invoiceItem.Rate != nil {
		rate = *invoiceItem.Rate
	} else {
		rate = nil
	}

	result, err := r.db.Exec("INSERT INTO invoice_items (invoice_id, item_id, item_name, previous_value, current_value, qty, rate, total, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		invoiceItem.InvoiceID, invoiceItem.ItemID, invoiceItem.ItemName, previousValue, currentValue, qty, rate, invoiceItem.Total, invoiceItem.Status)

	if err != nil {
		return invoiceItem, err
	}

	id, _ := result.LastInsertId()
	invoiceItem.ID = int(id)

	err = r.db.QueryRow("SELECT id, invoice_id, item_id, item_name, previous_value, current_value, qty, rate, total, status, created_at, updated_at FROM invoice_items WHERE id = ?", invoiceItem.ID).
		Scan(&invoiceItem.ID, &invoiceItem.InvoiceID, &invoiceItem.ItemID, &invoiceItem.ItemName, &invoiceItem.PreviousValue, &invoiceItem.CurrentValue, &invoiceItem.Qty, &invoiceItem.Rate, &invoiceItem.Total, &invoiceItem.Status, &invoiceItem.CreatedAt, &invoiceItem.UpdatedAt)

	return invoiceItem, err
}

func (r *invoiceItemRepo) CreateBatch(invoiceItems []InvoiceItem) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO invoice_items (invoice_id, item_id, item_name, previous_value, current_value, qty, rate, total, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, invoiceItem := range invoiceItems {
		var previousValue interface{}
		if invoiceItem.PreviousValue != nil {
			previousValue = *invoiceItem.PreviousValue
		} else {
			previousValue = nil
		}

		var currentValue interface{}
		if invoiceItem.CurrentValue != nil {
			currentValue = *invoiceItem.CurrentValue
		} else {
			currentValue = nil
		}

		var qty interface{}
		if invoiceItem.Qty != nil {
			qty = *invoiceItem.Qty
		} else {
			qty = nil
		}

		var rate interface{}
		if invoiceItem.Rate != nil {
			rate = *invoiceItem.Rate
		} else {
			rate = nil
		}

		_, err := stmt.Exec(invoiceItem.InvoiceID, invoiceItem.ItemID, invoiceItem.ItemName, previousValue, currentValue, qty, rate, invoiceItem.Total, invoiceItem.Status)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *invoiceItemRepo) GetByInvoiceID(invoiceID int) ([]InvoiceItem, error) {
	rows, err := r.db.Query("SELECT id, invoice_id, item_id, item_name, previous_value, current_value, qty, rate, total, status, created_at, updated_at FROM invoice_items WHERE invoice_id = ?", invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	invoiceItems := []InvoiceItem{}
	for rows.Next() {
		var invoiceItem InvoiceItem
		err := rows.Scan(&invoiceItem.ID, &invoiceItem.InvoiceID, &invoiceItem.ItemID, &invoiceItem.ItemName, &invoiceItem.PreviousValue, &invoiceItem.CurrentValue, &invoiceItem.Qty, &invoiceItem.Rate, &invoiceItem.Total, &invoiceItem.Status, &invoiceItem.CreatedAt, &invoiceItem.UpdatedAt)
		if err != nil {
			return nil, err
		}
		invoiceItems = append(invoiceItems, invoiceItem)
	}

	return invoiceItems, nil
}

func (r *invoiceItemRepo) GetByID(id int) (InvoiceItem, error) {
	var invoiceItem InvoiceItem
	err := r.db.QueryRow("SELECT id, invoice_id, item_id, item_name, previous_value, current_value, qty, rate, total, status, created_at, updated_at FROM invoice_items WHERE id = ?", id).
		Scan(&invoiceItem.ID, &invoiceItem.InvoiceID, &invoiceItem.ItemID, &invoiceItem.ItemName, &invoiceItem.PreviousValue, &invoiceItem.CurrentValue, &invoiceItem.Qty, &invoiceItem.Rate, &invoiceItem.Total, &invoiceItem.Status, &invoiceItem.CreatedAt, &invoiceItem.UpdatedAt)

	if err == sql.ErrNoRows {
		return invoiceItem, fmt.Errorf("invoice item not found")
	}

	return invoiceItem, err
}

