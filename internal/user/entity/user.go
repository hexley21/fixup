package entity

import (
	"github.com/hexley21/handy/internal/user/enum"
	"github.com/jackc/pgx/v5/pgtype"
)

type User struct {
	ID          int64
	FirstName   string
	LastName    string
	PhoneNumber string
	Email       string
	Hash        string
	Role        enum.UserRole
	UserStatus  pgtype.Bool
	CreatedAt   pgtype.Timestamp
}
