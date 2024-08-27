package utils

type Hasher interface {
	HashPassword(password string) string
	HashPasswordWithSalt(password string, salt string) (string, error)
	VerifyPassword(password string, hash string) error
	GetSalt() []byte
}
