package auth_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hexley21/fixup/internal/common/app_error"
	mock_jwt "github.com/hexley21/fixup/internal/common/auth_jwt/mock"
	"github.com/hexley21/fixup/internal/common/util/ctx_util"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/auth"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/user/enum"
	mock_service "github.com/hexley21/fixup/internal/user/service/mock"
	"github.com/hexley21/fixup/internal/user/service/verifier"
	mock_verifier "github.com/hexley21/fixup/internal/user/service/verifier/mock"
	"github.com/hexley21/fixup/pkg/hasher"
	"github.com/hexley21/fixup/pkg/http/binder/std_binder"
	"github.com/hexley21/fixup/pkg/http/json/std_json"
	"github.com/hexley21/fixup/pkg/http/rest"
	"github.com/hexley21/fixup/pkg/http/writer/json_writer"
	"github.com/hexley21/fixup/pkg/infra/postgres/pg_error"
	"github.com/hexley21/fixup/pkg/logger/std_logger"
	mock_validator "github.com/hexley21/fixup/pkg/validator/mock"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redis/go-redis/v9"
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
		UserStatus: false,
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

func setup(t *testing.T) (*gomock.Controller, *mock_service.MockAuthService, *mock_validator.MockValidator, *mock_verifier.MockJWTManager, *mock_jwt.MockJWTGenerator, *mock_jwt.MockJWTGenerator, *auth.HandlerFactory) {
	ctrl := gomock.NewController(t)
	mockAuthService := mock_service.NewMockAuthService(ctrl)
	mockValidator := mock_validator.NewMockValidator(ctrl)
	mockVerifierGenerator := mock_verifier.NewMockJWTManager(ctrl)
	mockAccessGenerator := mock_jwt.NewMockJWTGenerator(ctrl)
	mockRefreshGenerator := mock_jwt.NewMockJWTGenerator(ctrl)

	logger := std_logger.New()
	jsonManager := std_json.New()

	f := auth.NewFactory(
		logger,
		std_binder.New(jsonManager),
		mockValidator,
		json_writer.New(logger, jsonManager),
		mockAuthService,
	)

	return ctrl, mockAuthService, mockValidator, mockVerifierGenerator, mockAccessGenerator, mockRefreshGenerator, f
}

func TestRegisterCustomer_Success(t *testing.T) {
	ctx := context.Background()

	ctrl, mockAuthService, mockValidator, mockVerifierGenerator, _, _, f := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().RegisterCustomer(ctx, gomock.Any()).Return(userDto, nil)
	mockAuthService.EXPECT().SendConfirmationLetter(ctx, token, userDto.Email, userDto.FirstName).Return(nil)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockVerifierGenerator.EXPECT().GenerateJWT(userDto.ID, userDto.Email).Return(token, nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(registerCustomerJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	f.RegisterCustomer(mockVerifierGenerator).ServeHTTP(rec, req)

	assert.Empty(t, rec.Body.String())
	assert.Equal(t, http.StatusCreated, rec.Code)

	time.Sleep(time.Microsecond)
}

func TestRegisterCustomer_BindError(t *testing.T) {
	ctrl, _, _, _, _, _, f := setup(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(registerCustomerJSON))
	rec := httptest.NewRecorder()

	f.RegisterCustomer(nil).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgUnsupportedMedia, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, errResp.Status)
	}
}

func TestRegisterCustomer_InvalidArguments(t *testing.T) {
	ctrl, _, mockValidator, _, _, _, f := setup(t)
	defer ctrl.Finish()

	mockValidator.EXPECT().Validate(gomock.Any()).Return(rest.NewInvalidArgumentsError(nil))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(registerCustomerJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	f.RegisterCustomer(nil).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInvalidArguments, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, errResp.Status)
	}
}

func TestRegisterCustomer_Conflict(t *testing.T) {
	ctx := context.Background()

	ctrl, mockAuthService, mockValidator, _, _, _, f := setup(t)
	defer ctrl.Finish()

	uniqueViolationErr := &pgconn.PgError{Code: pgerrcode.UniqueViolation}

	mockAuthService.EXPECT().RegisterCustomer(ctx, gomock.Any()).Return(dto.User{}, uniqueViolationErr)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(registerCustomerJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	f.RegisterCustomer(nil).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, auth.MsgUserAlreadyExists, errResp.Message)
		assert.Equal(t, http.StatusConflict, errResp.Status)
	}
}

func TestRegisterCustomer_ServiceError(t *testing.T) {
	ctx := context.Background()

	ctrl, mockAuthService, mockValidator, _, _, _, f := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().RegisterCustomer(ctx, gomock.Any()).Return(dto.User{}, errors.New(""))
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(registerCustomerJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	f.RegisterCustomer(nil).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}

func TestRegisterProvider_Success(t *testing.T) {
	ctx := context.Background()
	ctrl, mockAuthService, mockValidator, mockVerifierGenerator, _, _, f := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().RegisterProvider(ctx, gomock.Any()).Return(userDto, nil)
	mockAuthService.EXPECT().SendConfirmationLetter(ctx, token, userDto.Email, userDto.FirstName).Return(nil)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockVerifierGenerator.EXPECT().GenerateJWT(userDto.ID, userDto.Email).Return(token, nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(registerProviderJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	f.RegisterProvider(mockVerifierGenerator).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	time.Sleep(time.Microsecond)
}

func TestRegisterProvider_BindError(t *testing.T) {
	ctrl, _, _, _, _, _, f := setup(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(registerProviderJSON))
	rec := httptest.NewRecorder()

	f.RegisterProvider(nil).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgUnsupportedMedia, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, errResp.Status)
	}
}

func TestRegisterProvider_InvalidArguments(t *testing.T) {
	ctrl, _, mockValidator, _, _, _, f := setup(t)
	defer ctrl.Finish()

	mockValidator.EXPECT().Validate(gomock.Any()).Return(rest.NewInvalidArgumentsError(errors.New("")))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(registerProviderJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	f.RegisterProvider(nil).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInvalidArguments, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, errResp.Status)
	}
}

func TestRegisterProvider_Conflict(t *testing.T) {
	ctx := context.Background()

	ctrl, mockAuthService, mockValidator, _, _, _, f := setup(t)
	defer ctrl.Finish()

	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockAuthService.EXPECT().RegisterProvider(ctx, gomock.Any()).Return(dto.User{}, &pgconn.PgError{Code: pgerrcode.UniqueViolation})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(registerProviderJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	f.RegisterProvider(nil).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, auth.MsgUserAlreadyExists, errResp.Message)
		assert.Equal(t, http.StatusConflict, errResp.Status)
	}
}

func TestRegisterProvider_ServiceError(t *testing.T) {
	ctx := context.Background()

	ctrl, mockAuthService, mockValidator, _, _, _, f := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().RegisterProvider(ctx, gomock.Any()).Return(dto.User{}, errors.New(""))
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(registerProviderJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	f.RegisterProvider(nil).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}

func TestResendConfirmationLetter_Success(t *testing.T) {
	ctx := context.Background()

	ctrl, mockAuthService, mockValidator, mockVerifierGenerator, _, _, f := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().GetUserConfirmationDetails(ctx, gomock.Any()).Return(userConfirmationDetailsDTO, nil)
	mockAuthService.EXPECT().SendConfirmationLetter(ctx, token, userDto.Email, userDto.FirstName).Return(nil)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockVerifierGenerator.EXPECT().GenerateJWT(userDto.ID, userDto.Email).Return(token, nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(emailJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	f.ResendConfirmationLetter(mockVerifierGenerator).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestResendConfirmationLetter_BindError(t *testing.T) {
	ctrl, _, _, _, _, _, f := setup(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(emailJSON))
	rec := httptest.NewRecorder()

	f.ResendConfirmationLetter(nil).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgUnsupportedMedia, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, errResp.Status)
	}
}

func TestResendConfirmationLetter_InvalidArguments(t *testing.T) {
	ctrl, _, mockValidator, _, _, _, f := setup(t)
	defer ctrl.Finish()

	mockValidator.EXPECT().Validate(gomock.Any()).Return(rest.NewInvalidArgumentsError(errors.New("")))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(emailJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	f.ResendConfirmationLetter(nil).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInvalidArguments, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, errResp.Status)
	}
}

func TestResendConfirmationLetter_Conflict(t *testing.T) {
	ctx := context.Background()

	ctrl, mockAuthService, mockValidator, _, _, _, f := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().GetUserConfirmationDetails(ctx, gomock.Any()).Return(dto.UserConfirmationDetails{UserStatus: true}, nil)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(emailJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	f.ResendConfirmationLetter(nil).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, auth.MsgUserAlreadyActivated, errResp.Message)
		assert.Equal(t, http.StatusConflict, errResp.Status)
	}
}

func TestResendConfirmationLetter_NotFound(t *testing.T) {
	ctx := context.Background()

	ctrl, mockAuthService, mockValidator, mockVerifierGenerator, _, _, f := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().GetUserConfirmationDetails(ctx, gomock.Any()).Return(dto.UserConfirmationDetails{}, pg_error.ErrNotFound)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(emailJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	f.ResendConfirmationLetter(mockVerifierGenerator).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, app_error.MsgUserNotFound, errResp.Message)
		assert.Equal(t, http.StatusNotFound, errResp.Status)
	}
}

func TestResendConfirmationLetter_Already(t *testing.T) {
	ctx := context.Background()

	ctrl, mockAuthService, mockValidator, mockVerifierGenerator, _, _, f := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().GetUserConfirmationDetails(ctx, gomock.Any()).Return(dto.UserConfirmationDetails{}, pg_error.ErrNotFound)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(emailJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	f.ResendConfirmationLetter(mockVerifierGenerator).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, app_error.MsgUserNotFound, errResp.Message)
		assert.Equal(t, http.StatusNotFound, errResp.Status)
	}
}

func TestResendConfirmationLetter_ServiceError(t *testing.T) {
	ctx := context.Background()

	ctrl, mockAuthService, mockValidator, _, _, _, f := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().GetUserConfirmationDetails(ctx, gomock.Any()).Return(dto.UserConfirmationDetails{}, errors.New(""))
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(registerProviderJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	f.ResendConfirmationLetter(nil).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}

func TestResendConfirmationLetter_MailError(t *testing.T) {
	ctx := context.Background()
	ctrl, mockAuthService, mockValidator, mockVerifierGenerator, _, _, f := setup(t)
	defer ctrl.Finish()

	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockAuthService.EXPECT().GetUserConfirmationDetails(ctx, gomock.Any()).Return(userConfirmationDetailsDTO, nil)
	mockVerifierGenerator.EXPECT().GenerateJWT(userDto.ID, userDto.Email).Return("", rest.NewUnauthorizedError(errors.New(""), app_error.MsgInvalidToken))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(emailJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	f.ResendConfirmationLetter(mockVerifierGenerator).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, app_error.MsgInvalidToken, errResp.Message)
		assert.Equal(t, http.StatusUnauthorized, errResp.Status)
	}
}

func TestLogin_Success(t *testing.T) {
	ctx := context.Background()

	ctrl, mockAuthService, mockValidator, _, mockAccessGenerator, mockRefreshGenerator, f := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().AuthenticateUser(ctx, gomock.Any()).Return(credentialsDto, nil)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockAccessGenerator.EXPECT().GenerateJWT(userDto.ID, userDto.Role, userDto.UserStatus).Return(token, nil)
	mockRefreshGenerator.EXPECT().GenerateJWT(userDto.ID, userDto.Role, userDto.UserStatus).Return(token, nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(loginJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	f.Login(mockAccessGenerator, mockRefreshGenerator).ServeHTTP(rec, req)

	cookies := rec.Header().Values("Set-Cookie")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, cookies[0], fmt.Sprintf("access_token=%s; HttpOnly; Secure", token))
	assert.Contains(t, cookies[1], fmt.Sprintf("refresh_token=%s; HttpOnly; Secure", token))
}

func TestLogin_BindError(t *testing.T) {
	ctrl, _, _, _, _, _, f := setup(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(loginJSON))
	rec := httptest.NewRecorder()

	f.Login(nil, nil).ServeHTTP(rec, req)

    var errResp rest.ErrorResponse
    if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
        assert.Equal(t, rest.MsgUnsupportedMedia, errResp.Message)
        assert.Equal(t, http.StatusBadRequest, errResp.Status)
    }
}

func TestLogin_InvalidArguments(t *testing.T) {
	ctrl, _, mockValidator, _, _, _, f := setup(t)
	defer ctrl.Finish()

	mockValidator.EXPECT().Validate(gomock.Any()).Return(rest.NewInvalidArgumentsError(errors.New("")))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(loginJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	
	f.Login(nil, nil).ServeHTTP(rec, req)

    var errResp rest.ErrorResponse
    if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInvalidArguments, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, errResp.Status)
	}
}

func TestLogin_AuthError(t *testing.T) {
	ctx := context.Background()

	ctrl, mockAuthService, mockValidator, _, _, _, f := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().AuthenticateUser(ctx, gomock.Any()).Return(dto.Credentials{}, hasher.ErrPasswordMismatch)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(loginJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	
	f.Login(nil, nil).ServeHTTP(rec, req)

    var errResp rest.ErrorResponse
    if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, auth.MsgIncorrectEmailOrPass, errResp.Message)
		assert.Equal(t, http.StatusUnauthorized, errResp.Status)
	}
}

func TestLogin_ServiceError(t *testing.T) {
	ctx := context.Background()

	ctrl, mockAuthService, mockValidator, _, _, _, f := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().AuthenticateUser(ctx, gomock.Any()).Return(dto.Credentials{}, errors.New(""))
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(loginJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	
	f.Login(nil, nil).ServeHTTP(rec, req)

    var errResp rest.ErrorResponse
    if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}

func TestLogin_AccessTokenError(t *testing.T) {
	ctx := context.Background()

	ctrl, mockAuthService, mockValidator, _, mockAccessGenerator, _, f := setup(t)
	defer ctrl.Finish()

	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockAuthService.EXPECT().AuthenticateUser(ctx, gomock.Any()).Return(credentialsDto, nil)
	mockAccessGenerator.EXPECT().GenerateJWT(userDto.ID, userDto.Role, userDto.UserStatus).Return("", rest.NewInternalServerError(errors.New("")))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(loginJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	
	f.Login(mockAccessGenerator, nil).ServeHTTP(rec, req)

    var errResp rest.ErrorResponse
    if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}

func TestLogin_RefreshTokenError(t *testing.T) {
	ctx := context.Background()

	ctrl, mockAuthService, mockValidator, _, mockAccessGenerator, mockRefreshGenerator, f := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().AuthenticateUser(ctx, gomock.Any()).Return(credentialsDto, nil)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockAccessGenerator.EXPECT().GenerateJWT(userDto.ID, userDto.Role, userDto.UserStatus).Return(token, nil)
	mockRefreshGenerator.EXPECT().GenerateJWT(userDto.ID, userDto.Role, userDto.UserStatus).Return("", rest.NewInternalServerError(errors.New("")))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(loginJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	
	f.Login(mockAccessGenerator, mockRefreshGenerator).ServeHTTP(rec, req)

    var errResp rest.ErrorResponse
    if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}

func TestLogout_Success(t *testing.T) {
	ctrl, _, _, _, _, _, f := setup(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	f.Logout(rec, req)

	cookies := rec.Header().Values("Set-Cookie")

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, cookies[0], "access_token=; Expires=Thu, 01 Jan 1970 00:00:00 GMT; Max-Age=0; HttpOnly")
	assert.Contains(t, cookies[1], "refresh_token=; Expires=Thu, 01 Jan 1970 00:00:00 GMT; Max-Age=0; HttpOnly")
}

func TestRefresh_Success(t *testing.T) {
	ctrl, _, _, _, mockAccessGenerator, _, f := setup(t)
	defer ctrl.Finish()

	mockAccessGenerator.EXPECT().GenerateJWT(userDto.ID, userDto.Role, userDto.UserStatus).Return(token, nil)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	ctx := ctx_util.SetJWTId(req.Context(), userDto.ID)
	ctx = ctx_util.SetJWTRole(ctx, enum.UserRole(userDto.Role))
	ctx = ctx_util.SetJWTUserStatus(ctx, userDto.UserStatus)
	
	f.Refresh(mockAccessGenerator).ServeHTTP(rec, req.WithContext(ctx))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Set-Cookie"), fmt.Sprintf("access_token=%s; HttpOnly; Secure", token))
}

func TestRefresh_JwtNotSet(t *testing.T) {
	t.Parallel()

	ctrl, _, _, _, _, _, f := setup(t)
	defer ctrl.Finish()

	handler := f.Refresh(nil)

	t.Run("JWT id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		var errResp rest.ErrorResponse
		if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
			assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
			assert.Equal(t, http.StatusInternalServerError, errResp.Status)
		}
		assert.NotContains(t, rec.Header().Get("Set-Cookie"), fmt.Sprintf("access_token=%s; HttpOnly; Secure", token))
	})

	t.Run("JWT role", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req.WithContext(ctx_util.SetJWTId(req.Context(), userDto.ID)))

		var errResp rest.ErrorResponse
		if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
			assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
			assert.Equal(t, http.StatusInternalServerError, errResp.Status)
		}
		assert.NotContains(t, rec.Header().Get("Set-Cookie"), fmt.Sprintf("access_token=%s; HttpOnly; Secure", token))
	})

	t.Run("JWT user status", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		
		ctx := ctx_util.SetJWTId(req.Context(), userDto.ID)
		ctx = ctx_util.SetJWTRole(ctx, enum.UserRole(userDto.Role))

		handler.ServeHTTP(rec, req.WithContext(ctx))

		var errResp rest.ErrorResponse
		if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
			assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
			assert.Equal(t, http.StatusInternalServerError, errResp.Status)
		}
		assert.NotContains(t, rec.Header().Get("Set-Cookie"), fmt.Sprintf("access_token=%s; HttpOnly; Secure", token))
	})
}

func TestVerifyEmail_Success(t *testing.T) {
	ctx := context.Background()

	ctrl, mockAuthService, _, mockVerifyJWT, _, _, f := setup(t)
	defer ctrl.Finish()

	mockVerifyJWT.EXPECT().VerifyJWT(gomock.Any()).Return(verifyClaims, nil)
	mockAuthService.EXPECT().VerifyUser(ctx, gomock.Any(), gomock.Any(), int64(1), verifyClaims.Email).Return(nil)
	mockAuthService.EXPECT().SendVerifiedLetter(verifyClaims.Email).Return(nil)

	q := make(url.Values)
	q.Set("token", token)

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	f.VerifyEmail(mockVerifyJWT).ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	time.Sleep(time.Microsecond)
}

func TestVerifyEmail_InvalidToken(t *testing.T) {
	ctrl, _, _, mockVerifyJWT, _, _, f := setup(t)
	defer ctrl.Finish()

	mockVerifyJWT.EXPECT().VerifyJWT(token).Return(verifier.VerifyClaims{}, rest.NewUnauthorizedError(errors.New(""), app_error.MsgInvalidToken))

	q := make(url.Values)
	q.Set("token", token)
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	f.VerifyEmail(mockVerifyJWT).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, app_error.MsgInvalidToken, errResp.Message)
		assert.Equal(t, http.StatusUnauthorized, errResp.Status)
	}
}

func TestVerifyEmail_ParseIDError(t *testing.T) {
	ctrl, _, _, mockVerifyJWT, _, _, f := setup(t)
	defer ctrl.Finish()

	mockVerifyJWT.EXPECT().VerifyJWT(token).Return(verifier.VerifyClaims{ID: "invalid-id"}, nil)

	q := make(url.Values)
	q.Set("token", token)
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	
	f.VerifyEmail(mockVerifyJWT).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}

func TestVerifyEmail_AlreadyActivated(t *testing.T) {
	ctx := context.Background()

	ctrl, mockAuthService, _, mockVerifyJWT, _, _, f := setup(t)
	defer ctrl.Finish()

	mockVerifyJWT.EXPECT().VerifyJWT(token).Return(verifyClaims, nil)
	mockAuthService.EXPECT().VerifyUser(ctx, gomock.Any(), gomock.Any(), int64(1), verifyClaims.Email).Return(redis.TxFailedErr)

	q := make(url.Values)
	q.Set("token", token)
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	
	f.VerifyEmail(mockVerifyJWT).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, auth.MsgTokenAlreadyUsed, errResp.Message)
		assert.Equal(t, http.StatusConflict, errResp.Status)
	}
}

func TestVerifyEmail_NotFound(t *testing.T) {
	ctx := context.Background()

	ctrl, mockAuthService, _, mockVerifyJWT, _, _, f := setup(t)
	defer ctrl.Finish()

	mockVerifyJWT.EXPECT().VerifyJWT(token).Return(verifyClaims, nil)
	mockAuthService.EXPECT().VerifyUser(ctx, gomock.Any(), gomock.Any(), int64(1), verifyClaims.Email).Return(pg_error.ErrNotFound)

	q := make(url.Values)
	q.Set("token", token)
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	
	f.VerifyEmail(mockVerifyJWT).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, app_error.MsgUserNotFound, errResp.Message)
		assert.Equal(t, http.StatusNotFound, errResp.Status)
	}
}

func TestVerifyEmail_ServiceError(t *testing.T) {
	ctx := context.Background()

	ctrl, mockAuthService, _, mockVerifyJWT, _, _, f := setup(t)
	defer ctrl.Finish()

	mockVerifyJWT.EXPECT().VerifyJWT(token).Return(verifyClaims, nil)
	mockAuthService.EXPECT().VerifyUser(ctx, gomock.Any(), gomock.Any(), int64(1), verifyClaims.Email).Return(errors.New(""))

	q := make(url.Values)
	q.Set("token", token)
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	
	f.VerifyEmail(mockVerifyJWT).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}

func TestVerifyEmail_MailError(t *testing.T) {
	ctx := context.Background()

	ctrl, mockAuthService, _, mockVerifyJWT, _, _, f := setup(t)
	defer ctrl.Finish()

	mockVerifyJWT.EXPECT().VerifyJWT(gomock.Any()).Return(verifyClaims, nil)
	mockAuthService.EXPECT().VerifyUser(ctx, gomock.Any(), gomock.Any(), int64(1), verifyClaims.Email).Return(nil)
	mockAuthService.EXPECT().SendVerifiedLetter(verifyClaims.Email).Return(errors.New(""))

	q := make(url.Values)
	q.Set("token", token)
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	f.VerifyEmail(mockVerifyJWT).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	time.Sleep(time.Microsecond)
}
