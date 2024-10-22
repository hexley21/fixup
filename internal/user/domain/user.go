package domain

import (
	"time"

	"github.com/hexley21/fixup/internal/common/enum"
)

type (
	User struct {
		ID           int64
		Picture      string
		PersonalInfo *UserPersonalInfo
		AccountInfo  UserAccountInfo
		CreatedAt    time.Time
	} // User domain entity

	UserPersonalInfo struct {
		Email       string
		PhoneNumber string
		FirstName   string
		LastName    string
	} // User personal info Value Object

	UserAccountInfo struct {
		Role     enum.UserRole
		Verified bool
	} // User account info Value Object

	UserIdentity struct {
		ID          int64
		AccountInfo UserAccountInfo
	} // Partial User domain entity representation
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

func NewUserAccountInfo(role enum.UserRole, verified bool) UserAccountInfo {
	return UserAccountInfo{
		Role:     role,
		Verified: verified,
	}
}

func NewUserIdentity(ID int64, accountInfo UserAccountInfo) UserIdentity {
	return UserIdentity{
		ID:          ID,
		AccountInfo: accountInfo,
	}
}
