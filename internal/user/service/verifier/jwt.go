package verifier

import (
	"fmt"
	"time"

	"github.com/hexley21/fixup/pkg/jwt"
)

type Jwt interface {
	JwtGenerator
	JwtVerifier
}

type JwtGenerator interface {
	GenerateJWT(id string, email string) (string, error)
}

type JwtVerifier interface {
	VerifyJWT(tokenString string) (VerifyClaims, error)
}

type verificationJwtImpl struct {
	secretKey string
	ttl       time.Duration
}

func NewVerificationJwt(secretKey string, ttl time.Duration) *verificationJwtImpl {
	return &verificationJwtImpl{secretKey, ttl}
}

func (j *verificationJwtImpl) GenerateJWT(id string, email string) (string, error) {
	token, err := jwt.GenerateJWT(newClaims(id, email, j.ttl), j.secretKey)
	if err != nil {
		return "", fmt.Errorf("error generating jwt: %w", err)
	}

	return token, nil
}

func (j *verificationJwtImpl) VerifyJWT(tokenString string) (VerifyClaims, error) {
	mapClaims, err := jwt.VerifyJWT(tokenString, j.secretKey)
	if err != nil {
		return VerifyClaims{}, err
	}

	return mapToClaim(mapClaims), nil
}
