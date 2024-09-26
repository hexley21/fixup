package jwt

import (
	"time"

	"github.com/hexley21/fixup/internal/common/app_error"
	"github.com/hexley21/fixup/pkg/http/rest"
	"github.com/hexley21/fixup/pkg/jwt"
)

type Jwt interface {
	JwtGenerator
	JwtVerifier
}

type JwtGenerator interface {
	GenerateJWT(id string, role string, verified bool) (string, *rest.ErrorResponse)
}

type JwtVerifier interface {
	VerifyJWT(tokenString string) (UserClaims, *rest.ErrorResponse)
}

type authJwtImpl struct {
	secretKey string
	ttl       time.Duration
}

func NewAuthJwtImpl(secretKey string, ttl time.Duration) Jwt {
	return &authJwtImpl{secretKey: secretKey, ttl: ttl}
}

func (j *authJwtImpl) GenerateJWT(id string, role string, verified bool) (string, *rest.ErrorResponse) {
	token, err := jwt.GenerateJWT(NewClaims(id, role, verified, j.ttl), j.secretKey)
	if err != nil {
		return "", rest.NewInternalServerError(err)
	}
	
	return token, nil
}

func (j *authJwtImpl) VerifyJWT(tokenString string) (UserClaims, *rest.ErrorResponse) {
	mapClaims, err := jwt.VerifyJWT(tokenString, j.secretKey)
	if err != nil {
		return UserClaims{}, rest.NewUnauthorizedError(err, app_error.MsgInvalidToken)
	}

	return mapToClaim(mapClaims), nil
}