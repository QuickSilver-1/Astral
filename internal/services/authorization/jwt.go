package authservice

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type Token struct {
	Login string
	jwt.StandardClaims
}

func NewToken(login string) *Token {
	return &Token{
		Login: login,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
		},
	}
}

func (s *AuthService) makeJWT(login string) (*string, error) {
	t := NewToken(login)

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, t)
	tokenStr, err := token.SignedString([]byte(s.secretJWT))
	if err != nil {
		s.logger.Error("failed to generate JWT token", "login", login, "error", err)
		return nil, errors.New("failed to generate JWT token")
	}

	return &tokenStr, nil
}

func (s *AuthService) decodeJWT(tokenStr string) (*Token, error) {
	claims := &Token{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.secretJWT), nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			s.logger.Warn("invalid JWT token signature", "token", tokenStr)
			return nil, ErrInvalidToken
		}

		s.logger.Error("generation token error", "token", tokenStr, "error", err)
		return nil, errors.New("failed to parse JWT token")
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
