package account_types

type AccountTypeResponse struct {
	ID         int    `json:"id"`
	TypeName   string `json:"typeName"`
	Type       string `json:"type"`
	SubType    string `json:"sub_type"`
	TypeStatus string `json:"typeStatus"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

func (a *AccountType) ToAccountTypeResponse() AccountTypeResponse {
	return AccountTypeResponse{
		ID:         a.ID,
		TypeName:   a.TypeName,
		Type:       a.Type,
		SubType:    a.SubType,
		TypeStatus: a.TypeStatus,
		CreatedAt:  a.CreatedAt,
		UpdatedAt:  a.UpdatedAt,
	}
}

