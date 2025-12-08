package people_types

import "strings"

type PeopleTypeResponse struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
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
