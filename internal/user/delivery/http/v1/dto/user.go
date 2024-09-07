package dto

import "time"

type User struct {
	ID          string    `json:"id"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	PhoneNumber string    `json:"phone_number"`
	Email       string    `json:"email"`
	PictureUrl  string    `json:"picture_url`
	Role        string    `json:"role"`
	UserStatus  bool      `json:"user_status"`
	CreatedAt   time.Time `json:"created_at"`
}
