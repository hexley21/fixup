package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/labstack/echo/v4"

	"github.com/hexley21/handy/internal/user/delivery/http/v1/dto"
	"github.com/hexley21/handy/internal/user/delivery/http/v1/jwt"
	"github.com/hexley21/handy/internal/user/service"
	"github.com/hexley21/handy/pkg/rest"
)

type authHandler struct {
	service service.AuthService
	authJwt jwt.AuthJwt
}

func NewAuthHandler(service service.AuthService, authJwt jwt.AuthJwt) *authHandler {
	return &authHandler{
		service,
		authJwt,
	}
}

func (h *authHandler) RegisterCustomer() echo.HandlerFunc {
	return func(c echo.Context) error {
		var dto dto.RegisterUser
		if err := c.Bind(&dto); err != nil {
			return err
		}

		user, err := h.service.RegisterCustomer(context.Background(), &dto)

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return rest.Conflict
		}
		if err != nil {
			return err
		}

		accessCookie := new(http.Cookie)
		refreshCookie := new(http.Cookie)

		accessCookie.Name = "access_token"
		refreshCookie.Name = "refresh_token"

		accessToken, err := h.authJwt.GenerateAccessKey(user.ID, user.Role)
		if err != nil {
			return err
		}
		refreshToken, err := h.authJwt.GenerateRefreshKey(user.ID, user.Role)
		if err != nil {
			return err
		}

		accessCookie.Value = accessToken
		refreshCookie.Value = refreshToken

		c.SetCookie(accessCookie)
		c.SetCookie(refreshCookie)
		return c.NoContent(http.StatusOK)
	}
}
