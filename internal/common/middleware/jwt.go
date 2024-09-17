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
				return rest.NewUnauthorizedError(nil, rest.MsgMissingAuthorizationHeader)
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				return rest.NewUnauthorizedError(nil, rest.MsgMissingBearerToken)
			}

			claims, err := jwtVerifier.VerifyJWT(tokenString)
			if err != nil {
				return err
			}

			if !claims.Role.Valid() {
				return rest.NewUnauthorizedError(nil, rest.MsgInvalidToken)
			}

			ctxutil.SetJwtId(c, claims.ID)
			ctxutil.SetJwtRole(c, claims.Role)

			return next(c)
		}
	}
}
