package authservice

import "errors"

var (
	ErrAccessDenied = errors.New("access denied")
	ErrInvalidToken = errors.New("invalid token")
)
