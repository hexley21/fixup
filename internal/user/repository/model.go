package repository

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Provider struct {
	PersonalIDNumber  []byte
	PersonalIDPreview string
	UserID            int64
}

type User struct {
	ID          int64            `json:"id"`
	FirstName   string           `json:"first_name"`
	LastName    string           `json:"last_name"`
	PhoneNumber string           `json:"phone_number"`
	Email       string           `json:"email"`
	Picture     pgtype.Text      `json:"picture"`
	Hash        string           `json:"hash"`
	Role        string           `json:"role"`
	Verified    pgtype.Bool      `json:"verified"`
	CreatedAt   pgtype.Timestamp `json:"created_at"`
}
