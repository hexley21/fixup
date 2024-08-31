package jwt

import (
	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWT[T jwt.Claims](claims T, secretKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

func VerifyJWT(tokenString string, secretKey string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		return nil, err
	}

	if mapClaims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return mapClaims, nil
	}

	return nil, jwt.ErrInvalidKey
}
