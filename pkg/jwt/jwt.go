package jwt

import (
	"github.com/golang-jwt/jwt/v5"
)

type ClaimsMapper[T any] interface {
	MapToClaim(mapClaims any) T
}

func Generate[T jwt.Claims](claims T, secretKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

func Verify(tokenString string, secretKey string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		return nil, err
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrInvalidKey
	}

	return mapClaims, nil
}
