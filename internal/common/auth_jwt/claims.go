package auth_jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hexley21/fixup/internal/common/enum"
	"github.com/hexley21/fixup/pkg/http/rest"
)

type ctxKey string

const (
	AuthJWTKey ctxKey = "auth_jwt"
)

var (
	ErrJWTNotSet = rest.NewInternalServerError(errors.New("auth jwt not set"))
)


type UserClaims struct {
	Data UserData
	jwt.RegisteredClaims
}

type UserData struct {
	ID       string        `json:"id"`
	Role     enum.UserRole `json:"role"`
	Verified bool          `json:"verified"`
}

func NewClaims(id string, role enum.UserRole, Verified bool, expiry time.Duration) UserClaims {
	return UserClaims{
		Data: UserData{
			ID:       id,
			Role:     role,
			Verified: Verified,
		},
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
		Data: UserData{
			ID:       claims["id"].(string),
			Role:     enum.UserRole(claims["role"].(string)),
			Verified: claims["verified"].(bool),
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Unix(int64(claims["exp"].(float64)), 0)),
		},
	}
}
