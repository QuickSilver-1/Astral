package contracts

import "astral/internal/domain/dto"

type ValidationInterface interface {
	ValidateUserData(dto.UserData) error
}
