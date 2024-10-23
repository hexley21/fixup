package playground_validator

import (
	"log"
	"regexp"

	"github.com/go-playground/validator/v10"
	"github.com/hexley21/fixup/pkg/http/rest"
)

type playgroundValidator struct {
	validator *validator.Validate
}

func New() *playgroundValidator {
	validate := validator.New()
	err := validate.RegisterValidation("phone", phoneNumberValidator)
	if err != nil {
		log.Fatalf("failed to register phone validator: %v", err)
	}

	err = validate.RegisterValidation("password", passwordValidator)
	if err != nil {
		log.Fatalf("failed to register password validator: %v", err)
	}

	return &playgroundValidator{validator: validate}
}

func (v *playgroundValidator) Validate(i any) *rest.ErrorResponse {
	err := v.validator.Struct(i)
	if err != nil {
		return rest.NewInvalidArgumentsError(err)
	}

	return nil
}

// phoneNumberValidator checks if string
// Starts with 1-9,
// Continues with 0-9,
// Between 7-14 characters,
// Expects only a phone number, without +.
func phoneNumberValidator(fl validator.FieldLevel) bool {
	phone := fl.Field().String()

	return regexp.MustCompile("^[1-9]?[0-9]{7,14}$").MatchString(phone)
}

// passwordValidator checks if string is any ASCII character but space and control characters.
func passwordValidator(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	return regexp.MustCompile(`^[\x21-\x7E]{8,36}$`).MatchString(password)
}
