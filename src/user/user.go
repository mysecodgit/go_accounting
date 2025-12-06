package user


type User struct {
	ID       int    `json:"id" validate:"omitempty,numeric"`
	Name     string `json:"name" validate:"required,min=2,max=50"`
	Username string `json:"username" validate:"required,min=3,max=20"`
	Phone    string `json:"phone" validate:"required,min=7,max=30"`
	Password string `json:"password" validate:"required,min=6,max=20"`
}

