package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/labstack/echo/v4"

	"github.com/hexley21/handy/internal/common/jwt"
	"github.com/hexley21/handy/internal/user/delivery/http/v1/dto"
	"github.com/hexley21/handy/internal/user/service"
	"github.com/hexley21/handy/pkg/rest"
)

type authHandler struct {
	service          service.AuthService
	accessGenerator  jwt.AuthJwtGenerator
	refreshGenerator jwt.AuthJwtGenerator
}

func NewAuthHandler(service service.AuthService, accessGenerator jwt.AuthJwtGenerator, refreshGenerator jwt.AuthJwtGenerator) *authHandler {
	return &authHandler{
		service,
		accessGenerator,
		refreshGenerator,
	}
}

func (h *authHandler) setCookies(c echo.Context, userId string, role string) error {
	accessToken, err := h.accessGenerator.GenerateToken(userId, role)
	if err != nil {
		return rest.NewInvalidArgumentsError(err)
	}
	refreshToken, err := h.refreshGenerator.GenerateToken(userId, role)
	if err != nil {
		return rest.NewInvalidArgumentsError(err)
	}

	accessCookie := http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Secure:   true,
		HttpOnly: true,
	}

	refreshCookie := http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Secure:   true,
		HttpOnly: true,
	}

	c.SetCookie(&accessCookie)
	c.SetCookie(&refreshCookie)
	return nil
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
			return rest.NewConflictError(err, "User already exists")
		}
		if err != nil {
			return rest.NewInternalServerError(err)
		}

		h.setCookies(c, user.ID, user.Role)

		return c.NoContent(http.StatusOK)
	}
}

func (h *authHandler) RegisterProvider() echo.HandlerFunc {
	return func(c echo.Context) error {
		var dto dto.RegisterProvider
		if err := c.Bind(&dto); err != nil {
			return err
		}

		user, err := h.service.RegisterProvider(context.Background(), &dto)

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return rest.NewConflictError(err, "User already exists")
		}
		if err != nil {
			return rest.NewInternalServerError(err)
		}

		h.setCookies(c, user.ID, user.Role)

		return c.NoContent(http.StatusOK)
	}
}
