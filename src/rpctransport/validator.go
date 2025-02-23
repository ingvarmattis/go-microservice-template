package rpctransport

import (
	"fmt"
	"net/mail"

	"github.com/go-playground/validator/v10"
)

func NewValidator() (*validator.Validate, error) {
	validate := validator.New()

	if err := validate.RegisterValidation("email", validateEmail); err != nil {
		return nil, fmt.Errorf("error while register validation `email` | %w", err)
	}

	if err := validate.RegisterValidation("password", validatePassword); err != nil {
		return nil, fmt.Errorf("error while register validation `password` | %w", err)
	}

	return validate, nil
}

func MustValidate() *validator.Validate {
	validate, err := NewValidator()
	if err != nil {
		panic(err)
	}

	return validate
}

func validateEmail(fl validator.FieldLevel) bool {
	email := fl.Field().String()

	_, err := mail.ParseAddress(email)

	return err == nil
}

func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	return len(password) < 8
}
