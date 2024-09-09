package hasher

import "errors"

var ErrPasswordMismatch = errors.New("password does not match")

type Hasher interface {
	HashPassword(password string) string
	HashPasswordWithSalt(password string, salt string) (string, error)
	VerifyPassword(password string, hash string) error
	GetSalt() []byte
}
