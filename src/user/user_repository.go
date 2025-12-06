// repository/user_repository.go
package user

import (
	"database/sql"
	"fmt"
)

type UserRepository interface {
	Create(user User) (User,error)
	GetByID(id int) (User, error)
	GetAll() ([]User, error)
}

type userRepo struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(user User) (User,error) {
	result, err := r.db.Exec("INSERT INTO users (name, username, phone, password) VALUES (?, ?, ?, ?)",
		user.Name, user.Username, user.Phone, user.Password)

	id, _ := result.LastInsertId()
	user.ID = int(id)

	return user,err
}

func (r *userRepo) GetByID(id int) (User, error) {
	var user User
	err := r.db.QueryRow("SELECT id, name, username, phone FROM users WHERE id = ?", id).
		Scan(&user.ID, &user.Name, &user.Username, &user.Phone)
	
	if err == sql.ErrNoRows {
		return user, fmt.Errorf("id does not exist")
	}
	
	return user, err
}

func (r *userRepo) GetAll() ([]User, error) {
	rows, err := r.db.Query("SELECT id, name, username, phone, password FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		var u User
		err := rows.Scan(&u.ID, &u.Name, &u.Username, &u.Phone, &u.Password)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}
