package jwt

import (
	"github.com/golang-jwt/jwt/v5"
)

// Generate creates a JWT token with the given claims and signs it using the provided secret key.
// It returns the signed token string or an error if signing fails.
func Generate[T jwt.Claims](claims T, secretKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// Verify parses and validates a JWT using the provided secret key.
// It returns the token claims if valid, or an error if the token is invalid.
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
