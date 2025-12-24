package invoice_applied_discounts

import (
	"database/sql"
	"fmt"
)

type InvoiceAppliedDiscountRepository interface {
	Create(appliedDiscount InvoiceAppliedDiscount) (InvoiceAppliedDiscount, error)
	Update(appliedDiscount InvoiceAppliedDiscount) (InvoiceAppliedDiscount, error)
	GetByID(id int) (InvoiceAppliedDiscount, error)
	GetByInvoiceID(invoiceID int) ([]InvoiceAppliedDiscount, error)
	GetByTransactionID(transactionID int) ([]InvoiceAppliedDiscount, error)
	GetAppliedAmountByInvoiceID(invoiceID int) (float64, error)
}

type invoiceAppliedDiscountRepo struct {
	db *sql.DB
}

func NewInvoiceAppliedDiscountRepository(db *sql.DB) InvoiceAppliedDiscountRepository {
	return &invoiceAppliedDiscountRepo{db: db}
}

func (r *invoiceAppliedDiscountRepo) Create(appliedDiscount InvoiceAppliedDiscount) (InvoiceAppliedDiscount, error) {
	result, err := r.db.Exec("INSERT INTO invoice_applied_discounts (invoice_id, transaction_id, ar_account, income_account, amount, description, date, status, reference) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		appliedDiscount.InvoiceID, appliedDiscount.TransactionID, appliedDiscount.ARAccount, appliedDiscount.IncomeAccount, appliedDiscount.Amount, appliedDiscount.Description, appliedDiscount.Date, appliedDiscount.Status, appliedDiscount.Reference)

	if err != nil {
		return appliedDiscount, err
	}

	id, _ := result.LastInsertId()
	appliedDiscount.ID = int(id)

	err = r.db.QueryRow("SELECT id, reference, invoice_id, transaction_id, ar_account, income_account, amount, description, date, status, created_at, updated_at FROM invoice_applied_discounts WHERE id = ?", appliedDiscount.ID).
		Scan(&appliedDiscount.ID, &appliedDiscount.Reference, &appliedDiscount.InvoiceID, &appliedDiscount.TransactionID, &appliedDiscount.ARAccount, &appliedDiscount.IncomeAccount, &appliedDiscount.Amount, &appliedDiscount.Description, &appliedDiscount.Date, &appliedDiscount.Status, &appliedDiscount.CreatedAt, &appliedDiscount.UpdatedAt)

	return appliedDiscount, err
}

func (r *invoiceAppliedDiscountRepo) GetByID(id int) (InvoiceAppliedDiscount, error) {
	var appliedDiscount InvoiceAppliedDiscount
	err := r.db.QueryRow("SELECT id, reference, invoice_id, transaction_id, ar_account, income_account, amount, description, date, status, created_at, updated_at FROM invoice_applied_discounts WHERE id = ?", id).
		Scan(&appliedDiscount.ID, &appliedDiscount.Reference, &appliedDiscount.InvoiceID, &appliedDiscount.TransactionID, &appliedDiscount.ARAccount, &appliedDiscount.IncomeAccount, &appliedDiscount.Amount, &appliedDiscount.Description, &appliedDiscount.Date, &appliedDiscount.Status, &appliedDiscount.CreatedAt, &appliedDiscount.UpdatedAt)

	if err == sql.ErrNoRows {
		return appliedDiscount, fmt.Errorf("invoice applied discount not found")
	}

	return appliedDiscount, err
}

func (r *invoiceAppliedDiscountRepo) GetByInvoiceID(invoiceID int) ([]InvoiceAppliedDiscount, error) {
	rows, err := r.db.Query("SELECT id, reference, invoice_id, transaction_id, ar_account, income_account, amount, description, date, status, created_at, updated_at FROM invoice_applied_discounts WHERE invoice_id = ? ORDER BY created_at DESC", invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	appliedDiscounts := []InvoiceAppliedDiscount{}
	for rows.Next() {
		var appliedDiscount InvoiceAppliedDiscount
		err := rows.Scan(&appliedDiscount.ID, &appliedDiscount.Reference, &appliedDiscount.InvoiceID, &appliedDiscount.TransactionID, &appliedDiscount.ARAccount, &appliedDiscount.IncomeAccount, &appliedDiscount.Amount, &appliedDiscount.Description, &appliedDiscount.Date, &appliedDiscount.Status, &appliedDiscount.CreatedAt, &appliedDiscount.UpdatedAt)
		if err != nil {
			return nil, err
		}
		appliedDiscounts = append(appliedDiscounts, appliedDiscount)
	}

	return appliedDiscounts, nil
}

func (r *invoiceAppliedDiscountRepo) GetByTransactionID(transactionID int) ([]InvoiceAppliedDiscount, error) {
	rows, err := r.db.Query("SELECT id, reference, invoice_id, transaction_id, ar_account, income_account, amount, description, date, status, created_at, updated_at FROM invoice_applied_discounts WHERE transaction_id = ? AND status = '1'", transactionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	appliedDiscounts := []InvoiceAppliedDiscount{}
	for rows.Next() {
		var appliedDiscount InvoiceAppliedDiscount
		err := rows.Scan(&appliedDiscount.ID, &appliedDiscount.Reference, &appliedDiscount.InvoiceID, &appliedDiscount.TransactionID, &appliedDiscount.ARAccount, &appliedDiscount.IncomeAccount, &appliedDiscount.Amount, &appliedDiscount.Description, &appliedDiscount.Date, &appliedDiscount.Status, &appliedDiscount.CreatedAt, &appliedDiscount.UpdatedAt)
		if err != nil {
			return nil, err
		}
		appliedDiscounts = append(appliedDiscounts, appliedDiscount)
	}

	return appliedDiscounts, nil
}

func (r *invoiceAppliedDiscountRepo) GetAppliedAmountByInvoiceID(invoiceID int) (float64, error) {
	var totalAmount sql.NullFloat64
	err := r.db.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM invoice_applied_discounts WHERE invoice_id = ? AND status = '1'", invoiceID).
		Scan(&totalAmount)

	if err != nil {
		return 0, err
	}

	if totalAmount.Valid {
		return totalAmount.Float64, nil
	}
	return 0, nil
}

func (r *invoiceAppliedDiscountRepo) Update(appliedDiscount InvoiceAppliedDiscount) (InvoiceAppliedDiscount, error) {
	_, err := r.db.Exec("UPDATE invoice_applied_discounts SET status = ? WHERE id = ?",
		appliedDiscount.Status, appliedDiscount.ID)

	if err != nil {
		return appliedDiscount, err
	}

	err = r.db.QueryRow("SELECT id, reference, invoice_id, transaction_id, ar_account, income_account, amount, description, date, status, created_at, updated_at FROM invoice_applied_discounts WHERE id = ?", appliedDiscount.ID).
		Scan(&appliedDiscount.ID, &appliedDiscount.Reference, &appliedDiscount.InvoiceID, &appliedDiscount.TransactionID, &appliedDiscount.ARAccount, &appliedDiscount.IncomeAccount, &appliedDiscount.Amount, &appliedDiscount.Description, &appliedDiscount.Date, &appliedDiscount.Status, &appliedDiscount.CreatedAt, &appliedDiscount.UpdatedAt)

	return appliedDiscount, err
}
