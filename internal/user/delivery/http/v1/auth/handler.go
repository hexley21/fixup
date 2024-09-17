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

const (
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
// @Success 204
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 409 {object} rest.ErrorResponse "Conflict - User already exists"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /auth/register/customer [post]
func (h *authHandler) registerCustomer(
	verGenerator verifier.JwtGenerator,
) echo.HandlerFunc {
	return func(c echo.Context) error {
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
					return rest.NewConflictError(err, rest.MsgUserAlreadyExists)
				}
			}
			return rest.NewInternalServerError(err)
		}

		go sendConfirmationLetter(c.Logger(), h.service, verGenerator, user.ID, user.Email, user.FirstName)

		return c.NoContent(http.StatusNoContent)
	}
}

// registerProvider godoc
// @Summary Register a new provider
// @Description Register a new provider with the provided details
// @Tags auth
// @Accept json
// @Produce json
// @Param user body dto.RegisterProvider true "User registration details"
// @Success 204
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 409 {object} rest.ErrorResponse "Conflict - User already exists"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /auth/register/provider [post]
func (h *authHandler) registerProvider(
	verGenerator verifier.JwtGenerator,
) echo.HandlerFunc {
	return func(c echo.Context) error {
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
			if errors.As(err, &pgErr) {
				switch pgErr.Code {
				case pgerrcode.UniqueViolation:
					return rest.NewConflictError(err, rest.MsgUserAlreadyExists)
				}
			}
			return rest.NewInternalServerError(err)
		}

		go sendConfirmationLetter(c.Logger(), h.service, verGenerator, user.ID, user.Email, user.FirstName)

		return c.NoContent(http.StatusNoContent)
	}
}

// resendConfirmationLetter godoc
// @Summary Resent confirmation letter
// @Description Resends a confirmation letter to email
// @Tags auth
// @Accept json
// @Produce json
// @Param user body dto.Email true "User email"
// @Success 204
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 409 {object} rest.ErrorResponse "Conflict"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /auth/resend-confirmation [post]
func (h *authHandler) resendConfirmationLetter(
	verGenerator verifier.JwtGenerator,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		dto := new(dto.Email)
		if err := c.Bind(dto); err != nil {
			return rest.NewInternalServerError(err)
		}

		details, err := h.service.GetUserConfirmationDetails(context.Background(), dto.Email)
		if err != nil {
			if errors.Is(err, service.ErrUserAlreadyActive) {
				return rest.NewConflictError(err, rest.MsgUserAlreadyExists)
			}
			if errors.Is(err, pgx.ErrNoRows) {
				return rest.NewNotFoundError(err, rest.MsgUserNotFound)
			}
			return rest.NewInternalServerError(err)
		}

		err = sendConfirmationLetter(c.Logger(), h.service, verGenerator, details.ID, dto.Email, details.Firstname)
		if err != nil {
			return rest.NewInternalServerError(err)
		}

		return c.NoContent(http.StatusNoContent)
	}
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
			return rest.NewUnauthorizedError(err, rest.MsgIncorrectEmailOrPass)
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

		userStatus, err := ctxutil.GetJwtUserStatus(c)
		if err != nil {
			return err
		}

		accessToken, err := accessGenerator.GenerateJWT(id, string(role), userStatus)
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
			return rest.NewUnauthorizedError(err, rest.MsgInvalidToken)
		}

		id, err := strconv.ParseInt(claims.ID, 10, 64)
		if err != nil {
			return rest.NewInternalServerError(err)
		}

		if err := h.service.VerifyUser(context.Background(), id, claims.Email); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return rest.NewNotFoundError(err, rest.MsgUserNotFound)
			}
			return rest.NewInternalServerError(err)
		}

		go func () {
			if err := h.service.SendVerifiedLetter(claims.Email); err != nil {
				c.Logger().Error(err)
			}
		}()

		return c.NoContent(http.StatusOK)
	}
}

func sendConfirmationLetter(logger echo.Logger, authService service.AuthService, verGenerator verifier.JwtGenerator, id string, email string, name string) error {
	jwt, err := verGenerator.GenerateJWT(id, email)
	if err == nil {
		if err := authService.SendConfirmationLetter(jwt, email, name); err == nil {
			return nil
		}
	}

	logger.Error(err)
	return err
}
