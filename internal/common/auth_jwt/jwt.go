package auth_jwt

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
	Generate(id string, role string, verified bool) (string, *rest.ErrorResponse)
}

type Verifier interface {
	Verify(tokenString string) (UserClaims, *rest.ErrorResponse)
}

type managerImpl struct {
	secretKey string
	ttl       time.Duration
}

func NewManager(secretKey string, ttl time.Duration) *managerImpl {
	return &managerImpl{secretKey: secretKey, ttl: ttl}
}

func (j *managerImpl) Generate(id string, role string, verified bool) (string, *rest.ErrorResponse) {
	token, err := jwt.Generate(NewClaims(id, role, verified, j.ttl), j.secretKey)
	if err != nil {
		return "", rest.NewInternalServerError(err)
	}
	
	return token, nil
}

func (j *managerImpl) Verify(tokenString string) (UserClaims, *rest.ErrorResponse) {
	mapClaims, err := jwt.Verify(tokenString, j.secretKey)
	if err != nil {
		return UserClaims{}, rest.NewUnauthorizedError(err, app_error.MsgInvalidToken)
	}

	return mapToClaim(mapClaims), nil
}