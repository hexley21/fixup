package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"

	authjwt "github.com/hexley21/fixup/internal/common/jwt"
	"github.com/hexley21/fixup/internal/common/rest"
	"github.com/hexley21/fixup/internal/common/util/ctxutil"
)

func JWT(jwtVerifier authjwt.JwtVerifier) echo.MiddlewareFunc {
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

			claims, err := jwtVerifier.VerifyJWT(tokenString)
			if err != nil {
				return rest.NewUnauthorizedError(err, "Invalid token")
			}


			if !claims.Role.Valid() {
				return rest.NewUnauthorizedError(nil, "Invalid token")
			}

			ctxutil.SetJwtId(c, claims.ID)
			ctxutil.SetJwtRole(c, claims.Role)

			return next(c)
		}
	}
}
