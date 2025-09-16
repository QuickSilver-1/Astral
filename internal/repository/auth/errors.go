package authrepo

import "errors"

var (
	ErrUserAlreadyExists = errors.New("user with this login already exists")
	ErrNoRows            = errors.New("no rows in result set")
)
