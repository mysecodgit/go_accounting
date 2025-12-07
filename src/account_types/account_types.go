package account_types

import "strings"

type AccountType struct {
	ID         int    `json:"id"`
	TypeName   string `json:"typeName"`
	Type       string `json:"type"`
	SubType    string `json:"sub_type"`
	TypeStatus string `json:"typeStatus"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

func (a *AccountType) Validate() map[string]string {
	errors := make(map[string]string)

	if strings.TrimSpace(a.TypeName) == "" {
		errors["typeName"] = "Type name cannot be empty"
	}

	if strings.TrimSpace(a.Type) == "" {
		errors["type"] = "Type cannot be empty"
	}

	if strings.TrimSpace(a.SubType) == "" {
		errors["sub_type"] = "Sub type cannot be empty"
	}

	if strings.TrimSpace(a.TypeStatus) == "" {
		errors["typeStatus"] = "Type status cannot be empty"
	}

	if len(errors) == 0 {
		return nil
	}

	return errors
}
