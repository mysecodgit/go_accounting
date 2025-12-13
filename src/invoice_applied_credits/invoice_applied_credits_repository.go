package invoice_applied_credits

import (
	"database/sql"
	"fmt"
)

type InvoiceAppliedCreditRepository interface {
	Create(appliedCredit InvoiceAppliedCredit) (InvoiceAppliedCredit, error)
	Update(appliedCredit InvoiceAppliedCredit) (InvoiceAppliedCredit, error)
	GetByID(id int) (InvoiceAppliedCredit, error)
	GetByInvoiceID(invoiceID int) ([]InvoiceAppliedCredit, error)
	GetByCreditMemoID(creditMemoID int) ([]InvoiceAppliedCredit, error)
	GetAppliedAmountByCreditMemoID(creditMemoID int) (float64, error)
}

type invoiceAppliedCreditRepo struct {
	db *sql.DB
}

func NewInvoiceAppliedCreditRepository(db *sql.DB) InvoiceAppliedCreditRepository {
	return &invoiceAppliedCreditRepo{db: db}
}

func (r *invoiceAppliedCreditRepo) Create(appliedCredit InvoiceAppliedCredit) (InvoiceAppliedCredit, error) {
	result, err := r.db.Exec("INSERT INTO invoice_applied_credits (transaction_id, invoice_id, credit_memo_id, amount, description, date, status) VALUES (?, ?, ?, ?, ?, ?, ?)",
		appliedCredit.TransactionID, appliedCredit.InvoiceID, appliedCredit.CreditMemoID, appliedCredit.Amount, appliedCredit.Description, appliedCredit.Date, appliedCredit.Status)

	if err != nil {
		return appliedCredit, err
	}

	id, _ := result.LastInsertId()
	appliedCredit.ID = int(id)

	err = r.db.QueryRow("SELECT id, transaction_id, invoice_id, credit_memo_id, amount, description, date, status, created_at, updated_at FROM invoice_applied_credits WHERE id = ?", appliedCredit.ID).
		Scan(&appliedCredit.ID, &appliedCredit.TransactionID, &appliedCredit.InvoiceID, &appliedCredit.CreditMemoID, &appliedCredit.Amount, &appliedCredit.Description, &appliedCredit.Date, &appliedCredit.Status, &appliedCredit.CreatedAt, &appliedCredit.UpdatedAt)

	return appliedCredit, err
}

func (r *invoiceAppliedCreditRepo) GetByID(id int) (InvoiceAppliedCredit, error) {
	var appliedCredit InvoiceAppliedCredit
	err := r.db.QueryRow("SELECT id, transaction_id, invoice_id, credit_memo_id, amount, description, date, status, created_at, updated_at FROM invoice_applied_credits WHERE id = ?", id).
		Scan(&appliedCredit.ID, &appliedCredit.TransactionID, &appliedCredit.InvoiceID, &appliedCredit.CreditMemoID, &appliedCredit.Amount, &appliedCredit.Description, &appliedCredit.Date, &appliedCredit.Status, &appliedCredit.CreatedAt, &appliedCredit.UpdatedAt)

	if err == sql.ErrNoRows {
		return appliedCredit, fmt.Errorf("invoice applied credit not found")
	}

	return appliedCredit, err
}

func (r *invoiceAppliedCreditRepo) GetByInvoiceID(invoiceID int) ([]InvoiceAppliedCredit, error) {
	rows, err := r.db.Query("SELECT id, transaction_id, invoice_id, credit_memo_id, amount, description, date, status, created_at, updated_at FROM invoice_applied_credits WHERE invoice_id = ? ORDER BY created_at DESC", invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	appliedCredits := []InvoiceAppliedCredit{}
	for rows.Next() {
		var appliedCredit InvoiceAppliedCredit
		err := rows.Scan(&appliedCredit.ID, &appliedCredit.TransactionID, &appliedCredit.InvoiceID, &appliedCredit.CreditMemoID, &appliedCredit.Amount, &appliedCredit.Description, &appliedCredit.Date, &appliedCredit.Status, &appliedCredit.CreatedAt, &appliedCredit.UpdatedAt)
		if err != nil {
			return nil, err
		}
		appliedCredits = append(appliedCredits, appliedCredit)
	}

	return appliedCredits, nil
}

func (r *invoiceAppliedCreditRepo) GetByCreditMemoID(creditMemoID int) ([]InvoiceAppliedCredit, error) {
	rows, err := r.db.Query("SELECT id, transaction_id, invoice_id, credit_memo_id, amount, description, date, status, created_at, updated_at FROM invoice_applied_credits WHERE credit_memo_id = ? AND status = '1'", creditMemoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	appliedCredits := []InvoiceAppliedCredit{}
	for rows.Next() {
		var appliedCredit InvoiceAppliedCredit
		err := rows.Scan(&appliedCredit.ID, &appliedCredit.TransactionID, &appliedCredit.InvoiceID, &appliedCredit.CreditMemoID, &appliedCredit.Amount, &appliedCredit.Description, &appliedCredit.Date, &appliedCredit.Status, &appliedCredit.CreatedAt, &appliedCredit.UpdatedAt)
		if err != nil {
			return nil, err
		}
		appliedCredits = append(appliedCredits, appliedCredit)
	}

	return appliedCredits, nil
}

func (r *invoiceAppliedCreditRepo) GetAppliedAmountByCreditMemoID(creditMemoID int) (float64, error) {
	var totalAmount sql.NullFloat64
	err := r.db.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM invoice_applied_credits WHERE credit_memo_id = ? AND status = '1'", creditMemoID).
		Scan(&totalAmount)

	if err != nil {
		return 0, err
	}

	if totalAmount.Valid {
		return totalAmount.Float64, nil
	}
	return 0, nil
}

func (r *invoiceAppliedCreditRepo) Update(appliedCredit InvoiceAppliedCredit) (InvoiceAppliedCredit, error) {
	_, err := r.db.Exec("UPDATE invoice_applied_credits SET status = ? WHERE id = ?",
		appliedCredit.Status, appliedCredit.ID)

	if err != nil {
		return appliedCredit, err
	}

	err = r.db.QueryRow("SELECT id, transaction_id, invoice_id, credit_memo_id, amount, description, date, status, created_at, updated_at FROM invoice_applied_credits WHERE id = ?", appliedCredit.ID).
		Scan(&appliedCredit.ID, &appliedCredit.TransactionID, &appliedCredit.InvoiceID, &appliedCredit.CreditMemoID, &appliedCredit.Amount, &appliedCredit.Description, &appliedCredit.Date, &appliedCredit.Status, &appliedCredit.CreatedAt, &appliedCredit.UpdatedAt)

	return appliedCredit, err
}

