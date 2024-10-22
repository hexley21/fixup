package auth

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/hexley21/fixup/internal/common/auth_jwt"
	"github.com/hexley21/fixup/internal/common/enum"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/user/domain"
	"github.com/hexley21/fixup/internal/user/jwt/refresh_jwt"
	"github.com/hexley21/fixup/internal/user/jwt/verify_jwt"
	"github.com/hexley21/fixup/internal/user/service"
	"github.com/hexley21/fixup/pkg/http/handler"
	"github.com/hexley21/fixup/pkg/http/rest"
)

const (
	accessTokenCookie  = "access_token"
	refreshTokenCookie = "refresh_token"
)

type Handler struct {
	*handler.Components
	service service.AuthService
}

func NewHandler(components *handler.Components, service service.AuthService) *Handler {
	return &Handler{
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

// RegisterCustomer
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
func (h *Handler) RegisterCustomer(generator verify_jwt.Generator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var registerDTO dto.RegisterUser
		if err := h.Binder.BindJSON(r, &registerDTO); err != nil {
			h.Writer.WriteError(w, err)
			return
		}

		if err := h.Validator.Validate(registerDTO); err != nil {
			h.Writer.WriteError(w, err)
			return
		}

		userEntity, err := h.service.RegisterCustomer(
			r.Context(),
			registerDTO.Password,
			domain.NewUserPersonalInfo(
				registerDTO.Email,
				registerDTO.PhoneNumber,
				registerDTO.FirstName,
				registerDTO.LastName,
			),
		)
		if err != nil {
			if errors.Is(err, service.ErrUserEmailTaken) {
				h.Writer.WriteError(w, rest.NewConflictError(err))
			}

			h.Writer.WriteError(w, rest.NewInternalServerErrorf("failed to register customer: %w", err))
			return
		}

		go h.sendVerificationLetter(
			context.Background(),
			generator, userEntity.ID,
			userEntity.PersonalInfo.Email,
			userEntity.PersonalInfo.FirstName,
		)

		h.Logger.Infof("Register customer - Email: %s, U-ID: %d", userEntity.PersonalInfo.Email, userEntity.ID)
		h.Writer.WriteNoContent(w, http.StatusCreated)
	}
}

// RegisterProvider
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
func (h *Handler) RegisterProvider(generator verify_jwt.Generator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var registerDTO dto.RegisterProvider
		if err := h.Binder.BindJSON(r, &registerDTO); err != nil {
			h.Writer.WriteError(w, err)
			return
		}

		if err := h.Validator.Validate(registerDTO); err != nil {
			h.Writer.WriteError(w, err)
			return
		}

		userEntity, err := h.service.RegisterProvider(
			r.Context(),
			registerDTO.Password,
			registerDTO.PersonalIDNumber,
			domain.NewUserPersonalInfo(
				registerDTO.Email,
				registerDTO.PhoneNumber,
				registerDTO.FirstName,
				registerDTO.LastName,
			),
		)
		if err != nil {
			if errors.Is(err, service.ErrUserEmailTaken) {
				h.Writer.WriteError(w, rest.NewConflictError(err))
			}

			h.Writer.WriteError(w, rest.NewInternalServerErrorf("failed to register provider: %w", err))
			return
		}

		go h.sendVerificationLetter(
			context.Background(),
			generator, userEntity.ID,
			userEntity.PersonalInfo.Email,
			userEntity.PersonalInfo.FirstName,
		)

		h.Logger.Infof("Register provider - Email: %s, U-ID: %d", userEntity.PersonalInfo.Email, userEntity.ID)
		h.Writer.WriteNoContent(w, http.StatusCreated)
	}
}

// ResendVerificationLetter
// @Summary Resent verification letter
// @Description Resends an verification letter to email
// @Tags auth
// @Accept json
// @Produce json
// @Param user body dto.Email true "User email"
// @Success 204
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 409 {object} rest.ErrorResponse "Conflict"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /auth/resend-verification [post]
func (h *Handler) ResendVerificationLetter(generator verify_jwt.Generator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var emailDTO dto.Email
		if err := h.Binder.BindJSON(r, &emailDTO); err != nil {
			h.Writer.WriteError(w, err)
			return
		}

		if err := h.Validator.Validate(emailDTO); err != nil {
			h.Writer.WriteError(w, err)
			return
		}

		tokenFunc := func(id int64) (string, error) {
			jwt, err := generator.Generate(id, emailDTO.Email)
			if err != nil {
				return "", err
			}

			return jwt, nil
		}

		if err := h.service.ResendVerificationLetter(r.Context(), tokenFunc, emailDTO.Email); err != nil {
			var errResp *rest.ErrorResponse
			switch {
			case errors.As(err, &errResp):
				h.Writer.WriteError(w, errResp)
			default:
				h.Writer.WriteError(w, rest.NewInternalServerErrorf("failed to resend verification letter: %w", err))
			}
			return
		}

		h.Logger.Infof("Resend user verification letter - Email %s", emailDTO.Email)
		h.Writer.WriteNoContent(w, http.StatusNoContent)
	}
}

// Login
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
func (h *Handler) Login(generator auth_jwt.Generator, refreshGenerator refresh_jwt.Generator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var loginDTO dto.Login
		if err := h.Binder.BindJSON(r, &loginDTO); err != nil {
			h.Writer.WriteError(w, err)
			return
		}

		if err := h.Validator.Validate(loginDTO); err != nil {
			h.Writer.WriteError(w, err)
			return
		}

		userIdentity, err := h.service.AuthenticateUser(r.Context(), loginDTO.Email, loginDTO.Password)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrIncorrectEmailOrPassword):
				h.Writer.WriteError(w, rest.NewUnauthorizedError(err))
			case errors.Is(err, service.ErrUserNotFound):
				h.Writer.WriteError(w, rest.NewNotFoundMessageError(err, service.ErrIncorrectEmailOrPassword.Error()))
			default:
				h.Writer.WriteError(w, rest.NewInternalServerErrorf("failed to login - uid: %d, error: %w", userIdentity.ID, err))
			}
			return
		}

		accessToken, jwtErr := generator.Generate(
			userIdentity.ID,
			userIdentity.AccountInfo.Role,
			userIdentity.AccountInfo.Verified,
		)
		if jwtErr != nil {
			h.Writer.WriteError(w, jwtErr)
			return
		}
		refreshToken, jwtErr := refreshGenerator.Generate(userIdentity.ID)
		if jwtErr != nil {
			h.Writer.WriteError(w, jwtErr)
			return
		}

		setCookies(w, accessToken, accessTokenCookie)
		setCookies(w, refreshToken, refreshTokenCookie)
		h.Logger.Infof("Login user - Role: %s, U-ID: %d", userIdentity.AccountInfo.Role, userIdentity.ID)
		h.Writer.WriteNoContent(w, http.StatusOK)
	}
}

// Logout
// @Summary Logout a user
// @Description Erase access and refresh tokens
// @Tags auth
// @Success 200 {string} string "Set-Cookie: access_token; HttpOnly, Set-Cookie: refresh_token; HttpOnly"
// @Router /auth/logout [post]
func (h *Handler) Logout(w http.ResponseWriter, _ *http.Request) {
	eraseCookie(w, accessTokenCookie)
	eraseCookie(w, refreshTokenCookie)

	h.Logger.Info("Logout user")
	h.Writer.WriteNoContent(w, http.StatusOK)
}

// Refresh
// @Summary Refresh access token
// @Description Refresh the access token using the refresh token
// @Tags auth
// @Success 200 {string} string "Set-Cookie: access_token; HttpOnly"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Security refresh_token
// @Router /auth/refresh [post]
func (h *Handler) Refresh(generator auth_jwt.Generator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := r.Context().Value(refreshJwtIdKet).(string)
		if !ok {
			h.Writer.WriteError(w, ErrRefreshTokenNotSet)
			return
		}

		intId, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			h.Writer.WriteError(w, rest.NewInvalidArgumentsError(err))
			return
		}

		tokenFunc := func(role enum.UserRole, verified bool) (string, error) {
			return generator.Generate(intId, role, verified)
		}

		accessToken, err := h.service.RefreshUserToken(r.Context(), intId, tokenFunc)
		if err != nil {
			var errResp *rest.ErrorResponse
			switch {
			case errors.Is(err, service.ErrUserNotFound):
				h.Writer.WriteError(w, rest.NewNotFoundError(err))
			case errors.As(err, &errResp):
				h.Writer.WriteError(w, errResp)
			default:
				h.Writer.WriteError(w, rest.NewInternalServerErrorf("failed to refresh access token: %w", err))
			}
			return
		}

		setCookies(w, accessToken, accessTokenCookie)

		h.Logger.Infof("Rotate jwt - U-ID: %s", id)
		h.Writer.WriteNoContent(w, http.StatusOK)
	}
}

// VerifyUser
// @Summary Verify user
// @Description Verifies user using a jwt provided as a query parameter.
// @Tags auth
// @Accept json
// @Produce json
// @Param token query string true "jwt for user verification"
// @Success 200
// @Failure 400 {object} rest.ErrorResponse "Invalid id parameter"
// @Failure 401 {object} rest.ErrorResponse "Invalid token"
// @Failure 404 {object} rest.ErrorResponse "User was not found"
// @Failure 500 {object} rest.ErrorResponse "Internal server error"
// @Router /auth/verify [get]
func (h *Handler) VerifyUser(verifier verify_jwt.Verifier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenParam := r.URL.Query().Get("token")

		claims, errResp := verifier.Verify(tokenParam)
		if errResp != nil {
			h.Writer.WriteError(w, errResp)
			return
		}

		id, err := strconv.ParseInt(claims.ID, 10, 64)
		if err != nil {
			h.Writer.WriteError(w, rest.NewInternalServerErrorf("failed to verify user due to id parse - uid: %s, error: %w", claims.ID, err))
			return
		}

		if err := h.service.VerifyUser(r.Context(), tokenParam, time.Until(claims.ExpiresAt.Time), id); err != nil {
			switch {
			case errors.Is(err, service.ErrVerificationTokenUsed):
				h.Writer.WriteError(w, rest.NewConflictError(err))
			default:
				h.Writer.WriteError(w, rest.NewInternalServerErrorf("failed to verify user - uid: %d, error: %w", id, err))
			}
			return
		}

		go func() {
			if err := h.service.SendVerificationSuccessLetter(claims.Email); err != nil {
				h.Logger.Errorf("failed to send verification success letter - email: %s, uid: %d - error: %v", claims.Email, id, err)
				return
			}

			h.Logger.Infof("send verified letter - Email: %s, U-ID: %d", claims.Email, id)
		}()

		h.Logger.Infof("verify user - Email: %s, U-ID: %d", claims.Email, id)
		h.Writer.WriteNoContent(w, http.StatusOK)
	}
}

// sendVerificationLetter generates a verification JWT and sends a verification email.
// It uses the provided generator to create the JWT and the service to send the email.
// It logs errors and returns an error if any step fails.
func (h *Handler) sendVerificationLetter(ctx context.Context, generator verify_jwt.Generator, id int64, email string, name string) *rest.ErrorResponse {
	jwt, err := generator.Generate(id, email)
	if err != nil {
		h.Logger.Error(err)
		return err
	}

	if err := h.service.SendVerificationLetter(ctx, jwt, email, name); err != nil {
		errResp := rest.NewInternalServerErrorf("failed to send verification letter - uid: %d, error: %w", id, err)
		h.Logger.Error(errResp)
		return errResp
	}

	h.Logger.Infof("send verification letter - Email: %s, U-ID: %d", email, id)
	return nil
}
