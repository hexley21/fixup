package auth_jwt

import (
	"time"

	"github.com/hexley21/fixup/internal/common/app_error"
	"github.com/hexley21/fixup/pkg/http/rest"
	"github.com/hexley21/fixup/pkg/jwt"
)

type JWTManager interface {
	JWTGenerator
	JWTVerifier
}

type JWTGenerator interface {
	GenerateJWT(id string, role string, verified bool) (string, *rest.ErrorResponse)
}

type JWTVerifier interface {
	VerifyJWT(tokenString string) (UserClaims, *rest.ErrorResponse)
}

type authJWTImpl struct {
	secretKey string
	ttl       time.Duration
}

func NewJWTManager(secretKey string, ttl time.Duration) *authJWTImpl {
	return &authJWTImpl{secretKey: secretKey, ttl: ttl}
}

func (j *authJWTImpl) GenerateJWT(id string, role string, verified bool) (string, *rest.ErrorResponse) {
	token, err := jwt.GenerateJWT(NewClaims(id, role, verified, j.ttl), j.secretKey)
	if err != nil {
		return "", rest.NewInternalServerError(err)
	}
	
	return token, nil
}

func (j *authJWTImpl) VerifyJWT(tokenString string) (UserClaims, *rest.ErrorResponse) {
	mapClaims, err := jwt.VerifyJWT(tokenString, j.secretKey)
	if err != nil {
		return UserClaims{}, rest.NewUnauthorizedError(err, app_error.MsgInvalidToken)
	}

	return mapToClaim(mapClaims), nil
}