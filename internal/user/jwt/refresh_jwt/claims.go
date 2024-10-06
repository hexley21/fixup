package refresh_jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type RefreshClaims struct {
	ID       string        `json:"id"`
	jwt.RegisteredClaims
}

func NewClaims(id string, expiry time.Duration) RefreshClaims {
	return RefreshClaims{
		ID:       id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
		},
	}
}

func mapToClaim(mapClaims any) RefreshClaims {
	claims, ok := mapClaims.(jwt.MapClaims)
	if !ok {
		return RefreshClaims{}
	}

	return RefreshClaims{
		ID: claims["id"].(string),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Unix(int64(claims["exp"].(float64)), 0)),
		},
	}
}
