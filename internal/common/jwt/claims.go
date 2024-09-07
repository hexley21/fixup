package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hexley21/fixup/internal/user/enum"
)

type UserClaims struct {
	ID   string        `json:"id"`
	Role enum.UserRole `json:"role"`
	jwt.RegisteredClaims
}

func newUserClaims(id string, role string, expiry time.Duration) UserClaims {
	var userRole enum.UserRole
	userRole.Scan(role)

	return UserClaims{
		ID:   id,
		Role: userRole,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
		},
	}
}

func MapToClaim(mapClaims any) UserClaims {
	claims, ok := mapClaims.(jwt.MapClaims)
	if !ok {
		return UserClaims{}
	}

	return newUserClaims(claims["id"].(string), claims["role"].(string), time.Duration(claims["exp"].(float64)))
}
