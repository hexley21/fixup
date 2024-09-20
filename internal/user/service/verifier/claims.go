package verifier

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

	return newClaims(claims["id"].(string), claims["email"].(string), time.Duration(claims["exp"].(float64)))
}
