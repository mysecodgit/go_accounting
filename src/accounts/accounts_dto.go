package accounts

import (
	"github.com/mysecodgit/go_accounting/src/account_types"
	"github.com/mysecodgit/go_accounting/src/building"
)

type AccountResponse struct {
	ID            int                     `json:"id"`
	AccountNumber int                     `json:"account_number"`
	AccountName   string                  `json:"account_name"`
	AccountType   account_types.AccountType `json:"account_type"`
	Building      building.Building       `json:"building"`
	IsDefault     int                     `json:"isDefault"`
	CreatedAt     string                  `json:"created_at"`
	UpdatedAt     string                  `json:"updated_at"`
}

func (a *Account) ToAccountResponse(accountType account_types.AccountType, b building.Building) AccountResponse {
	return AccountResponse{
		ID:            a.ID,
		AccountNumber: a.AccountNumber,
		AccountName:   a.AccountName,
		AccountType:   accountType,
		Building:      b,
		IsDefault:     a.IsDefault,
		CreatedAt:     a.CreatedAt,
		UpdatedAt:     a.UpdatedAt,
	}
}



