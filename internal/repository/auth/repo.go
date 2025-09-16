package authrepo

import (
	"astral/internal/domain/dto"
	"astral/internal/domain/user"
	"context"
)

type AuthRepo interface {
	CreateUser(ctx context.Context, user dto.UserData) (*user.User, error)
	GetUserByLogin(ctx context.Context, login string) (*user.User, error)
	CreateToken(ctx context.Context, token dto.TokenData) (*user.Token, error)
	GetTokensByLogin(ctx context.Context, login string) ([]user.Token, error)
	DeleteToken(ctx context.Context, token string) (*user.Token, error)
}
