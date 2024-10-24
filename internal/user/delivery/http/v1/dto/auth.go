package dto

type RegisterUser struct {
	Email       string `json:"email" validate:"required,email,max=40"`
	PhoneNumber string `json:"phone_number" validate:"required,phone"`
	FirstName   string `json:"first_name" validate:"required,alphaunicode,min=2,max=30"`
	LastName    string `json:"last_name" validate:"required,alphaunicode,min=2,max=30"`
	Password    string `json:"password" validate:"required,password"`
} // @name RegisterUserInput

type RegisterProvider struct {
	RegisterUser
	PersonalIDNumber string `json:"personal_id_number" validate:"required,number"`
} // @name RegisterProviderInput

type Login struct {
	Email    string `json:"email" validate:"required,email,max=40"`
	Password string `json:"password" validate:"required,password"`
} // @name LoginInput

type Email struct {
	Email string `json:"email" validate:"required,email,max=40"`
} // @name EmailInput
