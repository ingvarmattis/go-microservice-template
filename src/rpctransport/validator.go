package rpctransport

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

func NewValidator() (*validator.Validate, error) {
	validate := validator.New()

	if err := validate.RegisterValidation("serviceName", validateServiceName); err != nil {
		return nil, fmt.Errorf("error while register validation `serviceName` | %w", err)
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

func validateServiceName(fl validator.FieldLevel) bool {
	serviceName := fl.Field().String()

	return len(serviceName) != 0
}
