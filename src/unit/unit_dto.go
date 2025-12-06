package unit

import "github.com/mysecodgit/go_accounting/src/building"

type UnitResponse struct {
	ID        int                `json:"id"`
	Name      string             `json:"name"`
	Building  building.Building  `json:"building"`
	CreatedAt string             `json:"created_at"`
	UpdatedAt string             `json:"updated_at"`
}

func (u *Unit) ToUnitResponse(b building.Building) UnitResponse {
	return UnitResponse{
		ID:        u.ID,
		Name:      u.Name,
		Building:  b,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

