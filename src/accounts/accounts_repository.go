package accounts

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/mysecodgit/go_accounting/src/account_types"
	"github.com/mysecodgit/go_accounting/src/building"
)

type AccountRepository interface {
	Create(account Account) (Account, error)
	Update(account Account, id int) (Account, error)
	GetByID(id int) (Account, account_types.AccountType, building.Building, error)
	GetAll() ([]Account, []account_types.AccountType, []building.Building, error)
	AccountTypeIDExists(accountTypeID int) (bool, error)
	BuildingIDExists(buildingID int) (bool, error)
	CheckDuplicateAccountNumber(buildingID int, accountNumber int, excludeID int) (bool, error)
	CheckDuplicateAccountName(buildingID int, accountName string, excludeID int) (bool, error)
}

type accountRepo struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) AccountRepository {
	return &accountRepo{db: db}
}

func (r *accountRepo) Create(account Account) (Account, error) {
	// Check for duplicate account number before inserting
	exists, err := r.CheckDuplicateAccountNumber(account.BuildingID, account.AccountNumber, 0)
	if err != nil {
		return account, err
	}
	if exists {
		return account, fmt.Errorf("building cannot have duplicate account number")
	}

	// Check for duplicate account name before inserting
	exists, err = r.CheckDuplicateAccountName(account.BuildingID, account.AccountName, 0)
	if err != nil {
		return account, err
	}
	if exists {
		return account, fmt.Errorf("building cannot have duplicate account name")
	}

	result, err := r.db.Exec("INSERT INTO accounts (account_number, account_name, account_type, building_id, isDefault) VALUES (?, ?, ?, ?, ?)",
		account.AccountNumber, account.AccountName, account.AccountType, account.BuildingID, account.IsDefault)

	if err != nil {
		// Check if it's a duplicate key error
		if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "UNIQUE constraint") {
			if strings.Contains(err.Error(), "account_number") {
				return account, fmt.Errorf("building cannot have duplicate account number")
			}
			if strings.Contains(err.Error(), "account_name") {
				return account, fmt.Errorf("building cannot have duplicate account name")
			}
			return account, fmt.Errorf("duplicate entry detected")
		}
		return account, err
	}

	id, _ := result.LastInsertId()
	account.ID = int(id)

	// Fetch the created record to get created_at and updated_at
	err = r.db.QueryRow("SELECT id, account_number, account_name, account_type, building_id, isDefault, created_at, updated_at FROM accounts WHERE id = ?", account.ID).
		Scan(&account.ID, &account.AccountNumber, &account.AccountName, &account.AccountType, &account.BuildingID, &account.IsDefault, &account.CreatedAt, &account.UpdatedAt)

	return account, err
}

func (r *accountRepo) Update(account Account, id int) (Account, error) {
	// Check for duplicate account name before updating (excluding current account)
	exists, err := r.CheckDuplicateAccountName(account.BuildingID, account.AccountName, id)
	if err != nil {
		return account, err
	}
	if exists {
		return account, fmt.Errorf("building cannot have duplicate account name")
	}

	_, err = r.db.Exec("UPDATE accounts SET account_number=?, account_name=?, account_type=?, building_id=?, isDefault=?, updated_at=NOW() WHERE id=?",
		account.AccountNumber, account.AccountName, account.AccountType, account.BuildingID, account.IsDefault, id)

	if err != nil {
		// Check if it's a duplicate key error for account_name
		if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "UNIQUE constraint") {
			if strings.Contains(err.Error(), "account_name") {
				return account, fmt.Errorf("building cannot have duplicate account name")
			}
			return account, fmt.Errorf("duplicate entry detected")
		}
		return account, err
	}

	account.ID = id

	// Fetch the updated record to get created_at and updated_at
	err = r.db.QueryRow("SELECT id, account_number, account_name, account_type, building_id, isDefault, created_at, updated_at FROM accounts WHERE id = ?", id).
		Scan(&account.ID, &account.AccountNumber, &account.AccountName, &account.AccountType, &account.BuildingID, &account.IsDefault, &account.CreatedAt, &account.UpdatedAt)

	return account, err
}

func (r *accountRepo) GetByID(id int) (Account, account_types.AccountType, building.Building, error) {
	var account Account
	var accountType account_types.AccountType
	var b building.Building
	err := r.db.QueryRow("SELECT a.id, a.account_number, a.account_name, a.account_type, a.building_id, a.isDefault, a.created_at, a.updated_at, "+
		"at.id, at.typeName, at.`type`, at.sub_type, at.typeStatus, at.created_at, at.updated_at, "+
		"b.id, b.name, b.created_at, b.updated_at "+
		"FROM accounts a "+
		"INNER JOIN account_types at ON a.account_type = at.id "+
		"INNER JOIN buildings b ON a.building_id = b.id "+
		"WHERE a.id = ?", id).
		Scan(&account.ID, &account.AccountNumber, &account.AccountName, &account.AccountType, &account.BuildingID, &account.IsDefault, &account.CreatedAt, &account.UpdatedAt,
			&accountType.ID, &accountType.TypeName, &accountType.Type, &accountType.SubType, &accountType.TypeStatus, &accountType.CreatedAt, &accountType.UpdatedAt,
			&b.ID, &b.Name, &b.CreatedAt, &b.UpdatedAt)

	if err == sql.ErrNoRows {
		return account, accountType, b, fmt.Errorf("id does not exist")
	}

	return account, accountType, b, err
}

func (r *accountRepo) GetAll() ([]Account, []account_types.AccountType, []building.Building, error) {
	rows, err := r.db.Query("SELECT a.id, a.account_number, a.account_name, a.account_type, a.building_id, a.isDefault, a.created_at, a.updated_at, " +
		"at.id, at.typeName, at.`type`, at.sub_type, at.typeStatus, at.created_at, at.updated_at, " +
		"b.id, b.name, b.created_at, b.updated_at " +
		"FROM accounts a " +
		"INNER JOIN account_types at ON a.account_type = at.id " +
		"INNER JOIN buildings b ON a.building_id = b.id " +
		"ORDER BY a.created_at DESC")
	if err != nil {
		return nil, nil, nil, err
	}
	defer rows.Close()

	accounts := []Account{}
	accountTypes := []account_types.AccountType{}
	buildings := []building.Building{}
	for rows.Next() {
		var a Account
		var at account_types.AccountType
		var b building.Building
		err := rows.Scan(&a.ID, &a.AccountNumber, &a.AccountName, &a.AccountType, &a.BuildingID, &a.IsDefault, &a.CreatedAt, &a.UpdatedAt,
			&at.ID, &at.TypeName, &at.Type, &at.SubType, &at.TypeStatus, &at.CreatedAt, &at.UpdatedAt,
			&b.ID, &b.Name, &b.CreatedAt, &b.UpdatedAt)
		if err != nil {
			return nil, nil, nil, err
		}
		accounts = append(accounts, a)
		accountTypes = append(accountTypes, at)
		buildings = append(buildings, b)
	}
	return accounts, accountTypes, buildings, nil
}

func (r *accountRepo) AccountTypeIDExists(accountTypeID int) (bool, error) {
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM account_types WHERE id = ?)", accountTypeID).Scan(&exists)
	return exists, err
}

func (r *accountRepo) BuildingIDExists(buildingID int) (bool, error) {
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM buildings WHERE id = ?)", buildingID).Scan(&exists)
	return exists, err
}

func (r *accountRepo) CheckDuplicateAccountNumber(buildingID int, accountNumber int, excludeID int) (bool, error) {
	var count int
	var err error
	
	if excludeID > 0 {
		err = r.db.QueryRow(
			"SELECT COUNT(*) FROM accounts WHERE building_id = ? AND account_number = ? AND id != ?",
			buildingID, accountNumber, excludeID).Scan(&count)
	} else {
		err = r.db.QueryRow(
			"SELECT COUNT(*) FROM accounts WHERE building_id = ? AND account_number = ?",
			buildingID, accountNumber).Scan(&count)
	}
	
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}

func (r *accountRepo) CheckDuplicateAccountName(buildingID int, accountName string, excludeID int) (bool, error) {
	var count int
	var err error
	
	if excludeID > 0 {
		err = r.db.QueryRow(
			"SELECT COUNT(*) FROM accounts WHERE building_id = ? AND account_name = ? AND id != ?",
			buildingID, accountName, excludeID).Scan(&count)
	} else {
		err = r.db.QueryRow(
			"SELECT COUNT(*) FROM accounts WHERE building_id = ? AND account_name = ?",
			buildingID, accountName).Scan(&count)
	}
	
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}

