package user

import "strings"

type RegisterUserRequest struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type UpdateUserRequest struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Phone    string `json:"phone"`
}

type UserResponse struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Phone    string `json:"phone"`
}

func (r *RegisterUserRequest) Validate() map[string]string {
	errors := make(map[string]string)

	if len(strings.TrimSpace(r.Name)) < 2 || len(r.Name) > 50 {
		errors["name"] = "Name must be between 2 and 50 characters"
	}

	if len(strings.TrimSpace(r.Username)) < 3 || len(r.Username) > 20 {
		errors["username"] = "Username must be between 3 and 20 characters"
	}

	if len(strings.TrimSpace(r.Phone)) < 7 || len(r.Phone) > 30 {
		errors["phone"] = "Phone must be between 7 and 30 characters"
	}

	if len(r.Password) < 6 || len(r.Password) > 20 {
		errors["password"] = "Password must be between 6 and 20 characters"
	}

	if len(errors) == 0 {
		return nil
	}

	return errors
}


func (u *UpdateUserRequest) Validate() map[string]string {
	errors := make(map[string]string)

	if len(strings.TrimSpace(u.Name)) < 2 || len(u.Name) > 50 {
		errors["name"] = "Name must be between 2 and 50 characters"
	}

	if len(strings.TrimSpace(u.Username)) < 3 || len(u.Username) > 20 {
		errors["username"] = "Username must be between 3 and 20 characters"
	}

	if len(strings.TrimSpace(u.Phone)) < 7 || len(u.Phone) > 30 {
		errors["phone"] = "Phone must be between 7 and 30 characters"
	}

	if len(errors) == 0 {
		return nil
	}

	return errors
}


func (r *RegisterUserRequest) ToUser() User {
	return User{
		ID:0,
		Name: r.Name,
		Phone: r.Phone,
		Password: r.Password,
		Username: r.Username,
	}
}

func (r *User) ToUserResponse() UserResponse {
	return UserResponse{
		ID:r.ID,
		Name: r.Name,
		Phone: r.Phone,
		Username: r.Username,
	}
}