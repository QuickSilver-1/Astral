package contracts

import (
	"astral/internal/domain/dto"
	"astral/internal/domain/user"
)

type AuthInterface interface {
	Registration(adminToken string, user dto.UserData) (*user.User, error)
	Login(user dto.UserData) (*user.Token, error)
	Authorization(token string) (*user.Token, error)
	CloseSession(token dto.TokenData) (*user.Token, error)
}
