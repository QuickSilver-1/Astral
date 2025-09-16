package user

import "time"

type Token struct {
	Token     string     `json:"token"`
	Login     string     `json:"user_login"`
	CreatedAt *time.Time `json:"created_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

func NewToken(token, login string) *Token {
	return &Token{
		Token: token,
		Login: login,
	}
}
