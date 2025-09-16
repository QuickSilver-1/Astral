package authrepo

import (
	"astral/internal/domain/dto"
	"astral/internal/domain/user"
	pg "astral/internal/repository/db/postgres"
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/doug-martin/goqu/v9"
)

const (
	TABLE_USERS  = "users"
	TABLE_TOKENS = "tokens"
)

type AuthPersister struct {
	dial    goqu.DialectWrapper
	storage *pg.Storage
	logger  *slog.Logger
}

func NewUserPersister(storage *pg.Storage, logger *slog.Logger) *AuthPersister {
	return &AuthPersister{
		dial:    goqu.Dialect(pg.DRIVER),
		storage: storage,
		logger:  logger,
	}
}

func (p *AuthPersister) CreateUser(ctx context.Context, userData dto.UserData) (*user.User, error) {
	const op = "repository.user.persister.CreateUser"

	query, _, err := p.dial.Insert(TABLE_USERS).
		Rows(
			goqu.Record{
				"login":    userData.Login,
				"password": userData.Password,
			},
		).Returning("login", "password", "created_at", "updated_at").ToSQL()
	if err != nil {
		p.logger.Error("failed to build user creation query", "func", op, "login", userData.Login, "error", err)
		return nil, errors.New("failed to build user creation query")
	}

	var usr user.User
	err = p.storage.DB.QueryRowContext(ctx, query).Scan(&usr.Login, &usr.Password, &usr.CreatedAt, &usr.UpdatedAt)
	if err != nil {
		if pg.IsDuplicateKeyError(err) {
			p.logger.Info("user with this login already exists", "func", op, "login", userData.Login, "error", ErrUserAlreadyExists)
			return nil, ErrUserAlreadyExists
		}

		p.logger.Error("failed to execute create user query", "func", op, "login", userData.Login, "error", err)
		return nil, errors.New("failed to execute create user query")
	}

	return &usr, nil
}

func (p *AuthPersister) CreateToken(ctx context.Context, token dto.TokenData) (*user.Token, error) {
	const op = "repository.user.persister.CreateToken"
 
	query, _, err := p.dial.Insert(TABLE_TOKENS).
		Rows(
			goqu.Record{
				"token":      token.Token,
				"user_login": token.Login,
			},
		).Returning("token", "user_login", "created_at", "deleted_at").ToSQL()
	if err != nil {
		p.logger.Error("failed to build token creation query", "func", op, "login", token.Login, "error", err)
		return nil, errors.New("failed to build token creation query")
	}
	
	var tkn user.Token
	err = p.storage.DB.QueryRowContext(ctx, query).Scan(&tkn.Token, &tkn.Login, &tkn.CreatedAt, &tkn.DeletedAt)
	if err != nil {
		if pg.IsDuplicateKeyError(err) {
			p.logger.Info("user with this token already exists", "func", op, "login", token.Login, "error", ErrUserAlreadyExists)
			return nil, ErrUserAlreadyExists
		}

		p.logger.Error("failed to execute create user query", "func", op, "login", token.Login, "error", err)
		return nil, errors.New("failed to execute create user query")
	}

	return &tkn, nil
}

func (p *AuthPersister) GetUserByLogin(ctx context.Context, login string) (*user.User, error) {
	const op = "repository.user.persister.GetUserByLogin"

	query, _, err := p.dial.From(TABLE_USERS).
		Select("login", "password", "created_at", "updated_at").
		Where(goqu.C("login").Eq(login)).
		ToSQL()
	if err != nil {
		p.logger.Error("failed to build get user by login query", "func", op, "login", login, "error", err)
		return nil, errors.New("failed to build get user by login query")
	}

	var user user.User
	err = p.storage.DB.QueryRowContext(ctx, query).Scan(&user.Login, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			p.logger.Error("failed to find user by login", "func", op, "login", login, "error", ErrNoRows)
			return nil, ErrNoRows
		}

		p.logger.Error("failed to execute get user by login query", "func", op, "login", login, "error", err)
		return nil, errors.New("failed to execute get user by login query")
	}

	return &user, nil
}

func (p *AuthPersister) GetTokensByLogin(ctx context.Context, login string) ([]user.Token, error) {
	const op = "repository.user.persister.GetTokensByLogin"

	query, _, err := p.dial.From(TABLE_TOKENS).
		Select("token", "user_login", "created_at", "deleted_at").
		Where(goqu.C("user_login").Eq(login)).
		Where(goqu.C("deleted_at").IsNull()).
		ToSQL()
	if err != nil {
		p.logger.Error("failed to build get tokens by login query", "func", op, "login", login, "error", err)
		return nil, errors.New("failed to build get tokens by login query")
	}

	var tokens []user.Token
	rows, err := p.storage.DB.QueryContext(ctx, query)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			p.logger.Error("failed to find tokens by login", "func", op, "login", login, "error", ErrNoRows)
			return nil, ErrNoRows
		}

		p.logger.Error("failed to execute get tokens by query", "func", op, "login", login, "error", err)
		return nil, errors.New("failed to execute get tokens by query")
	}

	if rows.Next() {
		var token user.Token
		err = rows.Scan(&token.Token, &token.Login, &token.CreatedAt, &token.DeletedAt)
		if err != nil {
			p.logger.Error("failed to scan token by query")
			return nil, errors.New("failed to scan token by query")
		}

		tokens = append(tokens, token)
	}

	return tokens, nil
}

func (p *AuthPersister) DeleteToken(ctx context.Context, token string) (*user.Token, error) {
	const op = "repository.user.persister.DeleteToken"
	query, _, err := p.dial.Update(TABLE_TOKENS).
		Set(goqu.Record{"deleted_at": time.Now()}).
		Where(goqu.C("token").Eq(token)).
		Returning("token", "user_login").
		ToSQL()
	if err != nil {
		p.logger.Error("failed to build delete token query", "func", op, "token", token, "error", err)
		return nil, errors.New("failed to build delete token query")
	}

	var tkn user.Token
	err = p.storage.DB.QueryRowContext(ctx, query).Scan(&tkn.Token, &tkn.Login)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			p.logger.Error("failed to find token", "func", op, "token", token, "error", ErrNoRows)
			return nil, ErrNoRows
		}
	}

	return &tkn, nil
}
