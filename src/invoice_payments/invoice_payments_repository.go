package invoice_payments

import (
	"database/sql"
	"fmt"
)

type InvoicePaymentRepository interface {
	Create(payment InvoicePayment) (InvoicePayment, error)
	Update(payment InvoicePayment) (InvoicePayment, error)
	GetByID(id int) (InvoicePayment, error)
	GetByInvoiceID(invoiceID int) ([]InvoicePayment, error)
	GetByBuildingID(buildingID int) ([]InvoicePayment, error)
}

type invoicePaymentRepo struct {
	db *sql.DB
}

func NewInvoicePaymentRepository(db *sql.DB) InvoicePaymentRepository {
	return &invoicePaymentRepo{db: db}
}

func (r *invoicePaymentRepo) Create(payment InvoicePayment) (InvoicePayment, error) {
	result, err := r.db.Exec("INSERT INTO invoice_payments (transaction_id, reference, date, invoice_id, user_id, account_id, amount, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		payment.TransactionID, payment.Reference, payment.Date, payment.InvoiceID, payment.UserID, payment.AccountID, payment.Amount, payment.Status)

	if err != nil {
		return payment, err
	}

	id, _ := result.LastInsertId()
	payment.ID = int(id)

	err = r.db.QueryRow("SELECT id, transaction_id, reference, date, invoice_id, user_id, account_id, amount, status, createdAt, updatedAt FROM invoice_payments WHERE id = ?", payment.ID).
		Scan(&payment.ID, &payment.TransactionID, &payment.Reference, &payment.Date, &payment.InvoiceID, &payment.UserID, &payment.AccountID, &payment.Amount, &payment.Status, &payment.CreatedAt, &payment.UpdatedAt)

	return payment, err
}

func (r *invoicePaymentRepo) Update(payment InvoicePayment) (InvoicePayment, error) {
	_, err := r.db.Exec("UPDATE invoice_payments SET reference = ?, date = ?, account_id = ?, amount = ?, status = ? WHERE id = ?",
		payment.Reference, payment.Date, payment.AccountID, payment.Amount, payment.Status, payment.ID)

	if err != nil {
		return payment, err
	}

	err = r.db.QueryRow("SELECT id, transaction_id, reference, date, invoice_id, user_id, account_id, amount, status, createdAt, updatedAt FROM invoice_payments WHERE id = ?", payment.ID).
		Scan(&payment.ID, &payment.TransactionID, &payment.Reference, &payment.Date, &payment.InvoiceID, &payment.UserID, &payment.AccountID, &payment.Amount, &payment.Status, &payment.CreatedAt, &payment.UpdatedAt)

	return payment, err
}

func (r *invoicePaymentRepo) GetByID(id int) (InvoicePayment, error) {
	var payment InvoicePayment
	err := r.db.QueryRow("SELECT id, transaction_id, date, invoice_id, user_id, account_id, amount, status, createdAt, updatedAt FROM invoice_payments WHERE id = ?", id).
		Scan(&payment.ID, &payment.TransactionID, &payment.Date, &payment.InvoiceID, &payment.UserID, &payment.AccountID, &payment.Amount, &payment.Status, &payment.CreatedAt, &payment.UpdatedAt)

	if err == sql.ErrNoRows {
		return payment, fmt.Errorf("invoice payment not found")
	}

	return payment, err
}

func (r *invoicePaymentRepo) GetByInvoiceID(invoiceID int) ([]InvoicePayment, error) {
	rows, err := r.db.Query("SELECT id, transaction_id, reference, date, invoice_id, user_id, account_id, amount, status, createdAt, updatedAt FROM invoice_payments WHERE invoice_id = ? ORDER BY createdAt DESC", invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	payments := []InvoicePayment{}
	for rows.Next() {
		var payment InvoicePayment
		err := rows.Scan(&payment.ID, &payment.TransactionID, &payment.Reference, &payment.Date, &payment.InvoiceID, &payment.UserID, &payment.AccountID, &payment.Amount, &payment.Status, &payment.CreatedAt, &payment.UpdatedAt)
		if err != nil {
			return nil, err
		}
		payments = append(payments, payment)
	}

	return payments, nil
}

func (r *invoicePaymentRepo) GetByBuildingID(buildingID int) ([]InvoicePayment, error) {
	// Join with invoices table to filter by building_id
	rows, err := r.db.Query(`
		SELECT ip.id, ip.transaction_id, ip.reference, ip.date, ip.invoice_id, ip.user_id, ip.account_id, ip.amount, ip.status, ip.createdAt, ip.updatedAt 
		FROM invoice_payments ip
		INNER JOIN invoices i ON ip.invoice_id = i.id
		WHERE i.building_id = ?
		ORDER BY ip.createdAt DESC
	`, buildingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	payments := []InvoicePayment{}
	for rows.Next() {
		var payment InvoicePayment
		err := rows.Scan(&payment.ID, &payment.TransactionID, &payment.Reference, &payment.Date, &payment.InvoiceID, &payment.UserID, &payment.AccountID, &payment.Amount, &payment.Status, &payment.CreatedAt, &payment.UpdatedAt)
		if err != nil {
			return nil, err
		}
		payments = append(payments, payment)
	}

	return payments, nil
}

