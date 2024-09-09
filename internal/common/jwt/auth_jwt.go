package jwt

import (
	"fmt"
	"time"

	"github.com/hexley21/fixup/internal/common/rest"
	"github.com/hexley21/fixup/pkg/jwt"
)

type Jwt interface {
	JwtGenerator
	JwtVerifier
}

type JwtGenerator interface {
	GenerateJWT(id string, role string, verified bool) (string, error)
}

type JwtVerifier interface {
	VerifyJWT(tokenString string) (UserClaims, error)
}

type authJwtImpl struct {
	secretKey string
	ttl       time.Duration
}

func NewAuthJwtImpl(secretKey string, ttl time.Duration) Jwt {
	return &authJwtImpl{secretKey: secretKey, ttl: ttl}
}

func (j *authJwtImpl) GenerateJWT(id string, role string, verified bool) (string, error) {
	token, err := jwt.GenerateJWT(newClaims(id, role, verified, j.ttl), j.secretKey)
	if err != nil {
		return "", rest.NewInternalServerError(fmt.Errorf("error generating jwt: %w", err))
	}
	
	return token, nil
}

func (j *authJwtImpl) VerifyJWT(tokenString string) (UserClaims, error) {
	mapClaims, err := jwt.VerifyJWT(tokenString, j.secretKey)
	if err != nil {
		return UserClaims{}, err
	}

	return mapToClaim(mapClaims), nil
}