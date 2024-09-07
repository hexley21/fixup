package middleware

import (
	"github.com/labstack/echo/v4"

	"github.com/hexley21/fixup/internal/common/jwt"
	"github.com/hexley21/fixup/internal/user/enum"
	"github.com/hexley21/fixup/pkg/rest"
)

func EchoIsSelfOrAdminMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			userClaims, ok := c.Get("user").(jwt.UserClaims)
			if !ok {
				return rest.ErrJwtNotImplemented
			}

			idParam := c.Param("id")
			if !(userClaims.ID == idParam || idParam == "me" || userClaims.Role == enum.UserRoleADMIN) {
				return rest.ErrInsufficientRights
			}

			return next(c)
		}
	}
}
