package jwt

import (
	"time"

	"github.com/hexley21/handy/pkg/jwt"
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
	return jwt.GenerateJWT(newUserClaims(id, role, j.ttl), j.secretKey)
}
