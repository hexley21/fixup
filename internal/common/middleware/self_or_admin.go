package middleware

import (
	"github.com/labstack/echo/v4"

	"github.com/hexley21/handy/internal/common/jwt"
	"github.com/hexley21/handy/internal/user/enum"
	"github.com/hexley21/handy/pkg/rest"
)

func EchoIsSelfOrAdminMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			userClaims, ok := c.Get("user").(jwt.UserClaims)
			if !ok {
				return rest.ErrJwtNotImplemented
			}

			if !(userClaims.ID == c.Param("id") || userClaims.Role == enum.UserRoleADMIN) {
				return rest.ErrInsufficientRights
			}

			return next(c)
		}
	}
}
