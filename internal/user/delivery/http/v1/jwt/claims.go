package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthClaims struct {
	ID   string `json:"id"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func newAuthClaims(id, role string, expiry time.Duration) *AuthClaims {
	return &AuthClaims{
		ID:   id,
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
		},
	}
}
