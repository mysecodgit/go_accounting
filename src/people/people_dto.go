package people

import (
	"strings"
	"github.com/mysecodgit/go_accounting/src/building"
	"github.com/mysecodgit/go_accounting/src/people_types"
)

type PersonResponse struct {
	ID        int                        `json:"id"`
	Name      string                     `json:"name"`
	Phone     string                     `json:"phone"`
	Type      people_types.PeopleType    `json:"type"`
	Building  building.Building          `json:"building"`
	CreatedAt string                     `json:"created_at"`
	UpdatedAt string                     `json:"updated_at"`
}

func (p *Person) ToPersonResponse(pt people_types.PeopleType, b building.Building) PersonResponse {
	return PersonResponse{
		ID:        p.ID,
		Name:      p.Name,
		Phone:     p.Phone,
		Type:      pt,
		Building:  b,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

type UpdatePersonRequest struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

func (u *UpdatePersonRequest) Validate() map[string]string {
	errors := make(map[string]string)

	if strings.TrimSpace(u.Name) == "" {
		errors["name"] = "Name cannot be empty"
	}

	if strings.TrimSpace(u.Phone) == "" {
		errors["phone"] = "Phone cannot be empty"
	}

	if len(errors) == 0 {
		return nil
	}

	return errors
}

