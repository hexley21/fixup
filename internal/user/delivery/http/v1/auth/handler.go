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
	"github.com/hexley21/fixup/pkg/http/handler"
	"github.com/hexley21/fixup/pkg/http/rest"
	"github.com/hexley21/fixup/pkg/infra/postgres/pg_error"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redis/go-redis/v9"
)

// TODO: refactor refresh logic, to include actual user data

const (
	access_token_cookie  = "access_token"
	refresh_token_cookie = "refresh_token"
)

const (
	MsgUserAlreadyExists    = "User already exists"
	MsgUserAlreadyActivated = "User already activated"
	MsgTokenAlreadyUsed     = "Activation token already used"
	MsgIncorrectEmailOrPass = "Email or Password is incorrect"
)

type HandlerFactory struct {
	*handler.Components
	service service.AuthService
}

func NewFactory(components *handler.Components, service service.AuthService) *HandlerFactory {
	return &HandlerFactory{
		Components: components,
		service:    service,
	}
}

func setCookies(w http.ResponseWriter, token string, cookieName string) {
	cookie := http.Cookie{
		Name:     cookieName,
		Value:    token,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
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
		if err := f.Binder.BindJSON(r, dto); err != nil {
			f.Writer.WriteError(w, err)
			return
		}

		if err := f.Validator.Validate(dto); err != nil {
			f.Writer.WriteError(w, err)
			return
		}

		user, err := f.service.RegisterCustomer(r.Context(), *dto)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
				f.Writer.WriteError(w, rest.NewConflictError(err, MsgUserAlreadyExists))
				return
			}
			f.Writer.WriteError(w, rest.NewInternalServerError(err))
			return
		}

		go f.sendConfirmationLetter(r.Context(), verGenerator, user.ID, user.Email, user.FirstName)

		f.Logger.Infof("Register customer - Email: %s, U-ID: %d", user.Email, user.ID)
		f.Writer.WriteNoContent(w, http.StatusCreated)
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
		if err := f.Binder.BindJSON(r, dto); err != nil {
			f.Writer.WriteError(w, err)
			return
		}

		if err := f.Validator.Validate(dto); err != nil {
			f.Writer.WriteError(w, err)
			return
		}

		user, err := f.service.RegisterProvider(r.Context(), *dto)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
				f.Writer.WriteError(w, rest.NewConflictError(err, MsgUserAlreadyExists))
				return
			}

			f.Writer.WriteError(w, rest.NewInternalServerError(err))
			return
		}

		go f.sendConfirmationLetter(r.Context(), verGenerator, user.ID, user.Email, user.FirstName)

		f.Logger.Infof("Register provider - Email: %s, U-ID: %d", user.Email, user.ID)
		f.Writer.WriteNoContent(w, http.StatusCreated)
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
		if err := f.Binder.BindJSON(r, dto); err != nil {
			f.Writer.WriteError(w, err)
			return
		}

		if err := f.Validator.Validate(dto); err != nil {
			f.Writer.WriteError(w, err)
			return
		}

		details, err := f.service.GetUserConfirmationDetails(r.Context(), dto.Email)
		if err != nil {
			if errors.Is(err, pg_error.ErrNotFound) {
				f.Writer.WriteError(w, rest.NewNotFoundError(err, app_error.MsgUserNotFound))
				return
			}
			f.Writer.WriteError(w, rest.NewInternalServerError(err))
			return
		}

		if details.UserStatus {
			f.Writer.WriteError(w, rest.NewConflictError(err, MsgUserAlreadyActivated))
			return
		}

		if err := f.sendConfirmationLetter(r.Context(), verGenerator, details.ID, dto.Email, details.Firstname); err != nil {
			f.Writer.WriteError(w, err)
			return
		}

		f.Logger.Infof("Resend user confirmation letter - Email %s, U-ID: %d", dto.Email, details.ID)
		f.Writer.WriteNoContent(w, http.StatusNoContent)
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
		if err := f.Binder.BindJSON(r, dto); err != nil {
			f.Writer.WriteError(w, err)
			return
		}

		if err := f.Validator.Validate(dto); err != nil {
			f.Writer.WriteError(w, err)
			return
		}

		user, err := f.service.AuthenticateUser(r.Context(), *dto)
		if err != nil {
			if errors.Is(err, hasher.ErrPasswordMismatch) {
				f.Writer.WriteError(w, rest.NewUnauthorizedError(err, MsgIncorrectEmailOrPass))
				return
			}

			if errors.Is(err, pgx.ErrNoRows) {
				f.Writer.WriteError(w, rest.NewNotFoundError(err, app_error.MsgUserNotFound))
				return
			}
			f.Writer.WriteError(w, rest.NewInternalServerError(err))
			return
		}

		accessToken, jWTErr := accessGenerator.GenerateJWT(user.ID, user.Role, user.UserStatus)
		if jWTErr != nil {
			f.Writer.WriteError(w, jWTErr)
			return
		}
		refreshToken, jWTErr := refreshGenerator.GenerateJWT(user.ID, user.Role, user.UserStatus)
		if jWTErr != nil {
			f.Writer.WriteError(w, jWTErr)
			return
		}

		setCookies(w, accessToken, access_token_cookie)
		setCookies(w, refreshToken, refresh_token_cookie)

		f.Logger.Infof("Login user - Role: %s, U-ID: %d", user.Role, user.ID)
		f.Writer.WriteNoContent(w, http.StatusOK)
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

	f.Logger.Info("Logout user")
	f.Writer.WriteNoContent(w, http.StatusOK)
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
			f.Writer.WriteError(w, err)
			return
		}

		role, err := ctx_util.GetJWTRole(r.Context())
		if err != nil {
			f.Writer.WriteError(w, err)
			return
		}

		userStatus, err := ctx_util.GetJWTUserStatus(r.Context())
		if err != nil {
			f.Writer.WriteError(w, err)
			return
		}

		accessToken, err := accessGenerator.GenerateJWT(id, string(role), userStatus)
		if err != nil {
			f.Writer.WriteError(w, err)
			return
		}

		setCookies(w, accessToken, access_token_cookie)

		f.Logger.Infof("Rotate JWT - Role: %s, UserStatus: %v, U-ID: %d", role, userStatus, id)
		f.Writer.WriteNoContent(w, http.StatusOK)
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
			f.Writer.WriteError(w, errResp)
			return
		}

		id, err := strconv.ParseInt(claims.ID, 10, 64)
		if err != nil {
			f.Writer.WriteError(w, rest.NewInternalServerError(err))
			return
		}

		if err := f.service.VerifyUser(r.Context(), tokenParam, time.Until(claims.ExpiresAt.Time), id, claims.Email); err != nil {
			if errors.Is(err, redis.TxFailedErr) {
				f.Writer.WriteError(w, rest.NewConflictError(err, MsgTokenAlreadyUsed))
				return
			}

			if errors.Is(err, pg_error.ErrNotFound) {
				f.Writer.WriteError(w, rest.NewNotFoundError(err, app_error.MsgUserNotFound))
				return
			}

			f.Writer.WriteError(w, rest.NewInternalServerError(err))
			return
		}

		go func() {
			if err := f.service.SendVerifiedLetter(claims.Email); err != nil {
				f.Logger.Errorf("Fail send verified letter - Email: %s, U-ID: %d - cause: %v", claims.Email, id, err)
				return
			}

			f.Logger.Infof("Send verified letter - Email: %s, U-ID: %d", claims.Email, id)
		}()

		f.Logger.Infof("Verify user - Email: %s, U-ID: %d", claims.Email, id)
		f.Writer.WriteNoContent(w, http.StatusOK)
	}
}

func (f *HandlerFactory) sendConfirmationLetter(ctx context.Context, verGenerator verifier.JWTGenerator, id string, email string, name string) *rest.ErrorResponse {
	jWT, err := verGenerator.GenerateJWT(id, email)
	if err != nil {
		f.Logger.Error(err.Error())
		return err
	}

	if err := f.service.SendConfirmationLetter(ctx, jWT, email, name); err != nil {
		errResp := rest.NewInternalServerError(err)
		f.Logger.Error(errResp.Error())
		return errResp
	}

	f.Logger.Infof("Send confirmation letter - Email: %s, U-ID: %d", email, id)
	return nil
}
