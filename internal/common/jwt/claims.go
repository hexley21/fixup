package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hexley21/fixup/internal/user/enum"
)

type UserClaims struct {
	ID       string        `json:"id"`
	Role     enum.UserRole `json:"role"`
	Verified bool          `json:"verified"`
	jwt.RegisteredClaims
}

func NewClaims(id string, role string, Verified bool, expiry time.Duration) UserClaims {
	var userRole enum.UserRole
	userRole.Scan(role)

	return UserClaims{
		ID:       id,
		Role:     userRole,
		Verified: Verified,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
		},
	}
}

func mapToClaim(mapClaims any) UserClaims {
	claims, ok := mapClaims.(jwt.MapClaims)
	if !ok {
		return UserClaims{}
	}

	return NewClaims(claims["id"].(string), claims["role"].(string), claims["verified"].(bool), time.Duration(claims["exp"].(float64)))
}
