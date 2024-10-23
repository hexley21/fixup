package dto

import "time"

type User struct {
	ID           string            `json:"id"`
	*UserPersonalInfo
	PictureUrl   string            `json:"picture_url,omitempty"`
	Role         string            `json:"role"`
	Verified     bool              `json:"verified"`
	CreatedAt    time.Time         `json:"created_at"`
} // @name User

type UserPersonalInfo struct {
	Email       string `json:"email" validate:"omitempty,email,max=40"`
	PhoneNumber string `json:"phone_number" validate:"omitempty,phone"`
	FirstName   string `json:"first_name" validate:"omitempty,alphaunicode,min=2,max=30"`
	LastName    string `json:"last_name" validate:"omitempty,alphaunicode,min=2,max=30"`
} // @name UserPersonalInfo


type UpdatePassword struct {
	OldPassword string `json:"old_password" validate:"required,password"`
	NewPassword string `json:"new_password" validate:"required,password"`
} // @name UpdatePasswordInput
