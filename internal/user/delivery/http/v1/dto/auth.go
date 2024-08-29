package dto

type RegisterUser struct {
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Password    string `json:"password"`
}

type RegisterProvider struct {
	Email            string `json:"email"`
	PhoneNumber      string `json:"phone_number"`
	FirstName        string `json:"first_name"`
	LastName         string `json:"last_name"`
	Password         string `json:"password"`
	PersonalIDNumber string `json:"personal_id_number"`
}