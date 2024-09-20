package validator

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

type playgroundValidator struct {
	validator *validator.Validate
}

func NewValidator() *playgroundValidator {
	validate := validator.New()
	validate.RegisterValidation("phone", phoneNumberValidator)

	return &playgroundValidator{validator: validate}
}

func (v *playgroundValidator) Validate(i any) error {
	return v.validator.Struct(i)
}

func phoneNumberValidator(fl validator.FieldLevel) bool {
	phone := fl.Field().String()

	re := regexp.MustCompile(`^[1-9]?[0-9]{7,14}$`)
	return re.MatchString(phone)
}
