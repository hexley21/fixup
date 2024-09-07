package dto

type RegisterUser struct {
	Email       string `json:"email" validate:"required,email,max=40"`
	PhoneNumber string `json:"phone_number" validate:"required,phone"`
	FirstName   string `json:"first_name" validate:"required,alphaunicode,min=2,max=50"`
	LastName    string `json:"last_name" validate:"required,alphaunicode,min=2,max=50"`
	Password    string `json:"password" validate:"required,min=8"`
}

type RegisterProvider struct {
	RegisterUser
	PersonalIDNumber string `json:"personal_id_number" validate:"required,number"`
}

type Login struct {
	Email    string `json:"email" validate:"required,email,max=40"`
	Password string `json:"password" validate:"required,min=8"`
}
