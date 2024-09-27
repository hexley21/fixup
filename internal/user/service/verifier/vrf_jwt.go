package verifier

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
	GenerateJWT(id string, email string) (string, *rest.ErrorResponse)
}

type JWTVerifier interface {
	VerifyJWT(tokenString string) (VerifyClaims, *rest.ErrorResponse)
}

type verificationJWTImpl struct {
	secretKey string
	ttl       time.Duration
}

func NewJWTManager(secretKey string, ttl time.Duration) *verificationJWTImpl {
	return &verificationJWTImpl{secretKey, ttl}
}

func (j *verificationJWTImpl) GenerateJWT(id string, email string) (string, *rest.ErrorResponse) {
	token, err := jwt.GenerateJWT(newClaims(id, email, j.ttl), j.secretKey)
	if err != nil {
		return "", rest.NewInternalServerError(err)
	}

	return token, nil
}

func (j *verificationJWTImpl) VerifyJWT(tokenString string) (VerifyClaims, *rest.ErrorResponse) {
	mapClaims, err := jwt.VerifyJWT(tokenString, j.secretKey)
	if err != nil {
		return VerifyClaims{}, rest.NewUnauthorizedError(err, app_error.MsgInvalidToken)
	}

	return mapToClaim(mapClaims), nil
}
