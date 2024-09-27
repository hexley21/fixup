package auth

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/hexley21/fixup/internal/common/app_error"
	"github.com/hexley21/fixup/internal/common/auth_jwt"
	"github.com/hexley21/fixup/internal/common/util/ctx_util"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/user/service"
	"github.com/hexley21/fixup/internal/user/service/verifier"
	"github.com/hexley21/fixup/pkg/hasher"
	"github.com/hexley21/fixup/pkg/http/binder"
	"github.com/hexley21/fixup/pkg/http/rest"
	"github.com/hexley21/fixup/pkg/http/writer"
	"github.com/hexley21/fixup/pkg/infra/postgres/pg_error"
	"github.com/hexley21/fixup/pkg/logger"
	"github.com/hexley21/fixup/pkg/validator"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	access_token_cookie  = "access_token"
	refresh_token_cookie = "refresh_token"
)

const (
	MsgUserAlreadyExists = "User already exists"
	MsgIncorrectEmailOrPass = "Email or Password is incorrect"
)

type HandlerFactory struct {
	logger    logger.Logger
	binder    binder.JSONBinder
	validator validator.Validator
	writer    writer.HTTPWriter
	service   service.AuthService
}

func NewFactory(logger logger.Logger, binder binder.JSONBinder, validator validator.Validator, writer writer.HTTPWriter, service service.AuthService) *HandlerFactory {
	return &HandlerFactory{
		logger,
		binder,
		validator,
		writer,
		service,
	}
}

func setCookies(w http.ResponseWriter, token string, cookieName string) {
	cookie := http.Cookie{
		Name:     cookieName,
		Value:    token,
		Secure:   true,
		HttpOnly: true,
	}

	http.SetCookie(w, &cookie)
}

func eraseCookie(w http.ResponseWriter, cookieName string) {
	cookie := http.Cookie{
		Name:     cookieName,
		Value:    "",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
	}

	http.SetCookie(w, &cookie)
}


// @Summary Register a new customer
// @Description Register a new customer with the provided details
// @Tags auth
// @Accept json
// @Produce json
// @Param user body dto.RegisterUser true "User registration details"
// @Success 201
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 409 {object} rest.ErrorResponse "Conflict - User already exists"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /auth/register/customer [post]
func (f *HandlerFactory) RegisterCustomer(
	verGenerator verifier.JWTGenerator,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dto := new(dto.RegisterUser)
		if err := f.binder.BindJSON(r, dto); err != nil {
			f.writer.WriteError(w, err)
			return
		}

		if err := f.validator.Validate(dto); err != nil {
			f.writer.WriteError(w, err)
			return
		}

		user, err := f.service.RegisterCustomer(context.Background(), *dto)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
				f.writer.WriteError(w, rest.NewConflictError(err, MsgUserAlreadyExists))
				return
			}
			f.writer.WriteError(w, rest.NewInternalServerError(err))
			return
		}

		go f.sendConfirmationLetter(verGenerator, user.ID, user.Email, user.FirstName)

		f.logger.Infof("Customer was registered with ID: %s, Email: %s", user.ID, user.Email)
		f.writer.WriteNoContent(w, http.StatusCreated)
	}
}

// @Summary Register a new provider
// @Description Register a new provider with the provided details
// @Tags auth
// @Accept json
// @Produce json
// @Param user body dto.RegisterProvider true "User registration details"
// @Success 201
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 409 {object} rest.ErrorResponse "Conflict - User already exists"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /auth/register/provider [post]
func (f *HandlerFactory) RegisterProvider(
	verGenerator verifier.JWTGenerator,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dto := new(dto.RegisterProvider)
		if err := f.binder.BindJSON(r, dto); err != nil {
			f.writer.WriteError(w, err)
			return
		}

		if err := f.validator.Validate(dto); err != nil {
			f.writer.WriteError(w, err)
			return
		}

		user, err := f.service.RegisterProvider(context.Background(), *dto)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
				f.writer.WriteError(w, rest.NewConflictError(err, MsgUserAlreadyExists))
				return
			}

			f.writer.WriteError(w, rest.NewInternalServerError(err))
			return
		}

		go f.sendConfirmationLetter(verGenerator, user.ID, user.Email, user.FirstName)

		f.logger.Infof("Provider was registered with ID: %s, Email: %s", user.ID, user.Email)
		f.writer.WriteNoContent(w, http.StatusCreated)
	}
}

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
func (f *HandlerFactory) ResendConfirmationLetter(
	verGenerator verifier.JWTGenerator,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dto := new(dto.Email)
		if err := f.binder.BindJSON(r, dto); err != nil {
			f.writer.WriteError(w, err)
			return
		}

		if err := f.validator.Validate(dto); err != nil {
			f.writer.WriteError(w, err)
			return 
		}

		details, err := f.service.GetUserConfirmationDetails(context.Background(), dto.Email)
		if err != nil {
			if errors.Is(err, service.ErrUserAlreadyActive) {
				f.writer.WriteError(w, rest.NewConflictError(err, MsgUserAlreadyExists))
				return
			}
			if errors.Is(err, pg_error.ErrNotFound) {
				f.writer.WriteError(w, rest.NewNotFoundError(err, app_error.MsgUserNotFound))
				return
			}
			f.writer.WriteError(w, rest.NewInternalServerError(err))
			return
		}

		if err := f.sendConfirmationLetter(verGenerator, details.ID, dto.Email, details.Firstname); err != nil {
			f.writer.WriteError(w, rest.NewInternalServerError(err))
			return 
		}

		f.logger.Infof("Confirmation letter was resent to user with email: %s, ID: %s", dto.Email, details.ID)
		f.writer.WriteNoContent(w, http.StatusNoContent)
	}
}

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
func (f *HandlerFactory) Login(
	accessGenerator auth_jwt.JWTGenerator,
	refreshGenerator auth_jwt.JWTGenerator,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dto := new(dto.Login)
		if err := f.binder.BindJSON(r, dto); err != nil {
			f.writer.WriteError(w, err)
			return
		}

		if err := f.validator.Validate(dto); err != nil {
			f.writer.WriteError(w, err)
			return
		}

		user, err := f.service.AuthenticateUser(context.Background(), *dto)
		if err != nil {
			if errors.Is(err, hasher.ErrPasswordMismatch) {
				f.writer.WriteError(w, rest.NewUnauthorizedError(err, MsgIncorrectEmailOrPass))
				return
			}
			f.writer.WriteError(w, rest.NewInternalServerError(err))
			return
		}

		accessToken, jWTErr := accessGenerator.GenerateJWT(user.ID, user.Role, user.UserStatus)
		if jWTErr != nil {
			f.writer.WriteError(w, jWTErr)
			return
		}
		refreshToken, jWTErr := refreshGenerator.GenerateJWT(user.ID, user.Role, user.UserStatus)
		if jWTErr != nil {
			f.writer.WriteError(w, jWTErr)
			return
		}

		setCookies(w, accessToken, access_token_cookie)
		setCookies(w, refreshToken, refresh_token_cookie)

		f.logger.Infof("User logged in, user ID: %d, role: %s", user.ID, user.Role)
		f.writer.WriteNoContent(w, http.StatusOK)
	}
}

// @Summary Logout a user
// @Description Erase access and refresh tokens
// @Tags auth
// @Success 200 {string} string "Set-Cookie: access_token; HttpOnly, Set-Cookie: refresh_token; HttpOnly"
// @Router /auth/logout [post]
func (f *HandlerFactory) Logout(w http.ResponseWriter, r *http.Request) {
	eraseCookie(w, access_token_cookie)
	eraseCookie(w, refresh_token_cookie)

	f.logger.Info("User logged out, cookies erased")
	f.writer.WriteNoContent(w, http.StatusOK)
}

// @Summary Refresh access token
// @Description Refresh the access token using the refresh token
// @Tags auth
// @Success 200 {string} string "Set-Cookie: access_token; HttpOnly"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Security refresh_token
// @Router /auth/refresh [post]
func (f *HandlerFactory) Refresh(
	accessGenerator auth_jwt.JWTGenerator,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := ctx_util.GetJWTId(r.Context())
		if err != nil {
			f.writer.WriteError(w, err)
			return
		}

		role, err := ctx_util.GetJWTRole(r.Context())
		if err != nil {
			f.writer.WriteError(w, err)
			return
		}

		userStatus, err := ctx_util.GetJWTUserStatus(r.Context())
		if err != nil {
			f.writer.WriteError(w, err)
			return
		}

		accessToken, err := accessGenerator.GenerateJWT(id, string(role), userStatus)
		if err != nil {
			f.writer.WriteError(w, err)
			return
		}

		setCookies(w, accessToken, access_token_cookie)

		f.logger.Infof("JWT refreshed for user ID: %s, Role: %s, UserStatus: %s", id, role, userStatus)
		f.writer.WriteNoContent(w, http.StatusOK)
	}
}

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
func (f *HandlerFactory) VerifyEmail(
	jWTVerifier verifier.JWTVerifier,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenParam := r.URL.Query().Get("token")
		claims, errResp := jWTVerifier.VerifyJWT(tokenParam)
		if errResp != nil {
			f.writer.WriteError(w, errResp)
			return
		}

		id, err := strconv.ParseInt(claims.ID, 10, 64)
		if err != nil {
			f.writer.WriteError(w, rest.NewInternalServerError(err))
			return
		}

		if err := f.service.VerifyUser(context.Background(), id, claims.Email); err != nil {
			if errors.Is(err, pg_error.ErrNotFound) {
				f.writer.WriteError(w, rest.NewNotFoundError(err, app_error.MsgUserNotFound))
				return
			}
			f.writer.WriteError(w, rest.NewInternalServerError(err))
			return
		}

		go func() {
			if err := f.service.SendVerifiedLetter(claims.Email); err != nil {
				f.logger.Errorf("Failed to send verified letter to Email: %s, user ID: %s - cause: %w", claims.Email, id, err)
				return
			}

			f.logger.Infof("Verified letter sent to Email: %s, user ID: %s", claims.Email, id)
		}()

		f.logger.Infof("Email verification successful for user ID: %d, Email: %s", id, claims.Email)
		f.writer.WriteNoContent(w, http.StatusOK)
	}
}

func (f *HandlerFactory) sendConfirmationLetter(verGenerator verifier.JWTGenerator, id string, email string, name string) error {
	jWT, err := verGenerator.GenerateJWT(id, email)
	if err == nil {
		if err := f.service.SendConfirmationLetter(jWT, email, name); err == nil {
			f.logger.Infof("Confirmation letter sent to email: %s, user ID: %s", email, id)
			return nil
		}
	}

	f.logger.Error(err.Error())
	return err
}
