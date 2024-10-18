package refresh_jwt

import (
	"strconv"
	"time"

	"github.com/hexley21/fixup/pkg/http/rest"
	"github.com/hexley21/fixup/pkg/jwt"
)

type Manager interface {
	Generator
	Verifier
}

type Generator interface {
	Generate(id int64) (string, *rest.ErrorResponse)
}

type Verifier interface {
	Verify(tokenString string) (RefreshClaims, *rest.ErrorResponse)
}

type managerImpl struct {
	secretKey string
	ttl       time.Duration
}

func NewManager(secretKey string, ttl time.Duration) *managerImpl {
	return &managerImpl{secretKey: secretKey, ttl: ttl}
}

func (j *managerImpl) Generate(id int64) (string, *rest.ErrorResponse) {
	token, err := jwt.Generate(NewClaims(strconv.FormatInt(id, 10), j.ttl), j.secretKey)
	if err != nil {
		return "", rest.NewInternalServerError(err)
	}
	
	return token, nil
}

func (j *managerImpl) Verify(tokenString string) (RefreshClaims, *rest.ErrorResponse) {
	mapClaims, err := jwt.Verify(tokenString, j.secretKey)
	if err != nil {
		return RefreshClaims{}, rest.NewUnauthorizedError(err)
	}

	return mapToClaim(mapClaims), nil
}