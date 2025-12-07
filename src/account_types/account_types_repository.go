package account_types

import (
	"database/sql"
	"fmt"
)

type AccountTypeRepository interface {
	Create(accountType AccountType) (AccountType, error)
	Update(accountType AccountType, id int) (AccountType, error)
	GetByID(id int) (AccountType, error)
	GetAll() ([]AccountType, error)
}

type accountTypeRepo struct {
	db *sql.DB
}

func NewAccountTypeRepository(db *sql.DB) AccountTypeRepository {
	return &accountTypeRepo{db: db}
}

func (r *accountTypeRepo) Create(accountType AccountType) (AccountType, error) {
	result, err := r.db.Exec("INSERT INTO account_types (typeName, `type`, sub_type, typeStatus) VALUES (?, ?, ?, ?)",
		accountType.TypeName, accountType.Type, accountType.SubType, accountType.TypeStatus)

	if err != nil {
		return accountType, err
	}

	id, _ := result.LastInsertId()
	accountType.ID = int(id)

	// Fetch the created record to get created_at and updated_at
	err = r.db.QueryRow("SELECT id, typeName, `type`, sub_type, typeStatus, created_at, updated_at FROM account_types WHERE id = ?", accountType.ID).
		Scan(&accountType.ID, &accountType.TypeName, &accountType.Type, &accountType.SubType, &accountType.TypeStatus, &accountType.CreatedAt, &accountType.UpdatedAt)

	return accountType, err
}

func (r *accountTypeRepo) Update(accountType AccountType, id int) (AccountType, error) {
	_, err := r.db.Exec("UPDATE account_types SET typeName=?, `type`=?, sub_type=?, typeStatus=?, updated_at=NOW() WHERE id=?",
		accountType.TypeName, accountType.Type, accountType.SubType, accountType.TypeStatus, id)

	if err != nil {
		return accountType, err
	}

	accountType.ID = id

	// Fetch the updated record to get created_at and updated_at
	err = r.db.QueryRow("SELECT id, typeName, `type`, sub_type, typeStatus, created_at, updated_at FROM account_types WHERE id = ?", id).
		Scan(&accountType.ID, &accountType.TypeName, &accountType.Type, &accountType.SubType, &accountType.TypeStatus, &accountType.CreatedAt, &accountType.UpdatedAt)

	return accountType, err
}

func (r *accountTypeRepo) GetByID(id int) (AccountType, error) {
	var accountType AccountType
	err := r.db.QueryRow("SELECT id, typeName, `type`, sub_type, typeStatus, created_at, updated_at FROM account_types WHERE id = ?", id).
		Scan(&accountType.ID, &accountType.TypeName, &accountType.Type, &accountType.SubType, &accountType.TypeStatus, &accountType.CreatedAt, &accountType.UpdatedAt)

	if err == sql.ErrNoRows {
		return accountType, fmt.Errorf("id does not exist")
	}

	return accountType, err
}

func (r *accountTypeRepo) GetAll() ([]AccountType, error) {
	rows, err := r.db.Query("SELECT id, typeName, `type`, sub_type, typeStatus, created_at, updated_at FROM account_types ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accountTypes := []AccountType{}
	for rows.Next() {
		var a AccountType
		err := rows.Scan(&a.ID, &a.TypeName, &a.Type, &a.SubType, &a.TypeStatus, &a.CreatedAt, &a.UpdatedAt)
		if err != nil {
			return nil, err
		}
		accountTypes = append(accountTypes, a)
	}
	return accountTypes, nil
}

