package auth_jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hexley21/fixup/internal/common/enum"
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

	return UserClaims{
		ID: claims["id"].(string),
		Role: enum.UserRole(claims["role"].(string)),
		Verified: claims["verified"].(bool),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Unix(int64(claims["exp"].(float64)), 0)),
		},
	}
}
