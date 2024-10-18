package auth_jwt

import (
	"strconv"
	"time"

	"github.com/hexley21/fixup/internal/common/enum"
	"github.com/hexley21/fixup/pkg/http/rest"
	"github.com/hexley21/fixup/pkg/jwt"
)

type Manager interface {
	Generator
	Verifier
}

type Generator interface {
	Generate(id int64, role enum.UserRole, verified bool) (string, *rest.ErrorResponse)
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

func (j *managerImpl) Generate(id int64, role enum.UserRole, verified bool) (string, *rest.ErrorResponse) {
	token, err := jwt.Generate(NewClaims(strconv.FormatInt(id, 10), role, verified, j.ttl), j.secretKey)
	if err != nil {
		return "", rest.NewInternalServerError(err)
	}
	
	return token, nil
}

func (j *managerImpl) Verify(tokenString string) (UserClaims, *rest.ErrorResponse) {
	mapClaims, err := jwt.Verify(tokenString, j.secretKey)
	if err != nil {
		return UserClaims{}, rest.NewUnauthorizedError(err)
	}

	return mapToClaim(mapClaims), nil
}