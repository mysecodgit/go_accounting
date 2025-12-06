package people_types

import "strings"

type PeopleType struct {
	ID         int    `json:"id"`
	Title      string `json:"title"`
	BuildingID int    `json:"building_id"`
}

func (p *PeopleType) Validate() map[string]string {
	errors := make(map[string]string)

	if strings.TrimSpace(p.Title) == "" {
		errors["title"] = "Title cannot be empty"
	}

	if p.BuildingID <= 0 {
		errors["building_id"] = "Building ID must be greater than 0"
	}

	if len(errors) == 0 {
		return nil
	}

	return errors
}
