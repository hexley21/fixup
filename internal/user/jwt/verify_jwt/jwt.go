package verify_jwt

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
	Generate(id int64, email string) (string, *rest.ErrorResponse)
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

func (j *managerImpl) Generate(id int64, email string) (string, *rest.ErrorResponse) {
	token, err := jwt.Generate(newClaims(strconv.FormatInt(id, 10), email, j.ttl), j.secretKey)
	if err != nil {
		return "", rest.NewInternalServerError(err)
	}

	return token, nil
}

func (j *managerImpl) Verify(tokenString string) (VerifyClaims, *rest.ErrorResponse) {
	mapClaims, err := jwt.Verify(tokenString, j.secretKey)
	if err != nil {
		return VerifyClaims{}, rest.NewUnauthorizedError(err)
	}

	return mapToClaim(mapClaims), nil
}
