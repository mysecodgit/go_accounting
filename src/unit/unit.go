package unit

import "strings"

type Unit struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	BuildingID int    `json:"building_id"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

func (u *Unit) Validate() map[string]string {
	errors := make(map[string]string)

	if strings.TrimSpace(u.Name) == "" {
		errors["name"] = "Name cannot be empty"
	}

	if u.BuildingID <= 0 {
		errors["building_id"] = "Building ID must be greater than 0"
	}

	if len(errors) == 0 {
		return nil
	}

	return errors
}
