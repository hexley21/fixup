package middleware

import (
	"slices"

	"github.com/hexley21/fixup/internal/common/rest"
	"github.com/hexley21/fixup/internal/common/util/ctxutil"
	"github.com/hexley21/fixup/internal/user/enum"
	"github.com/labstack/echo/v4"
)

var (
	ErrInsufficientRights = rest.NewForbiddenError(nil, rest.MsgInsufficientRights)
	ErrUserVerified       = rest.NewForbiddenError(nil, rest.MsgUserIsVerified)
	ErrUserNotVerified    = rest.NewForbiddenError(nil, rest.MsgUserIsNotVerified)
)

func AllowRoles(roles ...enum.UserRole) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role, err := ctxutil.GetJwtRole(c)
			if err != nil {
				return err
			}

			if !slices.Contains(roles, role) {
				return ErrInsufficientRights
			}

			return next(c)
		}
	}
}

func AllowSelfOrRole(roles ...enum.UserRole) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			idParam := c.Param("id")

			role, err := ctxutil.GetJwtRole(c)
			if err != nil {
				return err
			}

			jwtId, err := ctxutil.GetJwtId(c)
			if err != nil {
				return err
			}

			if idParam == "me" {
				err := ctxutil.SetParamId(c, jwtId)
				if err != nil {
					return err
				}

				return next(c)
			}

			if (idParam == jwtId) || slices.Contains(roles, role) {
				err = ctxutil.SetParamId(c, idParam)
				if err != nil {
					return err
				}

				return next(c)

			}

			return ErrInsufficientRights
		}
	}
}

func AllowVerified(status bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			verified, err := ctxutil.GetJwtUserStatus(c)
			if err != nil {
				return err
			}

			if verified == status {
				return next(c)
			}

			if status {
				return ErrUserNotVerified
			}

			return ErrUserVerified
		}
	}
}
