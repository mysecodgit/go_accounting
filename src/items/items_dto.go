package items

import (
	"github.com/mysecodgit/go_accounting/src/accounts"
	"github.com/mysecodgit/go_accounting/src/building"
)

type ItemResponse struct {
	ID              int                  `json:"id"`
	Name            string               `json:"name"`
	Type            string               `json:"type"`
	Description     string               `json:"description"`
	AssetAccount    *accounts.Account    `json:"asset_account,omitempty"`
	IncomeAccount   *accounts.Account    `json:"income_account,omitempty"`
	COGSAccount     *accounts.Account    `json:"cogs_account,omitempty"`
	ExpenseAccount  *accounts.Account    `json:"expense_account,omitempty"`
	OnHand          float64              `json:"on_hand"`
	AvgCost         float64              `json:"avg_cost"`
	Date            string               `json:"date"`
	Building        building.Building    `json:"building"`
	CreatedAt       string               `json:"created_at"`
	UpdatedAt       string               `json:"updated_at"`
}

func (i *Item) ToItemResponse(b building.Building, assetAccount, incomeAccount, cogsAccount, expenseAccount *accounts.Account) ItemResponse {
	response := ItemResponse{
		ID:          i.ID,
		Name:        i.Name,
		Type:        i.Type,
		Description: i.Description,
		OnHand:      i.OnHand,
		AvgCost:     i.AvgCost,
		Date:        i.Date,
		Building:    b,
		CreatedAt:   i.CreatedAt,
		UpdatedAt:   i.UpdatedAt,
	}

	if assetAccount != nil {
		response.AssetAccount = assetAccount
	}
	if incomeAccount != nil {
		response.IncomeAccount = incomeAccount
	}
	if cogsAccount != nil {
		response.COGSAccount = cogsAccount
	}
	if expenseAccount != nil {
		response.ExpenseAccount = expenseAccount
	}

	return response
}

