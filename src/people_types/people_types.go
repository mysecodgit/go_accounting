package people_types

import "strings"

type PeopleType struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

func (p *PeopleType) Validate() map[string]string {
	errors := make(map[string]string)

	if strings.TrimSpace(p.Title) == "" {
		errors["title"] = "Title cannot be empty"
	}

	if len(errors) == 0 {
		return nil
	}

	return errors
}
