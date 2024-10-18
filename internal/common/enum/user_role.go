package enum

import "errors"

var (
	ErrInvalidRole = errors.New("invalid role")
)

type UserRole string

const (
	UserRoleCUSTOMER  UserRole = "CUSTOMER"
	UserRolePROVIDER  UserRole = "PROVIDER"
	UserRoleMODERATOR UserRole = "MODERATOR"
	UserRoleADMIN     UserRole = "ADMIN"
)

func (e UserRole) Valid() bool {
	switch e {
	case UserRoleCUSTOMER,
		UserRolePROVIDER,
		UserRoleMODERATOR,
		UserRoleADMIN:
		return true
	}
	return false
}

func ParseRole(r string) (UserRole, error) {
	role := UserRole(r)
	if !role.Valid() {
		return "", ErrInvalidRole
	}

	return role, nil
}
