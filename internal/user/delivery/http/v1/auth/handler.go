package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/labstack/echo/v4"

	"github.com/hexley21/fixup/internal/common/jwt"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/user/service"
	"github.com/hexley21/fixup/pkg/rest"
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

func setCookies(c echo.Context, jwtGenerator jwt.AuthJwtGenerator, cookieName string, userId string, role string) error {
	token, err := jwtGenerator.GenerateToken(userId, role)
	if err != nil {
		return rest.NewInvalidArgumentsError(err)
	}

	cookie := http.Cookie{
		Name:     cookieName,
		Value:    token,
		Secure:   true,
		HttpOnly: true,
	}

	c.SetCookie(&cookie)
	return nil
}

func (h *authHandler) registerCustomer(c echo.Context) error {
	dto := new(dto.RegisterUser)
	if err := c.Bind(dto); err != nil {
		return err
	}

	if err := c.Validate(dto); err != nil {
		return rest.NewInvalidArgumentsError(err)
	}

	user, err := h.service.RegisterCustomer(context.Background(), *dto)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return rest.NewConflictError(err, "User already exists")
	}
	if err != nil {
		return rest.NewInternalServerError(err)
	}

	setCookies(c, h.accessGenerator, "access_token", user.ID, user.Role)
	setCookies(c, h.refreshGenerator, "refresh_token", user.ID, user.Role)

	return c.NoContent(http.StatusOK)
}

func (h *authHandler) registerProvider(c echo.Context) error {
	dto := new(dto.RegisterProvider)
	if err := c.Bind(dto); err != nil {
		return err
	}

	if err := c.Validate(dto); err != nil {
		return rest.NewInvalidArgumentsError(err)
	}

	user, err := h.service.RegisterProvider(context.Background(), *dto)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return rest.NewConflictError(err, "User already exists")
	}
	if err != nil {
		return rest.NewInternalServerError(err)
	}

	setCookies(c, h.accessGenerator, "access_token", user.ID, user.Role)
	setCookies(c, h.refreshGenerator, "refresh_token", user.ID, user.Role)

	return c.NoContent(http.StatusOK)
}

func (h *authHandler) Login(c echo.Context) error {
	dto := new(dto.Login)
	if err := c.Bind(dto); err != nil {
		return err
	}

	user, err := h.service.AuthenticateUser(context.Background(), *dto)
	if err != nil {
		return rest.NewUnauthorizedError(err, "Incorrect email or password")
	}

	setCookies(c, h.accessGenerator, "access_token", user.ID, user.Role)
	setCookies(c, h.refreshGenerator, "refresh_token", user.ID, user.Role)
	return c.NoContent(http.StatusOK)
}

func (h *authHandler) Refresh(c echo.Context) error {
	user, ok := c.Get("user").(jwt.UserClaims)
	if !ok {
		return rest.ErrJwtNotImplemented
	}

	setCookies(c, h.accessGenerator, "access_token", user.ID, string(user.Role))
	return nil
}
