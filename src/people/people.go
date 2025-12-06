package people

import "strings"

type Person struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Phone      string `json:"phone"`
	TypeID     int    `json:"type_id"`
	BuildingID int    `json:"building_id,omitempty"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

func (p *Person) Validate() map[string]string {
	errors := make(map[string]string)

	if strings.TrimSpace(p.Name) == "" {
		errors["name"] = "Name cannot be empty"
	}

	if strings.TrimSpace(p.Phone) == "" {
		errors["phone"] = "Phone cannot be empty"
	}

	if p.TypeID <= 0 {
		errors["type_id"] = "Type ID must be greater than 0"
	}

	if len(errors) == 0 {
		return nil
	}

	return errors
}

