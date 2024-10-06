package verify_jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type VerifyClaims struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

func newClaims(id string, email string, expiry time.Duration) VerifyClaims {
	return VerifyClaims{
		ID:    id,
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
		},
	}
}

func mapToClaim(mapClaims any) VerifyClaims {
	claims, ok := mapClaims.(jwt.MapClaims)
	if !ok {
		return VerifyClaims{}
	}

	return VerifyClaims{
		ID: claims["id"].(string),
		Email: claims["email"].(string),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Unix(int64(claims["exp"].(float64)), 0)),
		},
	}
}
