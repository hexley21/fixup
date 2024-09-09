package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/labstack/echo/v4"

	"github.com/hexley21/fixup/internal/common/jwt"
	"github.com/hexley21/fixup/internal/common/rest"
	"github.com/hexley21/fixup/internal/common/util/ctxutil"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/user/service"
)

var (
	access_token_cookie  = "access_token"
	refresh_token_cookie = "refresh_token"
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
		return err
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

func eraseCookie(c echo.Context, cookieName string) {
	cookie := http.Cookie{
		Name:     cookieName,
		Value:    "",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
	}

	c.SetCookie(&cookie)
}

// registerCustomer godoc
// @Summary Register a new customer
// @Description Register a new customer with the provided details
// @Tags auth
// @Accept json
// @Produce json
// @Param user body dto.RegisterUser true "User registration details"
// @Success 200 {string} string "Set-Cookie: access_token; HttpOnly, Set-Cookie: refresh_token; HttpOnly"
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 409 {object} rest.ErrorResponse "Conflict - User already exists"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /auth/register/customer [post]
func (h *authHandler) registerCustomer(c echo.Context) error {
	dto := new(dto.RegisterUser)
	if err := c.Bind(dto); err != nil {
		return rest.NewInternalServerError(err)
	}

	if err := c.Validate(dto); err != nil {
		return rest.NewInvalidArgumentsError(err)
	}

	user, err := h.service.RegisterCustomer(context.Background(), *dto)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return rest.NewConflictError(err, "User already exists")
			}
		}
		return rest.NewInternalServerError(err)
	}

	setCookies(c, h.accessGenerator, access_token_cookie, user.ID, user.Role)
	setCookies(c, h.refreshGenerator, refresh_token_cookie, user.ID, user.Role)

	return c.NoContent(http.StatusOK)
}

// registerProvider godoc
// @Summary Register a new provider
// @Description Register a new provider with the provided details
// @Tags auth
// @Accept json
// @Produce json
// @Param user body dto.RegisterProvider true "User registration details"
// @Success 200 {string} string "Set-Cookie: access_token; HttpOnly, Set-Cookie: refresh_token; HttpOnly"
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 409 {object} rest.ErrorResponse "Conflict - User already exists"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /auth/register/provider [post]
func (h *authHandler) registerProvider(c echo.Context) error {
	dto := new(dto.RegisterProvider)
	if err := c.Bind(dto); err != nil {
		return rest.NewInternalServerError(err)
	}

	if err := c.Validate(dto); err != nil {
		return rest.NewInvalidArgumentsError(err)
	}

	user, err := h.service.RegisterProvider(context.Background(), *dto)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return rest.NewConflictError(err, "User already exists")
			}
		}
		return rest.NewInternalServerError(err)
	}

	setCookies(c, h.accessGenerator, access_token_cookie, user.ID, user.Role)
	setCookies(c, h.refreshGenerator, refresh_token_cookie, user.ID, user.Role)

	return c.NoContent(http.StatusOK)
}

// login godoc
// @Summary Login a user
// @Description Authenticate a user and set access and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param user body dto.Login true "User login details"
// @Success 200 {string} string "Set-Cookie: access_token; HttpOnly, Set-Cookie: refresh_token; HttpOnly"
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized - Incorrect email or password"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /auth/login [post]
func (h *authHandler) login(c echo.Context) error {
	dto := new(dto.Login)
	if err := c.Bind(dto); err != nil {
		return rest.NewInvalidArgumentsError(err)
	}

	user, err := h.service.AuthenticateUser(context.Background(), *dto)
	if err != nil {
		return rest.NewUnauthorizedError(err, "Incorrect email or password")
	}

	setCookies(c, h.accessGenerator, access_token_cookie, user.ID, user.Role)
	setCookies(c, h.refreshGenerator, refresh_token_cookie, user.ID, user.Role)
	return c.NoContent(http.StatusOK)
}

// logout godoc
// @Summary Logout a user
// @Description Erase access and refresh tokens
// @Tags auth
// @Success 200 {string} string "Set-Cookie: access_token; HttpOnly, Set-Cookie: refresh_token; HttpOnly"
// @Router /auth/logout [post]
func (h *authHandler) logout(c echo.Context) error {
	eraseCookie(c, access_token_cookie)
	eraseCookie(c, refresh_token_cookie)

	return c.NoContent(http.StatusOK)
}

// refresh godoc
// @Summary Refresh access token
// @Description Refresh the access token using the refresh token
// @Tags auth
// @Success 200 {string} string "Set-Cookie: access_token; HttpOnly"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /auth/refresh [post]
func (h *authHandler) refresh(c echo.Context) error {
	id, err := ctxutil.GetJwtId(c)
	if err != nil {
		return err
	}

	role, err := ctxutil.GetJwtId(c)
	if err != nil {
		return err
	}

	return setCookies(c, h.accessGenerator, access_token_cookie, id, role)
}
