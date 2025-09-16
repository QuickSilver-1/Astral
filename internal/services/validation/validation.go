package validationservice

import (
	"astral/internal/domain/dto"
	"log/slog"
	"regexp"
	"strings"
	"unicode"
)

type ValidationService struct {
	logger *slog.Logger
}

func NewValidationService(logger *slog.Logger) *ValidationService {
	return &ValidationService{
		logger: logger.With("service", "NewValidationService"),
	}
}

func (s *ValidationService) ValidateUserData(userData dto.UserData) error {
	op := "services.validation.validation.ValidateUserData"

	var err error
	defer func() {
		if err != nil {
			s.logger.Debug("validation error", "func", op, "error", err)
		}
	}()

	if err = validateLogin(userData.Login); err != nil {
		return err
	}

	if err = validatePassword(userData.Password); err != nil {
		return err
	}

	return nil
}

func validateLogin(login string) error {
	if len(login) < 8 {
		return NewErrValidationUserData("login must be at least 8 characters long")
	}

	loginRegex := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	if !loginRegex.MatchString(login) {
		return NewErrValidationUserData("login must contain only latin letters and digits")
	}

	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return NewErrValidationUserData("password must be at least 8 characters long")
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasDigit   bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case !unicode.IsLetter(char) && !unicode.IsDigit(char):
			hasSpecial = true
		}
	}

	var errorMessages []string
	if !hasUpper || !hasLower {
		errorMessages = append(errorMessages, "password must contain at least 2 letters in different cases")
	}

	if !hasDigit {
		errorMessages = append(errorMessages, "password must contain at least 1 digit")
	}

	if !hasSpecial {
		errorMessages = append(errorMessages, "password must contain at least 1 special character")
	}

	if len(errorMessages) > 0 {
		return NewErrValidationUserData(strings.Join(errorMessages, "; "))
	}

	return nil
}
