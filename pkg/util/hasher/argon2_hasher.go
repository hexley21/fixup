package hasher

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"github.com/hexley21/handy/pkg/config"
	"golang.org/x/crypto/argon2"
)

var ErrPasswordMismatch = errors.New("password does not match")

type Argon2Hasher struct {
	config.Argon2
}

func New(hasherCfg config.Argon2) *Argon2Hasher {
	return &Argon2Hasher{hasherCfg}
}

func (h *Argon2Hasher) HashPassword(password string) string {
	saltInBytes := h.GetSalt()
	salt := base64.RawStdEncoding.EncodeToString(saltInBytes)

	hash := base64.RawStdEncoding.EncodeToString(argon2.Key([]byte(password), saltInBytes, h.Time, h.Memory, h.Threads, h.KeyLen))

	return fmt.Sprint(hash, salt)
}

func (h *Argon2Hasher) HashPasswordWithSalt(password string, salt string) (string, error) {
	decodedSalt, err := base64.RawStdEncoding.DecodeString(salt)
	if err != nil {
		return "", err
	}

	return fmt.Sprint(base64.RawStdEncoding.EncodeToString(argon2.Key([]byte(password), decodedSalt, h.Time, h.Memory, h.Threads, h.KeyLen)), salt), nil
}

func (h *Argon2Hasher) VerifyPassword(password string, hash string) error {
	salt := hash[h.Breakpoint:]

	newHash, err := h.HashPasswordWithSalt(password, salt)

	if err != nil {
		return err
	}

	if newHash == hash {
		return nil
	}

	return ErrPasswordMismatch
}

func (h *Argon2Hasher) GetSalt() []byte {
	salt := make([]byte, h.SaltLen)
	io.ReadFull(rand.Reader, salt)

	return salt
}
