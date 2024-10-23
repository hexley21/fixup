package argon2_test

import (
	"testing"

	"github.com/hexley21/fixup/pkg/config"
	"github.com/hexley21/fixup/pkg/hasher/argon2"
	"github.com/stretchr/testify/assert"
)

const (
	hashLen = 128
	normalPassword = "abcdefghijklmnopqrstuvwxyz123456789"
	crazyPassword = "abcdefghijklmnopqrstuvwxyz123456789😀😀😀😀無無無無無無無無無"
)

var (
	argon2Hasher = argon2.NewHasher(config.Argon2{
		SaltLen: 16,
		KeyLen: 79,
		Time: 1,
		Memory: 47104,
		Threads: 1,
	})
)


func TestHashPassword_Success(t *testing.T) {
	tests := []struct{
		name string
		password string
	}{
		{
			name: "Normal Password",
			password: normalPassword,
		},
		{
			name: "Crazy Password",
			password: crazyPassword,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := argon2Hasher.HashPassword(tt.password)
			assert.NoError(t, err)
			assert.Len(t, hash, hashLen)
		})
	}
}