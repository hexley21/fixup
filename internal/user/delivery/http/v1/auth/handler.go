package auth

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	authjwt "github.com/hexley21/fixup/internal/common/jwt"
	"github.com/hexley21/fixup/internal/common/rest"
	"github.com/hexley21/fixup/internal/common/util/ctxutil"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/user/service"
	"github.com/hexley21/fixup/internal/user/service/verifier"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/labstack/echo/v4"
)

var (
	access_token_cookie  = "access_token"
	refresh_token_cookie = "refresh_token"
)

type authHandler struct {
	service service.AuthService
}

func NewAuthHandler(service service.AuthService) *authHandler {
	return &authHandler{
		service,
	}
}

func setCookies(c echo.Context, token string, cookieName string) error {
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
// @Success 200
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

	_, err := h.service.RegisterCustomer(context.Background(), *dto)
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

	return c.NoContent(http.StatusOK)
}

// TODO: manage already registered actions + consider status

// registerProvider godoc
// @Summary Register a new provider
// @Description Register a new provider with the provided details
// @Tags auth
// @Accept json
// @Produce json
// @Param user body dto.RegisterProvider true "User registration details"
// @Success 200
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

	_, err := h.service.RegisterProvider(context.Background(), *dto)
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
func (h *authHandler) login(
	accessGenerator authjwt.JwtGenerator,
	refreshGenerator authjwt.JwtGenerator,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		dto := new(dto.Login)
		if err := c.Bind(dto); err != nil {
			return rest.NewInvalidArgumentsError(err)
		}

		user, err := h.service.AuthenticateUser(context.Background(), *dto)
		if err != nil {
			return rest.NewUnauthorizedError(err, "Incorrect email or password")
		}

		accessToken, err := accessGenerator.GenerateJWT(user.ID, user.Role, user.UserStatus)
		if err != nil {
			return err
		}
		refreshToken, err := refreshGenerator.GenerateJWT(user.ID, user.Role, user.UserStatus)
		if err != nil {
			return err
		}

		setCookies(c, accessToken, access_token_cookie)
		setCookies(c, refreshToken, refresh_token_cookie)
		return c.NoContent(http.StatusOK)
	}
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
// @Security refresh_token
// @Router /auth/refresh [post]
func (h *authHandler) refresh(
	accessGenerator authjwt.JwtGenerator,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := ctxutil.GetJwtId(c)
		if err != nil {
			return err
		}

		role, err := ctxutil.GetJwtRole(c)
		if err != nil {
			return err
		}

		verification, err := ctxutil.GetJwtVerification(c)
		if err != nil {
			return err
		}

		accessToken, err := accessGenerator.GenerateJWT(id, string(role), verification)
		if err != nil {
			return err
		}
		err = setCookies(c, accessToken, access_token_cookie)
		if err != nil {
			return err
		}

		return c.NoContent(http.StatusOK)
	}
}

// verifyEmail godoc
// @Summary Verify email
// @Description Verifies the email of a user using a JWT token provided as a query parameter.
// @Tags auth
// @Accept json
// @Produce json
// @Param token query string true "JWT token for email verification"
// @Success 200
// @Failure 400 {object} rest.ErrorResponse "Invalid id parameter"
// @Failure 401 {object} rest.ErrorResponse "Invalid token"
// @Failure 404 {object} rest.ErrorResponse "User was not found"
// @Failure 500 {object} rest.ErrorResponse "Internal server error"
// @Router /auth/verify-email [get]
func (h *authHandler) verifyEmail(
	jwtVerifier verifier.JwtVerifier,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, err := jwtVerifier.VerifyJWT(c.QueryParam("token"))
		if err != nil {
			return rest.NewUnauthorizedError(err, "Invalid token")
		}

		id, err := strconv.ParseInt(claims.ID, 10, 64)
		if err != nil {
			return rest.NewInternalServerError(err)
		}

		if err := h.service.VerifyUser(context.Background(), id, claims.Email); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return rest.NewNotFoundError(err, "User was not found")
			}
			return rest.NewInternalServerError(err)
		}

		return c.NoContent(http.StatusOK)
	}
}
