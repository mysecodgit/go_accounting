package people_types

import (
	"strings"
	"github.com/mysecodgit/go_accounting/src/building"
)

type PeopleTypeResponse struct {
	ID       int               `json:"id"`
	Title    string            `json:"title"`
	Building building.Building `json:"building"`
}

func (p *PeopleType) ToPeopleTypeResponse(b building.Building) PeopleTypeResponse {
	return PeopleTypeResponse{
		ID:       p.ID,
		Title:    p.Title,
		Building: b,
	}
}

type UpdatePeopleTypeRequest struct {
	Title string `json:"title"`
}

func (u *UpdatePeopleTypeRequest) Validate() map[string]string {
	errors := make(map[string]string)

	if strings.TrimSpace(u.Title) == "" {
		errors["title"] = "Title cannot be empty"
	}

	if len(errors) == 0 {
		return nil
	}

	return errors
}
