package jwt

import (
	"fmt"
	"time"

	"github.com/hexley21/fixup/internal/common/rest"
	"github.com/hexley21/fixup/pkg/jwt"
)

type AuthJwtGenerator interface {
	GenerateToken(id string, role string) (string, error)
}

type AuthJwtMapper interface {
	MapToClaim(mapClaims any) *UserClaims
}

type authJwtImpl struct {
	secretKey string
	ttl       time.Duration
}

func NewAuthJwtImpl(secretKey string, ttl time.Duration) *authJwtImpl {
	return &authJwtImpl{secretKey: secretKey, ttl: ttl}
}

func (j *authJwtImpl) GenerateToken(id string, role string) (string, error) {
	token, err := jwt.GenerateJWT(newUserClaims(id, role, j.ttl), j.secretKey)
	if err != nil {
		return "", rest.NewInternalServerError(fmt.Errorf("error generating jwt: %w", err))
	}
	
	return token, nil
}
