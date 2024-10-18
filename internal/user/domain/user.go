package domain

import (
	"time"

	"github.com/hexley21/fixup/internal/common/enum"
)

// User domain entity
type (
	User struct {
		ID           int64
		Picture      string
		PersonalInfo *UserPersonalInfo
		AccountInfo  UserAccountInfo
		CreatedAt    time.Time
	}

	UserPersonalInfo struct {
		Email       string
		PhoneNumber string
		FirstName   string
		LastName    string
	}

	UserAccountInfo struct {
		Role   enum.UserRole
		Active bool
	}
)

func NewUser(id int64, picture string, personalInfo *UserPersonalInfo, accountInfo UserAccountInfo, createdAt time.Time) *User {
	return &User{
		ID:           id,
		Picture:      picture,
		PersonalInfo: personalInfo,
		AccountInfo:  accountInfo,
		CreatedAt:    createdAt,
	}
}

func NewUserPersonalInfo(email string, phoneNumber string, firstName string, lastName string) *UserPersonalInfo {
	return &UserPersonalInfo{
		Email:       email,
		PhoneNumber: phoneNumber,
		FirstName:   firstName,
		LastName:    lastName,
	}
}

func NewUserAccountInfo(role enum.UserRole, active bool) (UserAccountInfo) {
	return UserAccountInfo{
		Role:   role,
		Active: active,
	}
}
