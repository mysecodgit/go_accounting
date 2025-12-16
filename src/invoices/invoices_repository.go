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
	GetByBuildingIDWithTotalsAndFilters(buildingID int, startDate, endDate *string, peopleID *int, status *string) ([]InvoiceListItem, error)
	GetNextInvoiceNo(buildingID int) (string, error)
	CheckDuplicateInvoiceNo(buildingID int, invoiceNo string, excludeID int) (bool, error)
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

	result, err := r.db.Exec("INSERT INTO invoices (invoice_no, transaction_id, sales_date, due_date, ar_account_id, unit_id, people_id, user_id, amount, description, cancel_reason, status, building_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		invoice.InvoiceNo, invoice.TransactionID, invoice.SalesDate, invoice.DueDate, arAccountID, unitID, peopleID, invoice.UserID, invoice.Amount, invoice.Description, cancelReason, invoice.Status, invoice.BuildingID)

	if err != nil {
		return invoice, err
	}

	id, _ := result.LastInsertId()
	invoice.ID = int(id)

	err = r.db.QueryRow("SELECT id, invoice_no, transaction_id, sales_date, due_date, ar_account_id, unit_id, people_id, user_id, amount, description, cancel_reason, status, building_id, createdAt, updatedAt FROM invoices WHERE id = ?", invoice.ID).
		Scan(&invoice.ID, &invoice.InvoiceNo, &invoice.TransactionID, &invoice.SalesDate, &invoice.DueDate, &invoice.ARAccountID, &invoice.UnitID, &invoice.PeopleID, &invoice.UserID, &invoice.Amount, &invoice.Description, &invoice.CancelReason, &invoice.Status, &invoice.BuildingID, &invoice.CreatedAt, &invoice.UpdatedAt)

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

	_, err := r.db.Exec("UPDATE invoices SET invoice_no = ?, sales_date = ?, due_date = ?, ar_account_id = ?, unit_id = ?, people_id = ?, amount = ?, description = ?, cancel_reason = ? WHERE id = ?",
		invoice.InvoiceNo, invoice.SalesDate, invoice.DueDate, arAccountID, unitID, peopleID, invoice.Amount, invoice.Description, cancelReason, invoice.ID)

	if err != nil {
		return invoice, err
	}

	err = r.db.QueryRow("SELECT id, invoice_no, transaction_id, sales_date, due_date, ar_account_id, unit_id, people_id, user_id, amount, description, cancel_reason, status, building_id, createdAt, updatedAt FROM invoices WHERE id = ?", invoice.ID).
		Scan(&invoice.ID, &invoice.InvoiceNo, &invoice.TransactionID, &invoice.SalesDate, &invoice.DueDate, &invoice.ARAccountID, &invoice.UnitID, &invoice.PeopleID, &invoice.UserID, &invoice.Amount, &invoice.Description, &invoice.CancelReason, &invoice.Status, &invoice.BuildingID, &invoice.CreatedAt, &invoice.UpdatedAt)

	return invoice, err
}

func (r *invoiceRepo) GetByID(id int) (Invoice, error) {
	var invoice Invoice
	err := r.db.QueryRow("SELECT id, invoice_no, transaction_id, sales_date, due_date, ar_account_id, unit_id, people_id, user_id, amount, description, cancel_reason, status, building_id, createdAt, updatedAt FROM invoices WHERE id = ?", id).
		Scan(&invoice.ID, &invoice.InvoiceNo, &invoice.TransactionID, &invoice.SalesDate, &invoice.DueDate, &invoice.ARAccountID, &invoice.UnitID, &invoice.PeopleID, &invoice.UserID, &invoice.Amount, &invoice.Description, &invoice.CancelReason, &invoice.Status, &invoice.BuildingID, &invoice.CreatedAt, &invoice.UpdatedAt)

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
		err := rows.Scan(&invoice.ID, &invoice.InvoiceNo, &invoice.TransactionID, &invoice.SalesDate, &invoice.DueDate, &invoice.ARAccountID, &invoice.UnitID, &invoice.PeopleID, &invoice.UserID, &invoice.Amount, &invoice.Description, &invoice.CancelReason, &invoice.Status, &invoice.BuildingID, &invoice.CreatedAt, &invoice.UpdatedAt)
		if err != nil {
			return nil, err
		}
		invoices = append(invoices, invoice)
	}

	return invoices, nil
}

func (r *invoiceRepo) GetNextInvoiceNo(buildingID int) (string, error) {
	var maxNo sql.NullString
	err := r.db.QueryRow("SELECT MAX(CAST(invoice_no AS UNSIGNED)) FROM invoices WHERE building_id = ? AND invoice_no REGEXP '^[0-9]+$'", buildingID).Scan(&maxNo)
	if err != nil {
		return "1", err
	}
	if maxNo.Valid {
		// Try to parse as int and increment
		var nextNum int
		fmt.Sscanf(maxNo.String, "%d", &nextNum)
		return fmt.Sprintf("%d", nextNum+1), nil
	}
	return "1", nil
}

func (r *invoiceRepo) CheckDuplicateInvoiceNo(buildingID int, invoiceNo string, excludeID int) (bool, error) {
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
	return r.GetByBuildingIDWithTotalsAndFilters(buildingID, nil, nil, nil, nil)
}

func (r *invoiceRepo) GetByBuildingIDWithTotalsAndFilters(buildingID int, startDate, endDate *string, peopleID *int, status *string) ([]InvoiceListItem, error) {
	query := `
		SELECT 
			i.id, i.invoice_no, i.transaction_id, i.sales_date, i.due_date, 
			i.ar_account_id, i.unit_id, i.people_id, i.user_id, i.amount, 
			i.description, i.cancel_reason, i.status, i.building_id, 
			i.createdAt, i.updatedAt,
			COALESCE((
				SELECT SUM(ip.amount) 
				FROM invoice_payments ip 
				WHERE ip.invoice_id = i.id AND ip.status = '1'
			), 0) as paid_amount,
			COALESCE((
				SELECT SUM(iac.amount) 
				FROM invoice_applied_credits iac 
				WHERE iac.invoice_id = i.id AND iac.status = '1'
			), 0) as applied_credits_total
		FROM invoices i
		WHERE i.building_id = ?
	`
	
	args := []interface{}{buildingID}
	
	// Add filters
	if startDate != nil && *startDate != "" {
		query += " AND DATE(i.sales_date) >= ?"
		args = append(args, *startDate)
	}
	
	if endDate != nil && *endDate != "" {
		query += " AND DATE(i.sales_date) <= ?"
		args = append(args, *endDate)
	}
	
	if peopleID != nil && *peopleID > 0 {
		query += " AND i.people_id = ?"
		args = append(args, *peopleID)
	}
	
	if status != nil && *status != "" {
		query += " AND i.status = ?"
		args = append(args, *status)
	}
	
	query += " ORDER BY i.createdAt DESC"
	
	rows, err := r.db.Query(query, args...)
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
			&invoice.Description, &invoice.CancelReason, &invoice.Status, &invoice.BuildingID,
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
