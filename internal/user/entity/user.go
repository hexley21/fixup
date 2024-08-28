package entity

import (
	"github.com/hexley21/handy/internal/user/enum"
	"github.com/jackc/pgx/v5/pgtype"
)

type User struct {
	ID          int64            `json:"id"`
	FirstName   string           `json:"first_name"`
	LastName    string           `json:"last_name"`
	PhoneNumber string           `json:"phone_number"`
	Email       string           `json:"email"`
	Hash        string           `json:"hash"`
	Role        enum.UserRole    `json:"role"`
	UserStatus  pgtype.Bool      `json:"user_status"`
	CreatedAt   pgtype.Timestamp `json:"created_at"`
}
