package authservice

import (
	"astral/internal/domain/contracts"
	"astral/internal/domain/dto"
	"astral/internal/domain/user"
	authrepo "astral/internal/repository/auth"
	"context"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	DEFAULT_TIMEOUT = time.Second * 5
)

type AuthService struct {
	repo       authrepo.AuthRepo
	validation contracts.ValidationInterface
	logger     *slog.Logger
	secretJWT  string
	adminToken string
}

func NewAuthService(repo authrepo.AuthRepo, logger *slog.Logger, validation contracts.ValidationInterface, secret, adminToken string) *AuthService {
	return &AuthService{
		repo:       repo,
		validation: validation,
		logger:     logger.With("service", "AuthService"),
		secretJWT:  secret,
		adminToken: adminToken,
	}
}

func (s *AuthService) Registration(adminToken string, userData dto.UserData) (*user.User, error) {
	const op = "services.authorization.auth.Registration"
	s.logger.Info("Usecase start", "func", op, "login", userData.Login)

	if adminToken != s.adminToken {
		s.logger.Warn("invalid admin token", "func", op, "login", userData.Login)
		return nil, ErrAccessDenied
	}

	err := s.validation.ValidateUserData(userData)
	if err != nil {
		return nil, err
	}

	hashedPass, err := s.hashPassword(userData.Password)
	if err != nil {
		s.logger.Error("failed to hash password", "func", op, "login", userData.Login, "error", err)
		return nil, err
	}

	userData.Password = *hashedPass

	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_TIMEOUT)
	defer cancel()

	user, err := s.repo.CreateUser(ctx, userData)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(userData dto.UserData) (*user.Token, error) {
	const op = "services.authorization.auth.Login"
	s.logger.Info("Usecase start", "func", op, "login", userData.Login)

	access, err := s.validatePassword(userData)
	if err != nil {
		return nil, err
	}

	if !access {
		s.logger.Warn("failed to get access", "func", op, "login", userData.Login)
		return nil, ErrAccessDenied
	}

	jwtToken, err := s.makeJWT(userData.Login)
	if err != nil {
		return nil, err
	}

	token := user.NewToken(*jwtToken, userData.Login)
	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_TIMEOUT)
	defer cancel()
	
	tokenData := dto.TokenData{
		Token: token.Token,
		Login: token.Login,
	}

	token, err = s.repo.CreateToken(ctx, tokenData)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (s *AuthService) Authorization(token string) (*user.Token, error) {
	const op = "services.authorization.auth.Authorization"
	s.logger.Info("Usecase start", "func", op, "token", token)

	tokenData, err := s.decodeJWT(token)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_TIMEOUT)
	defer cancel()

	tokens, err := s.repo.GetTokensByLogin(ctx, tokenData.Login)
	if err != nil {
		return nil, err
	}

	t := s.findToken(token, tokens)
	if t == nil {
		s.logger.Warn("token does not belong to user", "func", op, "login", tokenData.Login)
		return nil, ErrAccessDenied
	}

	return t, nil
}

func (s *AuthService) CloseSession(tokenData dto.TokenData) (*user.Token, error) {
	const op = "services.authorization.auth.CloseSession"
	s.logger.Info("Usecase start", "func", op, "login", tokenData.Login)

	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_TIMEOUT)
	defer cancel()

	tokens, err := s.repo.GetTokensByLogin(ctx, tokenData.Login)
	if err != nil {
		return nil, err
	}

	if s.findToken(tokenData.Token, tokens) == nil {
		s.logger.Warn("token does not belong to user", "func", op, "login", tokenData.Login)
		return nil, ErrAccessDenied
	}

	ctx, cancel = context.WithTimeout(context.Background(), DEFAULT_TIMEOUT)
	defer cancel()
	deletedToken, err := s.repo.DeleteToken(ctx, tokenData.Token)
	if err != nil {
		return nil, err
	}

	return deletedToken, nil
}

func (s *AuthService) hashPassword(password string) (*string, error) {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	hashedPassStr := string(hashedPass)
	return &hashedPassStr, nil
}

func (s *AuthService) validatePassword(userData dto.UserData) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_TIMEOUT)
	defer cancel()

	user, err := s.repo.GetUserByLogin(ctx, userData.Login)
	if err != nil {
		return false, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(userData.Password))
	if err != nil {
		return false, nil
	}

	return true, nil
}

func (s *AuthService) findToken(token string, tokens []user.Token) *user.Token {	
	for _, t := range tokens {
		if t.Token == token {
			return &t
		}
	}

	return nil
}
