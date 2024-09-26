package verifier

import (
	"fmt"
	"time"

	"github.com/hexley21/fixup/pkg/jwt"
)

type JWTManager interface {
	JWTGenerator
	JWTVerifier
}

type JWTGenerator interface {
	GenerateJWT(id string, email string) (string, error)
}

type JWTVerifier interface {
	VerifyJWT(tokenString string) (VerifyClaims, error)
}

type verificationJWTImpl struct {
	secretKey string
	ttl       time.Duration
}

func NewJWTManager(secretKey string, ttl time.Duration) *verificationJWTImpl {
	return &verificationJWTImpl{secretKey, ttl}
}

func (j *verificationJWTImpl) GenerateJWT(id string, email string) (string, error) {
	token, err := jwt.GenerateJWT(newClaims(id, email, j.ttl), j.secretKey)
	if err != nil {
		return "", fmt.Errorf("error generating jwt: %w", err)
	}

	return token, nil
}

func (j *verificationJWTImpl) VerifyJWT(tokenString string) (VerifyClaims, error) {
	mapClaims, err := jwt.VerifyJWT(tokenString, j.secretKey)
	if err != nil {
		return VerifyClaims{}, err
	}

	return mapToClaim(mapClaims), nil
}
