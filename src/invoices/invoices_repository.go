package invoices

import (
	"database/sql"
	"fmt"
)

type InvoiceRepository interface {
	Create(invoice Invoice) (Invoice, error)
	Update(invoice Invoice) (Invoice, error)
	GetByID(id int) (Invoice, error)
	GetByBuildingID(buildingID int) ([]Invoice, error)
	GetByBuildingIDWithTotals(buildingID int) ([]InvoiceListItem, error)
	GetNextInvoiceNo(buildingID int) (int, error)
	CheckDuplicateInvoiceNo(buildingID int, invoiceNo int, excludeID int) (bool, error)
}

type invoiceRepo struct {
	db *sql.DB
}

func NewInvoiceRepository(db *sql.DB) InvoiceRepository {
	return &invoiceRepo{db: db}
}

func (r *invoiceRepo) Create(invoice Invoice) (Invoice, error) {
	var unitID interface{}
	if invoice.UnitID != nil {
		unitID = *invoice.UnitID
	} else {
		unitID = nil
	}

	var peopleID interface{}
	if invoice.PeopleID != nil {
		peopleID = *invoice.PeopleID
	} else {
		peopleID = nil
	}

	var cancelReason interface{}
	if invoice.CancelReason != nil {
		cancelReason = *invoice.CancelReason
	} else {
		cancelReason = nil
	}

	var arAccountID interface{}
	if invoice.ARAccountID != nil {
		arAccountID = *invoice.ARAccountID
	} else {
		arAccountID = nil
	}

	result, err := r.db.Exec("INSERT INTO invoices (invoice_no, transaction_id, sales_date, due_date, ar_account_id, unit_id, people_id, user_id, amount, description, refrence, cancel_reason, status, building_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		invoice.InvoiceNo, invoice.TransactionID, invoice.SalesDate, invoice.DueDate, arAccountID, unitID, peopleID, invoice.UserID, invoice.Amount, invoice.Description, invoice.Reference, cancelReason, invoice.Status, invoice.BuildingID)

	if err != nil {
		return invoice, err
	}

	id, _ := result.LastInsertId()
	invoice.ID = int(id)

	err = r.db.QueryRow("SELECT id, invoice_no, transaction_id, sales_date, due_date, ar_account_id, unit_id, people_id, user_id, amount, description, refrence, cancel_reason, status, building_id, createdAt, updatedAt FROM invoices WHERE id = ?", invoice.ID).
		Scan(&invoice.ID, &invoice.InvoiceNo, &invoice.TransactionID, &invoice.SalesDate, &invoice.DueDate, &invoice.ARAccountID, &invoice.UnitID, &invoice.PeopleID, &invoice.UserID, &invoice.Amount, &invoice.Description, &invoice.Reference, &invoice.CancelReason, &invoice.Status, &invoice.BuildingID, &invoice.CreatedAt, &invoice.UpdatedAt)

	return invoice, err
}

func (r *invoiceRepo) Update(invoice Invoice) (Invoice, error) {
	var unitID interface{}
	if invoice.UnitID != nil {
		unitID = *invoice.UnitID
	} else {
		unitID = nil
	}

	var peopleID interface{}
	if invoice.PeopleID != nil {
		peopleID = *invoice.PeopleID
	} else {
		peopleID = nil
	}

	var cancelReason interface{}
	if invoice.CancelReason != nil {
		cancelReason = *invoice.CancelReason
	} else {
		cancelReason = nil
	}

	var arAccountID interface{}
	if invoice.ARAccountID != nil {
		arAccountID = *invoice.ARAccountID
	} else {
		arAccountID = nil
	}

	_, err := r.db.Exec("UPDATE invoices SET invoice_no = ?, sales_date = ?, due_date = ?, ar_account_id = ?, unit_id = ?, people_id = ?, amount = ?, description = ?, refrence = ?, cancel_reason = ? WHERE id = ?",
		invoice.InvoiceNo, invoice.SalesDate, invoice.DueDate, arAccountID, unitID, peopleID, invoice.Amount, invoice.Description, invoice.Reference, cancelReason, invoice.ID)

	if err != nil {
		return invoice, err
	}

	err = r.db.QueryRow("SELECT id, invoice_no, transaction_id, sales_date, due_date, ar_account_id, unit_id, people_id, user_id, amount, description, refrence, cancel_reason, status, building_id, createdAt, updatedAt FROM invoices WHERE id = ?", invoice.ID).
		Scan(&invoice.ID, &invoice.InvoiceNo, &invoice.TransactionID, &invoice.SalesDate, &invoice.DueDate, &invoice.ARAccountID, &invoice.UnitID, &invoice.PeopleID, &invoice.UserID, &invoice.Amount, &invoice.Description, &invoice.Reference, &invoice.CancelReason, &invoice.Status, &invoice.BuildingID, &invoice.CreatedAt, &invoice.UpdatedAt)

	return invoice, err
}

func (r *invoiceRepo) GetByID(id int) (Invoice, error) {
	var invoice Invoice
	err := r.db.QueryRow("SELECT id, invoice_no, transaction_id, sales_date, due_date, ar_account_id, unit_id, people_id, user_id, amount, description, refrence, cancel_reason, status, building_id, createdAt, updatedAt FROM invoices WHERE id = ?", id).
		Scan(&invoice.ID, &invoice.InvoiceNo, &invoice.TransactionID, &invoice.SalesDate, &invoice.DueDate, &invoice.ARAccountID, &invoice.UnitID, &invoice.PeopleID, &invoice.UserID, &invoice.Amount, &invoice.Description, &invoice.Reference, &invoice.CancelReason, &invoice.Status, &invoice.BuildingID, &invoice.CreatedAt, &invoice.UpdatedAt)

	if err == sql.ErrNoRows {
		return invoice, fmt.Errorf("invoice not found")
	}

	return invoice, err
}

func (r *invoiceRepo) GetByBuildingID(buildingID int) ([]Invoice, error) {
	rows, err := r.db.Query("SELECT id, invoice_no, transaction_id, sales_date, due_date, ar_account_id, unit_id, people_id, user_id, amount, description, refrence, cancel_reason, status, building_id, createdAt, updatedAt FROM invoices WHERE building_id = ? ORDER BY createdAt DESC", buildingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	invoices := []Invoice{}
	for rows.Next() {
		var invoice Invoice
		err := rows.Scan(&invoice.ID, &invoice.InvoiceNo, &invoice.TransactionID, &invoice.SalesDate, &invoice.DueDate, &invoice.ARAccountID, &invoice.UnitID, &invoice.PeopleID, &invoice.UserID, &invoice.Amount, &invoice.Description, &invoice.Reference, &invoice.CancelReason, &invoice.Status, &invoice.BuildingID, &invoice.CreatedAt, &invoice.UpdatedAt)
		if err != nil {
			return nil, err
		}
		invoices = append(invoices, invoice)
	}

	return invoices, nil
}

func (r *invoiceRepo) GetNextInvoiceNo(buildingID int) (int, error) {
	var maxNo sql.NullInt64
	err := r.db.QueryRow("SELECT MAX(invoice_no) FROM invoices WHERE building_id = ?", buildingID).Scan(&maxNo)
	if err != nil {
		return 1, err
	}
	if maxNo.Valid {
		return int(maxNo.Int64) + 1, nil
	}
	return 1, nil
}

func (r *invoiceRepo) CheckDuplicateInvoiceNo(buildingID int, invoiceNo int, excludeID int) (bool, error) {
	var count int
	var err error

	if excludeID > 0 {
		err = r.db.QueryRow("SELECT COUNT(*) FROM invoices WHERE building_id = ? AND invoice_no = ? AND id != ?", buildingID, invoiceNo, excludeID).Scan(&count)
	} else {
		err = r.db.QueryRow("SELECT COUNT(*) FROM invoices WHERE building_id = ? AND invoice_no = ?", buildingID, invoiceNo).Scan(&count)
	}

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *invoiceRepo) GetByBuildingIDWithTotals(buildingID int) ([]InvoiceListItem, error) {
	query := `
		SELECT 
			i.id, i.invoice_no, i.transaction_id, i.sales_date, i.due_date, 
			i.ar_account_id, i.unit_id, i.people_id, i.user_id, i.amount, 
			i.description, i.refrence, i.cancel_reason, i.status, i.building_id, 
			i.createdAt, i.updatedAt,
			COALESCE(SUM(CASE WHEN ip.status = '1' THEN ip.amount ELSE 0 END), 0) as paid_amount,
			COALESCE(SUM(CASE WHEN iac.status = '1' THEN iac.amount ELSE 0 END), 0) as applied_credits_total
		FROM invoices i
		LEFT JOIN invoice_payments ip ON i.id = ip.invoice_id
		LEFT JOIN invoice_applied_credits iac ON i.id = iac.invoice_id
		WHERE i.building_id = ?
		GROUP BY i.id, i.invoice_no, i.transaction_id, i.sales_date, i.due_date, 
			i.ar_account_id, i.unit_id, i.people_id, i.user_id, i.amount, 
			i.description, i.refrence, i.cancel_reason, i.status, i.building_id, 
			i.createdAt, i.updatedAt
		ORDER BY i.createdAt DESC
	`
	
	rows, err := r.db.Query(query, buildingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	invoices := []InvoiceListItem{}
	for rows.Next() {
		var invoice InvoiceListItem
		err := rows.Scan(
			&invoice.ID, &invoice.InvoiceNo, &invoice.TransactionID, &invoice.SalesDate, &invoice.DueDate,
			&invoice.ARAccountID, &invoice.UnitID, &invoice.PeopleID, &invoice.UserID, &invoice.Amount,
			&invoice.Description, &invoice.Reference, &invoice.CancelReason, &invoice.Status, &invoice.BuildingID,
			&invoice.CreatedAt, &invoice.UpdatedAt,
			&invoice.PaidAmount, &invoice.AppliedCreditsTotal,
		)
		if err != nil {
			return nil, err
		}
		invoices = append(invoices, invoice)
	}

	return invoices, nil
}
