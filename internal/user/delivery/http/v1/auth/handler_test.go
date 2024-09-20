package auth_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	mock_jwt "github.com/hexley21/fixup/internal/common/jwt/mock"
	"github.com/hexley21/fixup/internal/common/rest"
	"github.com/hexley21/fixup/internal/common/util/ctxutil"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/auth"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/user/enum"
	"github.com/hexley21/fixup/internal/user/service"
	mock_service "github.com/hexley21/fixup/internal/user/service/mock"
	"github.com/hexley21/fixup/internal/user/service/verifier"
	mock_verifier "github.com/hexley21/fixup/internal/user/service/verifier/mock"
	"github.com/hexley21/fixup/pkg/hasher"
	mock_validator "github.com/hexley21/fixup/pkg/validator/mock"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var (
	userDto = dto.User{
		ID:          "1",
		FirstName:   "Larry",
		LastName:    "Page",
		PhoneNumber: "995111222333",
		Email:       "larry@page.com",
		PictureUrl:  "larrypage.png",
		Role:        string(enum.UserRoleADMIN),
		UserStatus:  true,
		CreatedAt:   time.Now(),
	}

	userConfirmationDetailsDTO = dto.UserConfirmationDetails{
		ID:         "1",
		UserStatus: true,
		Firstname:  "Larry",
	}

	credentialsDto = dto.Credentials{
		ID:         "1",
		Role:       string(enum.UserRoleADMIN),
		UserStatus: true,
	}

	verifyClaims = verifier.VerifyClaims{
		ID:    "1",
		Email: "larry@page.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	registerCustomerJSON = `{"email": "larry@page.com", "phone_number": "995111222333", "first_name": "Larry", "last_name": "Page", "password": "larrypage123"}`
	registerProviderJSON = `{"email": "larry@page.com", "phone_number": "995111222333", "first_name": "Larry", "last_name": "Page", "password": "larrypage123", "personal_id_number": "1234567890"}`
	emailJSON            = `{"email": "larry@page.com"}`
	loginJSON            = `{"email": "larry@page.com", "password": "larry@page.com"}`

	token = "Ehx0DNg86zL"
)

func setup(t *testing.T) (*gomock.Controller, *mock_service.MockAuthService, *mock_validator.MockValidator, *mock_verifier.MockJwt, *mock_jwt.MockJwtGenerator, *mock_jwt.MockJwtGenerator, *auth.AuthHandler, *echo.Echo) {
	ctrl := gomock.NewController(t)
	mockAuthService := mock_service.NewMockAuthService(ctrl)
	mockValidator := mock_validator.NewMockValidator(ctrl)
	mockVerifierGenerator := mock_verifier.NewMockJwt(ctrl)
	mockAccessGenerator := mock_jwt.NewMockJwtGenerator(ctrl)
	mockRefreshGenerator := mock_jwt.NewMockJwtGenerator(ctrl)

	h := auth.NewAuthHandler(mockAuthService)

	e := echo.New()
	e.Validator = mockValidator

	return ctrl, mockAuthService, mockValidator, mockVerifierGenerator, mockAccessGenerator, mockRefreshGenerator, h, e
}

func TestRegisterCustomer_Success(t *testing.T) {
	ctx := context.Background()

	ctrl, mockAuthService, mockValidator, mockVerifierGenerator, _, _, h, e := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().RegisterCustomer(ctx, gomock.Any()).Return(userDto, nil)
	mockAuthService.EXPECT().SendConfirmationLetter(token, userDto.Email, userDto.FirstName).Return(nil)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockVerifierGenerator.EXPECT().GenerateJWT(userDto.ID, userDto.Email).Return(token, nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(registerCustomerJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	assert.NoError(t, h.RegisterCustomer(mockVerifierGenerator)(c))
	assert.Equal(t, http.StatusCreated, rec.Code)

	time.Sleep(time.Microsecond)
}

func TestRegisterCustomer_BindError(t *testing.T) {
	h := auth.NewAuthHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(registerCustomerJSON))
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.RegisterCustomer(nil)(c), &errResp) {
		assert.Equal(t, rest.MsgInvalidArguments, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, errResp.Status)
	}
}

func TestRegisterCustomer_InvalidArguments(t *testing.T) {
	ctrl, _, mockValidator, _, _, _, h, e := setup(t)
	defer ctrl.Finish()

	mockValidator.EXPECT().Validate(gomock.Any()).Return(errors.New(""))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(registerCustomerJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.RegisterCustomer(nil)(c), &errResp) {
		assert.Equal(t, rest.MsgInvalidArguments, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, errResp.Status)
	}
}

func TestRegisterCustomer_Conflict(t *testing.T) {
	ctx := context.Background()

	ctrl, mockAuthService, mockValidator, _, _, _, h, e := setup(t)
	defer ctrl.Finish()

	uniqueViolationErr := &pgconn.PgError{Code: pgerrcode.UniqueViolation}

	mockAuthService.EXPECT().RegisterCustomer(ctx, gomock.Any()).Return(dto.User{}, uniqueViolationErr)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(registerCustomerJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.RegisterCustomer(nil)(c), &errResp) {
		assert.ErrorIs(t, errResp.Cause, uniqueViolationErr)
		assert.Equal(t, rest.MsgUserAlreadyExists, errResp.Message)
		assert.Equal(t, http.StatusConflict, errResp.Status)
	}
}


func TestRegisterCustomer_ServiceError(t *testing.T) {
	ctx := context.Background()

	ctrl, mockAuthService, mockValidator, _, _, _, h, e := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().RegisterCustomer(ctx, gomock.Any()).Return(dto.User{}, errors.New(""))
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(registerCustomerJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.RegisterCustomer(nil)(c), &errResp) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}

func TestRegisterProvider_Success(t *testing.T) {
	ctx := context.Background()
	ctrl, mockAuthService, mockValidator, mockVerifierGenerator, _, _, h, e := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().RegisterProvider(ctx, gomock.Any()).Return(userDto, nil)
	mockAuthService.EXPECT().SendConfirmationLetter(token, userDto.Email, userDto.FirstName).Return(nil)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockVerifierGenerator.EXPECT().GenerateJWT(userDto.ID, userDto.Email).Return(token, nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(registerProviderJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	assert.NoError(t, h.RegisterProvider(mockVerifierGenerator)(c))
	assert.Equal(t, http.StatusCreated, rec.Code)

	time.Sleep(time.Microsecond)
}

func TestRegisterProvider_BindError(t *testing.T) {
	h := auth.NewAuthHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(registerProviderJSON))
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.RegisterProvider(nil)(c), &errResp) {
		assert.Equal(t, rest.MsgInvalidArguments, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, errResp.Status)
	}
}

func TestRegisterProvider_InvalidArguments(t *testing.T) {
	ctrl, _, mockValidator, _, _, _, h, e := setup(t)
	defer ctrl.Finish()

	mockValidator.EXPECT().Validate(gomock.Any()).Return(errors.New(""))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(registerProviderJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.RegisterProvider(nil)(c), &errResp) {
		assert.Equal(t, rest.MsgInvalidArguments, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, errResp.Status)
	}
}

func TestRegisterProvider_Conflict(t *testing.T) {
	ctx := context.Background()

	ctrl, mockAuthService, mockValidator, _, _, _, h, e := setup(t)
	defer ctrl.Finish()

	uniqueViolationErr := &pgconn.PgError{Code: pgerrcode.UniqueViolation}

	mockAuthService.EXPECT().RegisterProvider(ctx, gomock.Any()).Return(dto.User{}, uniqueViolationErr)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(registerProviderJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.RegisterProvider(nil)(c), &errResp) {
		assert.ErrorIs(t, errResp.Cause, uniqueViolationErr)
		assert.Equal(t, rest.MsgUserAlreadyExists, errResp.Message)
		assert.Equal(t, http.StatusConflict, errResp.Status)
	}
}

func TestRegisterProvider_ServiceError(t *testing.T) {
	ctx := context.Background()

	ctrl, mockAuthService, mockValidator, _, _, _, h, e := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().RegisterProvider(ctx, gomock.Any()).Return(dto.User{}, errors.New(""))
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(registerProviderJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.RegisterProvider(nil)(c), &errResp) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}

func TestResendConfirmationLetter_Success(t *testing.T) {
	ctx := context.Background()
	ctrl, mockAuthService, mockValidator, mockVerifierGenerator, _, _, h, e := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().GetUserConfirmationDetails(ctx, gomock.Any()).Return(userConfirmationDetailsDTO, nil)
	mockAuthService.EXPECT().SendConfirmationLetter(token, userDto.Email, userDto.FirstName).Return(nil)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockVerifierGenerator.EXPECT().GenerateJWT(userDto.ID, userDto.Email).Return(token, nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(emailJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	assert.NoError(t, h.ResendConfirmationLetter(mockVerifierGenerator)(c))
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestResendConfirmationLetter_BindError(t *testing.T) {
	h := auth.NewAuthHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(emailJSON))
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.ResendConfirmationLetter(nil)(c), &errResp) {
		assert.Equal(t, rest.MsgInvalidArguments, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, errResp.Status)
	}
}

func TestResendConfirmationLetter_InvalidArguments(t *testing.T) {
	ctrl, _, mockValidator, _, _, _, h, e := setup(t)
	defer ctrl.Finish()

	mockValidator.EXPECT().Validate(gomock.Any()).Return(errors.New(""))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(emailJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.ResendConfirmationLetter(nil)(c), &errResp) {
		assert.Equal(t, rest.MsgInvalidArguments, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, errResp.Status)
	}
}

func TestResendConfirmationLetter_Conflict(t *testing.T) {
	ctx := context.Background()

	ctrl, mockAuthService, mockValidator, _, _, _, h, e := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().GetUserConfirmationDetails(ctx, gomock.Any()).Return(dto.UserConfirmationDetails{}, service.ErrUserAlreadyActive)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(registerProviderJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.ResendConfirmationLetter(nil)(c), &errResp) {
		assert.ErrorIs(t, service.ErrUserAlreadyActive, errResp.Cause)
		assert.Equal(t, rest.MsgUserAlreadyExists, errResp.Message)
		assert.Equal(t, http.StatusConflict, errResp.Status)
	}
}

func TestResendConfirmationLetter_NotFound(t *testing.T) {
	ctx := context.Background()

	ctrl, mockAuthService, mockValidator, _, _, _, h, e := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().GetUserConfirmationDetails(ctx, gomock.Any()).Return(dto.UserConfirmationDetails{}, pgx.ErrNoRows)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(registerProviderJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.ResendConfirmationLetter(nil)(c), &errResp) {
		assert.ErrorIs(t, pgx.ErrNoRows, errResp.Cause)
		assert.Equal(t, rest.MsgUserNotFound, errResp.Message)
		assert.Equal(t, http.StatusNotFound, errResp.Status)
	}
}

func TestResendConfirmationLetter_ServiceError(t *testing.T) {
	ctx := context.Background()

	ctrl, mockAuthService, mockValidator, _, _, _, h, e := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().GetUserConfirmationDetails(ctx, gomock.Any()).Return(dto.UserConfirmationDetails{}, errors.New(""))
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(registerProviderJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.ResendConfirmationLetter(nil)(c), &errResp) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}

func TestResendConfirmationLetter_MailError(t *testing.T) {
	ctx := context.Background()
	ctrl, mockAuthService, mockValidator, mockVerifierGenerator, _, _, h, e := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().GetUserConfirmationDetails(ctx, gomock.Any()).Return(userConfirmationDetailsDTO, nil)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockVerifierGenerator.EXPECT().GenerateJWT(userDto.ID, userDto.Email).Return("", errors.New(""))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(emailJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.ResendConfirmationLetter(mockVerifierGenerator)(c), &errResp) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}

func TestLogin_Success(t *testing.T) {
	ctx := context.Background()

	ctrl, mockAuthService, mockValidator, _, mockAccessGenerator, mockRefreshGenerator, h, e := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().AuthenticateUser(ctx, gomock.Any()).Return(credentialsDto, nil)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockAccessGenerator.EXPECT().GenerateJWT(userDto.ID, userDto.Role, userDto.UserStatus).Return(token, nil)
	mockRefreshGenerator.EXPECT().GenerateJWT(userDto.ID, userDto.Role, userDto.UserStatus).Return(token, nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(loginJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	assert.NoError(t, h.Login(mockAccessGenerator, mockRefreshGenerator)(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Values("Set-Cookie")[0], fmt.Sprintf("access_token=%s; HttpOnly; Secure", token))
	assert.Contains(t, rec.Header().Values("Set-Cookie")[1], fmt.Sprintf("refresh_token=%s; HttpOnly; Secure", token))
}

func TestLogin_BindError(t *testing.T) {
	h := auth.NewAuthHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(loginJSON))
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.Login(nil, nil)(c), &errResp) {
		assert.Equal(t, rest.MsgInvalidArguments, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, errResp.Status)
	}
}

func TestLogin_InvalidArguments(t *testing.T) {
	ctrl, _, mockValidator, _, _, _, h, e := setup(t)
	defer ctrl.Finish()

	mockValidator.EXPECT().Validate(gomock.Any()).Return(errors.New(""))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(loginJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.Login(nil, nil)(c), &errResp) {
		assert.Equal(t, rest.MsgInvalidArguments, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, errResp.Status)
	}
}

func TestLogin_AuthError(t *testing.T) {
	ctx := context.Background()

	ctrl, mockAuthService, mockValidator, _, _, _, h, e := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().AuthenticateUser(ctx, gomock.Any()).Return(dto.Credentials{}, hasher.ErrPasswordMismatch)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(loginJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.Login(nil, nil)(c), &errResp) {
		assert.ErrorIs(t, hasher.ErrPasswordMismatch, errResp.Cause)
		assert.Equal(t, rest.MsgIncorrectEmailOrPass, errResp.Message)
		assert.Equal(t, http.StatusUnauthorized, errResp.Status)
	}
}

func TestLogin_ServiceError(t *testing.T) {
	ctx := context.Background()

	ctrl, mockAuthService, mockValidator, _, _, _, h, e := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().AuthenticateUser(ctx, gomock.Any()).Return(dto.Credentials{}, errors.New(""))
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(loginJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.Login(nil, nil)(c), &errResp) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}

func TestLogin_AccessTokenError(t *testing.T) {
	ctx := context.Background()

	ctrl, mockAuthService, mockValidator, _, mockAccessGenerator, _, h, e := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().AuthenticateUser(ctx, gomock.Any()).Return(credentialsDto, nil)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockAccessGenerator.EXPECT().GenerateJWT(userDto.ID, userDto.Role, userDto.UserStatus).Return("", rest.NewInternalServerError(errors.New("")))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(loginJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.Login(mockAccessGenerator, nil)(c), &errResp) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}

func TestLogin_RefreshTokenError(t *testing.T) {
	ctx := context.Background()

	ctrl, mockAuthService, mockValidator, _, mockAccessGenerator, mockRefreshGenerator, h, e := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().AuthenticateUser(ctx, gomock.Any()).Return(credentialsDto, nil)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockAccessGenerator.EXPECT().GenerateJWT(userDto.ID, userDto.Role, userDto.UserStatus).Return(token, nil)
	mockRefreshGenerator.EXPECT().GenerateJWT(userDto.ID, userDto.Role, userDto.UserStatus).Return("", rest.NewInternalServerError(errors.New("")))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(loginJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.Login(mockAccessGenerator, mockRefreshGenerator)(c), &errResp) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}

func TestLogout_Success(t *testing.T) {
	h := auth.NewAuthHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	assert.NoError(t, h.Logout(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Values("Set-Cookie")[0], "access_token=; Expires=Thu, 01 Jan 1970 00:00:00 GMT; Max-Age=0; HttpOnly")
	assert.Contains(t, rec.Header().Values("Set-Cookie")[1], "refresh_token=; Expires=Thu, 01 Jan 1970 00:00:00 GMT; Max-Age=0; HttpOnly")
}

func TestRefresh_Success(t *testing.T) {
	ctrl, _, _, _, mockAccessGenerator, _, h, e := setup(t)
	defer ctrl.Finish()

	mockAccessGenerator.EXPECT().GenerateJWT(userDto.ID, userDto.Role, userDto.UserStatus).Return(token, nil)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	ctxutil.SetJwtId(c, userDto.ID)
	ctxutil.SetJwtRole(c, enum.UserRole(userDto.Role))
	ctxutil.SetJwtUserStatus(c, userDto.UserStatus)

	assert.NoError(t, h.Refresh(mockAccessGenerator)(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Set-Cookie"), fmt.Sprintf("access_token=%s; HttpOnly; Secure", token))
}

func TestRefresh_JwtNotImplemented(t *testing.T) {
	t.Parallel()

	h := auth.NewAuthHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/", nil)

	t.Run("JWT id", func(t *testing.T) {
		rec := httptest.NewRecorder()
		e := echo.New()
		c := e.NewContext(req, rec)

		var errResp *rest.ErrorResponse
		if assert.ErrorAs(t, h.Refresh(nil)(c), &errResp) {
			assert.ErrorIs(t, ctxutil.ErrJwtNotImplemented.Cause, errResp.Cause)
			assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
			assert.Equal(t, http.StatusInternalServerError, errResp.Status)
		}
		assert.NotContains(t, rec.Header().Get("Set-Cookie"), fmt.Sprintf("access_token=%s; HttpOnly; Secure", token))
	})

	t.Run("JWT role", func(t *testing.T) {
		rec := httptest.NewRecorder()
		e := echo.New()
		c := e.NewContext(req, rec)
		ctxutil.SetJwtId(c, userDto.ID)

		var errResp *rest.ErrorResponse
		if assert.ErrorAs(t, h.Refresh(nil)(c), &errResp) {
			assert.ErrorIs(t, ctxutil.ErrJwtNotImplemented.Cause, errResp.Cause)
			assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
			assert.Equal(t, http.StatusInternalServerError, errResp.Status)
		}
		assert.NotContains(t, rec.Header().Get("Set-Cookie"), fmt.Sprintf("access_token=%s; HttpOnly; Secure", token))
	})

	t.Run("JWT user status", func(t *testing.T) {
		rec := httptest.NewRecorder()
		e := echo.New()
		c := e.NewContext(req, rec)
		ctxutil.SetJwtId(c, userDto.ID)
		ctxutil.SetJwtRole(c, enum.UserRole(userDto.Role))

		var errResp *rest.ErrorResponse
		if assert.ErrorAs(t, h.Refresh(nil)(c), &errResp) {
			assert.ErrorIs(t, ctxutil.ErrJwtNotImplemented.Cause, errResp.Cause)
			assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
			assert.Equal(t, http.StatusInternalServerError, errResp.Status)
		}
		assert.NotContains(t, rec.Header().Get("Set-Cookie"), fmt.Sprintf("access_token=%s; HttpOnly; Secure", token))
	})
}

func TestVerifyEmail_Success(t *testing.T) {
	ctx := context.Background()

	ctrl, mockAuthService, _, mockVerifyJWT, _, _, h, e := setup(t)
	defer ctrl.Finish()

	mockVerifyJWT.EXPECT().VerifyJWT(gomock.Any()).Return(verifyClaims, nil)
	mockAuthService.EXPECT().VerifyUser(ctx, int64(1), verifyClaims.Email).Return(nil)
	mockAuthService.EXPECT().SendVerifiedLetter(verifyClaims.Email).Return(nil)

	q := make(url.Values)
	q.Set("token", token)
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	assert.NoError(t, h.VerifyEmail(mockVerifyJWT)(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	time.Sleep(time.Microsecond)
}

func TestVerifyEmail_InvalidToken(t *testing.T) {
	ctrl, _, _, mockVerifyJWT, _, _, h, e := setup(t)
	defer ctrl.Finish()

	mockVerifyJWT.EXPECT().VerifyJWT(token).Return(verifier.VerifyClaims{}, errors.New(rest.MsgInvalidToken))

	q := make(url.Values)
	q.Set("token", token)
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.VerifyEmail(mockVerifyJWT)(c), &errResp) {
		assert.Equal(t, rest.MsgInvalidToken, errResp.Message)
		assert.Equal(t, http.StatusUnauthorized, errResp.Status)
	}
}

func TestVerifyEmail_ParseIDError(t *testing.T) {
	ctrl, _, _, mockVerifyJWT, _, _, h, e := setup(t)
	defer ctrl.Finish()

	mockVerifyJWT.EXPECT().VerifyJWT(token).Return(verifier.VerifyClaims{ID: "invalid-id"}, nil)

	q := make(url.Values)
	q.Set("token", token)
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.VerifyEmail(mockVerifyJWT)(c), &errResp) {
		assert.ErrorIs(t, errResp.Cause, strconv.ErrSyntax)
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}

func TestVerifyEmail_NotFound(t *testing.T) {
	ctx := context.Background()

    ctrl, mockAuthService, _, mockVerifyJWT, _, _, h, e := setup(t)
    defer ctrl.Finish()

    mockVerifyJWT.EXPECT().VerifyJWT(token).Return(verifyClaims, nil)
    mockAuthService.EXPECT().VerifyUser(ctx, int64(1), verifyClaims.Email).Return(pgx.ErrNoRows)

	q := make(url.Values)
	q.Set("token", token)
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)

    var errResp *rest.ErrorResponse
    if assert.ErrorAs(t, h.VerifyEmail(mockVerifyJWT)(c), &errResp) {
		assert.ErrorIs(t, errResp.Cause, pgx.ErrNoRows)
        assert.Equal(t, rest.MsgUserNotFound, errResp.Message)
        assert.Equal(t, http.StatusNotFound, errResp.Status)
    }
}

func TestVerifyEmail_ServiceError(t *testing.T) {
	ctx := context.Background()

    ctrl, mockAuthService, _, mockVerifyJWT, _, _, h, e := setup(t)
    defer ctrl.Finish()

    mockVerifyJWT.EXPECT().VerifyJWT(token).Return(verifyClaims, nil)
    mockAuthService.EXPECT().VerifyUser(ctx, int64(1), verifyClaims.Email).Return(errors.New(""))

	q := make(url.Values)
	q.Set("token", token)
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)

    var errResp *rest.ErrorResponse
    if assert.ErrorAs(t, h.VerifyEmail(mockVerifyJWT)(c), &errResp) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
        assert.Equal(t, http.StatusInternalServerError, errResp.Status)
    }
}

func TestVerifyEmail_MailError(t *testing.T) {
	ctx := context.Background()

	ctrl, mockAuthService, _, mockVerifyJWT, _, _, h, e := setup(t)
	defer ctrl.Finish()

	mockVerifyJWT.EXPECT().VerifyJWT(gomock.Any()).Return(verifyClaims, nil)
	mockAuthService.EXPECT().VerifyUser(ctx, int64(1), verifyClaims.Email).Return(nil)
	mockAuthService.EXPECT().SendVerifiedLetter(verifyClaims.Email).Return(errors.New(""))

	q := make(url.Values)
	q.Set("token", token)
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)

    assert.NoError(t, h.VerifyEmail(mockVerifyJWT)(c))
    assert.Equal(t, http.StatusOK, rec.Code)
	
	time.Sleep(time.Microsecond)
}
