package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"

	authjwt "github.com/hexley21/fixup/internal/common/jwt"
	"github.com/hexley21/fixup/internal/common/rest"
	"github.com/hexley21/fixup/internal/common/util/ctxutil"
	"github.com/hexley21/fixup/pkg/jwt"
)

func JWT(secretKey string) echo.MiddlewareFunc {
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

			mapClaims, err := jwt.VerifyJWT(tokenString, secretKey)
			if err != nil {
				return rest.NewUnauthorizedError(err, "Invalid token")
			}

			claims := authjwt.MapToClaim(mapClaims)

			if !claims.Role.Valid() {
				return rest.NewUnauthorizedError(nil, "Invalid token")
			}

			ctxutil.SetJwtId(c, claims.ID)
			ctxutil.SetJwtRole(c, claims.Role)

			return next(c)
		}
	}
}
