package enum

import (
	"database/sql/driver"
	"fmt"
)

type UserRole string

const (
	UserRoleCUSTOMER  UserRole = "CUSTOMER"
	UserRolePROVIDER  UserRole = "PROVIDER"
	UserRoleMODERATOR UserRole = "MODERATOR"
	UserRoleADMIN     UserRole = "ADMIN"
)

func (e *UserRole) Scan(src any) error {
	switch s := src.(type) {
	case []byte:
		*e = UserRole(s)
	case string:
		*e = UserRole(s)
	default:
		return fmt.Errorf("unsupported scan type for UserRole: %T", src)
	}
	return nil
}

type NullUserRole struct {
	UserRole UserRole `json:"user_role"`
	Valid    bool     `json:"valid"` // Valid is true if UserRole is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullUserRole) Scan(value any) error {
	if value == nil {
		ns.UserRole, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.UserRole.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullUserRole) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.UserRole), nil
}

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
