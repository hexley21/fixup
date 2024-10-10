package argon2

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/hexley21/fixup/pkg/config"
	"github.com/hexley21/fixup/pkg/hasher"
	"golang.org/x/crypto/argon2"
)

type argon2Hasher struct {
	config.Argon2
}

func NewHasher(hasherCfg config.Argon2) *argon2Hasher {
	return &argon2Hasher{hasherCfg}
}

func (h *argon2Hasher) HashPassword(password string) (string, error) {
	saltInBytes, err := h.GetSalt()
	if err != nil {
		return "", err
	}
	salt := base64.RawStdEncoding.EncodeToString(saltInBytes)

	hash := base64.RawStdEncoding.EncodeToString(argon2.Key([]byte(password), saltInBytes, h.Time, h.Memory, h.Threads, h.KeyLen))

	return fmt.Sprint(hash, salt), nil
}

func (h *argon2Hasher) HashPasswordWithSalt(password string, salt string) (string, error) {
	decodedSalt, err := base64.RawStdEncoding.DecodeString(salt)
	if err != nil {
		return "", err
	}

	return fmt.Sprint(base64.RawStdEncoding.EncodeToString(argon2.Key([]byte(password), decodedSalt, h.Time, h.Memory, h.Threads, h.KeyLen)), salt), nil
}

func (h *argon2Hasher) VerifyPassword(password string, hash string) error {
	salt := hash[h.Breakpoint:]

	newHash, err := h.HashPasswordWithSalt(password, salt)

	if err != nil {
		return err
	}

	if newHash == hash {
		return nil
	}

	return hasher.ErrPasswordMismatch
}

func (h *argon2Hasher) GetSalt() ([]byte, error) {
	salt := make([]byte, h.SaltLen)
	_, err := io.ReadFull(rand.Reader, salt)
	if err != nil {
		return nil, err
	}

	return salt, nil
}
