package building

import "strings"

type Building struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (b *Building) Validate() map[string]string {
	errors := make(map[string]string)

	if strings.TrimSpace(b.Name) == "" {
		errors["name"] = "Name cannot be empty"
	}

	if len(errors) == 0 {
		return nil
	}

	return errors
}
