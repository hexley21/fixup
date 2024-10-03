package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

type aesEncryptor struct {
	key []byte
}

func NewAesEncryptor(key string) *aesEncryptor {
	return &aesEncryptor{key: []byte(key)}
}


func (e *aesEncryptor) Encrypt(value []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(value))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(value))

	return ciphertext, nil
}

func (e *aesEncryptor) Decrypt(value []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, err
	}

	if len(value) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	iv := value[:aes.BlockSize]
	value = value[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	dec := make([]byte, len(value))
	stream.XORKeyStream(dec, value)

	return dec, nil
}