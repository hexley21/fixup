package dto

import "time"

type User struct {
	ID          string    `json:"id"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	PhoneNumber string    `json:"phone_number"`
	Email       string    `json:"email"`
	PictureUrl  string    `json:"picture_url,omitempty"`
	Role        string    `json:"role"`
	UserStatus  bool      `json:"user_status"`
	CreatedAt   time.Time `json:"created_at"`
} // @name User

type Profile struct {
	ID          string    `json:"id"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	PhoneNumber string    `json:"phone_number,omitempty"`
	Email       string    `json:"email,omitempty"`
	PictureUrl  string    `json:"picture_url,omitempty"`
	Role        string    `json:"role"`
	UserStatus  bool      `json:"user_status"`
	CreatedAt   time.Time `json:"created_at"`
} // @name profile

type UpdateUser struct {
	Email       *string `json:"email,omitempty" validate:"omitempty,email,max=40"`
	PhoneNumber *string `json:"phone_number,omitempty" validate:"omitempty,phone"`
	FirstName   *string `json:"first_name,omitempty" validate:"omitempty,alphaunicode,min=2,max=30"`
	LastName    *string `json:"last_name,omitempty" validate:"omitempty,alphaunicode,min=2,max=30"`
} // @name UpdateUserInput

type UpdatePassword struct {
	OldPassword string `json:"old_password" validate:"required,min=8"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
} // @name UpdatePasswordInput
