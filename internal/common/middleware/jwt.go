package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"

	authjwt "github.com/hexley21/handy/internal/common/jwt"
	"github.com/hexley21/handy/pkg/jwt"
	"github.com/hexley21/handy/pkg/rest"
)

func EchoJWTMiddleware(secretKey string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return rest.NewUnauthorizedError(nil, "Authorization header is missing")
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				return rest.NewUnauthorizedError(nil, "Bearer token is missing")
			}

			claims, err := jwt.VerifyJWT(tokenString, secretKey)
			if err != nil {
				return rest.NewUnauthorizedError(err, "invalid token")
			}

			c.Set("user", authjwt.MapToClaim(claims))

			return next(c)
		}
	}
}
