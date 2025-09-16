package user

import "time"

type User struct {
	Login     string     `json:"login"`
	Password  string     `json:"password"`
	UpdatedAt *time.Time `json:"updated_at"`
	CreatedAt *time.Time `json:"created_at"`
}

func NewUser(login, password string) *User {
	return &User{
		Login:    login,
		Password: password,
	}
}
