package verify_jwt

import (
	"time"

	"github.com/hexley21/fixup/internal/common/app_error"
	"github.com/hexley21/fixup/pkg/http/rest"
	"github.com/hexley21/fixup/pkg/jwt"
)

type Manager interface {
	Generator
	Verifier
}

type Generator interface {
	Generate(id string, email string) (string, *rest.ErrorResponse)
}

type Verifier interface {
	Verify(tokenString string) (VerifyClaims, *rest.ErrorResponse)
}

type managerImpl struct {
	secretKey string
	ttl       time.Duration
}

func NewManager(secretKey string, ttl time.Duration) *managerImpl {
	return &managerImpl{secretKey, ttl}
}

func (j *managerImpl) Generate(id string, email string) (string, *rest.ErrorResponse) {
	token, err := jwt.Generate(newClaims(id, email, j.ttl), j.secretKey)
	if err != nil {
		return "", rest.NewInternalServerError(err)
	}

	return token, nil
}

func (j *managerImpl) Verify(tokenString string) (VerifyClaims, *rest.ErrorResponse) {
	mapClaims, err := jwt.Verify(tokenString, j.secretKey)
	if err != nil {
		return VerifyClaims{}, rest.NewUnauthorizedError(err, app_error.MsgInvalidToken)
	}

	return mapToClaim(mapClaims), nil
}
