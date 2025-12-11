package items

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/mysecodgit/go_accounting/src/accounts"
	"github.com/mysecodgit/go_accounting/src/building"
)

type ItemRepository interface {
	Create(item Item) (Item, error)
	Update(item Item, id int) (Item, error)
	GetByID(id int) (Item, building.Building, *accounts.Account, *accounts.Account, *accounts.Account, *accounts.Account, error)
	GetAll() ([]Item, []building.Building, []*accounts.Account, []*accounts.Account, []*accounts.Account, []*accounts.Account, error)
	GetByBuildingID(buildingID int) ([]Item, []building.Building, []*accounts.Account, []*accounts.Account, []*accounts.Account, []*accounts.Account, error)
	BuildingIDExists(buildingID int) (bool, error)
	AccountIDExists(accountID int) (bool, error)
}

type itemRepo struct {
	db *sql.DB
}

func NewItemRepository(db *sql.DB) ItemRepository {
	return &itemRepo{db: db}
}

func (r *itemRepo) Create(item Item) (Item, error) {
	result, err := r.db.Exec("INSERT INTO items (name, type, description, asset_account, income_account, cogs_account, expense_account, on_hand, avg_cost, date, building_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		item.Name, item.Type, item.Description, item.AssetAccount, item.IncomeAccount, item.COGSAccount, item.ExpenseAccount, item.OnHand, item.AvgCost, item.Date, item.BuildingID)

	if err != nil {
		if strings.Contains(err.Error(), "foreign key constraint") {
			return item, fmt.Errorf("invalid foreign key reference")
		}
		return item, err
	}

	id, _ := result.LastInsertId()
	item.ID = int(id)

	// Fetch the created record to get created_at and updated_at
	err = r.db.QueryRow("SELECT id, name, type, description, asset_account, income_account, cogs_account, expense_account, on_hand, avg_cost, date, building_id, created_at, updated_at FROM items WHERE id = ?", item.ID).
		Scan(&item.ID, &item.Name, &item.Type, &item.Description, &item.AssetAccount, &item.IncomeAccount, &item.COGSAccount, &item.ExpenseAccount, &item.OnHand, &item.AvgCost, &item.Date, &item.BuildingID, &item.CreatedAt, &item.UpdatedAt)

	return item, err
}

func (r *itemRepo) Update(item Item, id int) (Item, error) {
	_, err := r.db.Exec("UPDATE items SET name=?, type=?, description=?, asset_account=?, income_account=?, cogs_account=?, expense_account=?, on_hand=?, avg_cost=?, date=?, building_id=?, updated_at=NOW() WHERE id=?",
		item.Name, item.Type, item.Description, item.AssetAccount, item.IncomeAccount, item.COGSAccount, item.ExpenseAccount, item.OnHand, item.AvgCost, item.Date, item.BuildingID, id)

	if err != nil {
		if strings.Contains(err.Error(), "foreign key constraint") {
			return item, fmt.Errorf("invalid foreign key reference")
		}
		return item, err
	}

	item.ID = id

	// Fetch the updated record to get created_at and updated_at
	err = r.db.QueryRow("SELECT id, name, type, description, asset_account, income_account, cogs_account, expense_account, on_hand, avg_cost, date, building_id, created_at, updated_at FROM items WHERE id = ?", id).
		Scan(&item.ID, &item.Name, &item.Type, &item.Description, &item.AssetAccount, &item.IncomeAccount, &item.COGSAccount, &item.ExpenseAccount, &item.OnHand, &item.AvgCost, &item.Date, &item.BuildingID, &item.CreatedAt, &item.UpdatedAt)

	return item, err
}

func (r *itemRepo) GetByID(id int) (Item, building.Building, *accounts.Account, *accounts.Account, *accounts.Account, *accounts.Account, error) {
	var item Item
	var b building.Building
	var assetAccountID, incomeAccountID, cogsAccountID, expenseAccountID sql.NullInt64
	var assetAccount, incomeAccount, cogsAccount, expenseAccount *accounts.Account

	err := r.db.QueryRow("SELECT i.id, i.name, i.type, i.description, i.asset_account, i.income_account, i.cogs_account, i.expense_account, i.on_hand, i.avg_cost, i.date, i.building_id, i.created_at, i.updated_at, "+
		"b.id, b.name, b.created_at, b.updated_at "+
		"FROM items i "+
		"INNER JOIN buildings b ON i.building_id = b.id "+
		"WHERE i.id = ?", id).
		Scan(&item.ID, &item.Name, &item.Type, &item.Description, &assetAccountID, &incomeAccountID, &cogsAccountID, &expenseAccountID, &item.OnHand, &item.AvgCost, &item.Date, &item.BuildingID, &item.CreatedAt, &item.UpdatedAt,
			&b.ID, &b.Name, &b.CreatedAt, &b.UpdatedAt)

	if err == sql.ErrNoRows {
		return item, b, nil, nil, nil, nil, fmt.Errorf("id does not exist")
	}
	if err != nil {
		return item, b, nil, nil, nil, nil, err
	}

	// Fetch account details if they exist
	if assetAccountID.Valid {
		acc, err := r.getAccountByID(int(assetAccountID.Int64))
		if err == nil {
			assetAccount = acc
		}
	}
	if incomeAccountID.Valid {
		acc, err := r.getAccountByID(int(incomeAccountID.Int64))
		if err == nil {
			incomeAccount = acc
		}
	}
	if cogsAccountID.Valid {
		acc, err := r.getAccountByID(int(cogsAccountID.Int64))
		if err == nil {
			cogsAccount = acc
		}
	}
	if expenseAccountID.Valid {
		acc, err := r.getAccountByID(int(expenseAccountID.Int64))
		if err == nil {
			expenseAccount = acc
		}
	}

	return item, b, assetAccount, incomeAccount, cogsAccount, expenseAccount, nil
}

func (r *itemRepo) GetAll() ([]Item, []building.Building, []*accounts.Account, []*accounts.Account, []*accounts.Account, []*accounts.Account, error) {
	rows, err := r.db.Query("SELECT i.id, i.name, i.type, i.description, i.asset_account, i.income_account, i.cogs_account, i.expense_account, i.on_hand, i.avg_cost, i.date, i.building_id, i.created_at, i.updated_at, " +
		"b.id, b.name, b.created_at, b.updated_at " +
		"FROM items i " +
		"INNER JOIN buildings b ON i.building_id = b.id " +
		"ORDER BY i.created_at DESC")
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	defer rows.Close()

	items := []Item{}
	buildings := []building.Building{}
	assetAccounts := []*accounts.Account{}
	incomeAccounts := []*accounts.Account{}
	cogsAccounts := []*accounts.Account{}
	expenseAccounts := []*accounts.Account{}

	for rows.Next() {
		var item Item
		var b building.Building
		var assetAccountID, incomeAccountID, cogsAccountID, expenseAccountID sql.NullInt64

		err := rows.Scan(&item.ID, &item.Name, &item.Type, &item.Description, &assetAccountID, &incomeAccountID, &cogsAccountID, &expenseAccountID, &item.OnHand, &item.AvgCost, &item.Date, &item.BuildingID, &item.CreatedAt, &item.UpdatedAt,
			&b.ID, &b.Name, &b.CreatedAt, &b.UpdatedAt)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, err
		}

		items = append(items, item)
		buildings = append(buildings, b)

		// Fetch account details if they exist
		var assetAccount, incomeAccount, cogsAccount, expenseAccount *accounts.Account
		if assetAccountID.Valid {
			acc, err := r.getAccountByID(int(assetAccountID.Int64))
			if err == nil {
				assetAccount = acc
			}
		}
		if incomeAccountID.Valid {
			acc, err := r.getAccountByID(int(incomeAccountID.Int64))
			if err == nil {
				incomeAccount = acc
			}
		}
		if cogsAccountID.Valid {
			acc, err := r.getAccountByID(int(cogsAccountID.Int64))
			if err == nil {
				cogsAccount = acc
			}
		}
		if expenseAccountID.Valid {
			acc, err := r.getAccountByID(int(expenseAccountID.Int64))
			if err == nil {
				expenseAccount = acc
			}
		}

		assetAccounts = append(assetAccounts, assetAccount)
		incomeAccounts = append(incomeAccounts, incomeAccount)
		cogsAccounts = append(cogsAccounts, cogsAccount)
		expenseAccounts = append(expenseAccounts, expenseAccount)
	}

	return items, buildings, assetAccounts, incomeAccounts, cogsAccounts, expenseAccounts, nil
}

func (r *itemRepo) GetByBuildingID(buildingID int) ([]Item, []building.Building, []*accounts.Account, []*accounts.Account, []*accounts.Account, []*accounts.Account, error) {
	rows, err := r.db.Query("SELECT i.id, i.name, i.type, i.description, i.asset_account, i.income_account, i.cogs_account, i.expense_account, i.on_hand, i.avg_cost, i.date, i.building_id, i.created_at, i.updated_at, "+
		"b.id, b.name, b.created_at, b.updated_at "+
		"FROM items i "+
		"INNER JOIN buildings b ON i.building_id = b.id "+
		"WHERE i.building_id = ? "+
		"ORDER BY i.created_at DESC", buildingID)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	defer rows.Close()

	items := []Item{}
	buildings := []building.Building{}
	assetAccounts := []*accounts.Account{}
	incomeAccounts := []*accounts.Account{}
	cogsAccounts := []*accounts.Account{}
	expenseAccounts := []*accounts.Account{}

	for rows.Next() {
		var item Item
		var b building.Building
		var assetAccountID, incomeAccountID, cogsAccountID, expenseAccountID sql.NullInt64

		err := rows.Scan(&item.ID, &item.Name, &item.Type, &item.Description, &assetAccountID, &incomeAccountID, &cogsAccountID, &expenseAccountID, &item.OnHand, &item.AvgCost, &item.Date, &item.BuildingID, &item.CreatedAt, &item.UpdatedAt,
			&b.ID, &b.Name, &b.CreatedAt, &b.UpdatedAt)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, err
		}

		items = append(items, item)
		buildings = append(buildings, b)

		// Fetch account details if they exist
		var assetAccount, incomeAccount, cogsAccount, expenseAccount *accounts.Account
		if assetAccountID.Valid {
			acc, err := r.getAccountByID(int(assetAccountID.Int64))
			if err == nil {
				assetAccount = acc
			}
		}
		if incomeAccountID.Valid {
			acc, err := r.getAccountByID(int(incomeAccountID.Int64))
			if err == nil {
				incomeAccount = acc
			}
		}
		if cogsAccountID.Valid {
			acc, err := r.getAccountByID(int(cogsAccountID.Int64))
			if err == nil {
				cogsAccount = acc
			}
		}
		if expenseAccountID.Valid {
			acc, err := r.getAccountByID(int(expenseAccountID.Int64))
			if err == nil {
				expenseAccount = acc
			}
		}

		assetAccounts = append(assetAccounts, assetAccount)
		incomeAccounts = append(incomeAccounts, incomeAccount)
		cogsAccounts = append(cogsAccounts, cogsAccount)
		expenseAccounts = append(expenseAccounts, expenseAccount)
	}

	return items, buildings, assetAccounts, incomeAccounts, cogsAccounts, expenseAccounts, nil
}

func (r *itemRepo) BuildingIDExists(buildingID int) (bool, error) {
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM buildings WHERE id = ?)", buildingID).Scan(&exists)
	return exists, err
}

func (r *itemRepo) AccountIDExists(accountID int) (bool, error) {
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM accounts WHERE id = ?)", accountID).Scan(&exists)
	return exists, err
}

func (r *itemRepo) getAccountByID(id int) (*accounts.Account, error) {
	var account accounts.Account
	err := r.db.QueryRow("SELECT id, account_number, account_name, account_type, building_id, isDefault, created_at, updated_at FROM accounts WHERE id = ?", id).
		Scan(&account.ID, &account.AccountNumber, &account.AccountName, &account.AccountType, &account.BuildingID, &account.IsDefault, &account.CreatedAt, &account.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

