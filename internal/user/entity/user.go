package entity

import (
	"github.com/hexley21/fixup/internal/common/enum"
	"github.com/jackc/pgx/v5/pgtype"
)

type User struct {
	ID          int64
	FirstName   string
	LastName    string
	PhoneNumber string
	Email       string
	PictureName pgtype.Text
	Hash        string
	Role        enum.UserRole
	UserStatus  pgtype.Bool
	CreatedAt   pgtype.Timestamp
}
