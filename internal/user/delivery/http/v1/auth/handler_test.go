package auth_test

import (
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
	mock_auth_jwt "github.com/hexley21/fixup/internal/common/auth_jwt/mock"
	"github.com/hexley21/fixup/internal/common/enum"
	"github.com/hexley21/fixup/internal/common/util/ctx_util"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/auth"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/dto"
	mock_refresh_jwt "github.com/hexley21/fixup/internal/user/jwt/refresh_jwt/mock"
	"github.com/hexley21/fixup/internal/user/jwt/verify_jwt"
	mock_verify_jwt "github.com/hexley21/fixup/internal/user/jwt/verify_jwt/mock"
	mock_service "github.com/hexley21/fixup/internal/user/service/mock"
	"github.com/hexley21/fixup/pkg/hasher"
	"github.com/hexley21/fixup/pkg/http/binder/std_binder"
	"github.com/hexley21/fixup/pkg/http/handler"
	"github.com/hexley21/fixup/pkg/http/json/std_json"
	"github.com/hexley21/fixup/pkg/http/rest"
	"github.com/hexley21/fixup/pkg/http/writer/json_writer"
	"github.com/hexley21/fixup/pkg/infra/postgres/pg_error"
	"github.com/hexley21/fixup/pkg/logger/std_logger"
	mock_validator "github.com/hexley21/fixup/pkg/validator/mock"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
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

	UserIdentity = dto.UserIdentity{
		ID:         "1",
		Role:       string(enum.UserRoleADMIN),
		UserStatus: true,
	}

	UserRoleAndStatus = dto.UserRoleAndStatus{
		Role:       string(enum.UserRoleADMIN),
		UserStatus: true,
	}

	verifyClaims = verify_jwt.VerifyClaims{
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

func setup(t *testing.T) (
	ctrl *gomock.Controller,
	mockAuthService *mock_service.MockAuthService,
	mockValidator *mock_validator.MockValidator,
	mockverify_jwtGenerator *mock_verify_jwt.MockJWTManager,
	mockAccessGenerator *mock_auth_jwt.MockJWTGenerator,
	mockRefreshGenerator *mock_refresh_jwt.MockJWTGenerator,
	h *auth.Handler,
) {
	ctrl = gomock.NewController(t)
	mockAuthService = mock_service.NewMockAuthService(ctrl)
	mockValidator = mock_validator.NewMockValidator(ctrl)
	mockverify_jwtGenerator = mock_verify_jwt.NewMockJWTManager(ctrl)
	mockAccessGenerator = mock_auth_jwt.NewMockJWTGenerator(ctrl)
	mockRefreshGenerator = mock_refresh_jwt.NewMockJWTGenerator(ctrl)

	logger := std_logger.New()
	jsonManager := std_json.New()

	h = auth.NewFactory(
		handler.NewComponents(logger, std_binder.New(jsonManager), mockValidator, json_writer.New(logger, jsonManager)),
		mockAuthService,
	)

	return
}

func TestRegisterCustomer_Success(t *testing.T) {
	ctrl, mockAuthService, mockValidator, mockverify_jwtGenerator, _, _, h := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().RegisterCustomer(gomock.Any(), gomock.Any()).Return(userDto, nil)
	mockAuthService.EXPECT().SendConfirmationLetter(gomock.Any(), token, userDto.Email, userDto.FirstName).Return(nil)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockverify_jwtGenerator.EXPECT().GenerateJWT(userDto.ID, userDto.Email).Return(token, nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(registerCustomerJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.RegisterCustomer(mockverify_jwtGenerator).ServeHTTP(rec, req)

	assert.Empty(t, rec.Body.String())
	assert.Equal(t, http.StatusCreated, rec.Code)

	time.Sleep(500 * time.Millisecond)
}

func TestRegisterCustomer_BindError(t *testing.T) {
	ctrl, _, _, _, _, _, h := setup(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(registerCustomerJSON))
	rec := httptest.NewRecorder()

	h.RegisterCustomer(nil).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgUnsupportedMedia, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}

func TestRegisterCustomer_InvalidArguments(t *testing.T) {
	ctrl, _, mockValidator, _, _, _, h := setup(t)
	defer ctrl.Finish()

	mockValidator.EXPECT().Validate(gomock.Any()).Return(rest.NewInvalidArgumentsError(nil))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(registerCustomerJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.RegisterCustomer(nil).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInvalidArguments, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}

func TestRegisterCustomer_Conflict(t *testing.T) {
	ctrl, mockAuthService, mockValidator, _, _, _, h := setup(t)
	defer ctrl.Finish()

	uniqueViolationErr := &pgconn.PgError{Code: pgerrcode.UniqueViolation}

	mockAuthService.EXPECT().RegisterCustomer(gomock.Any(), gomock.Any()).Return(dto.User{}, uniqueViolationErr)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(registerCustomerJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.RegisterCustomer(nil).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, auth.MsgUserAlreadyExists, errResp.Message)
		assert.Equal(t, http.StatusConflict, rec.Code)
	}
}

func TestRegisterCustomer_ServiceError(t *testing.T) {
	ctrl, mockAuthService, mockValidator, _, _, _, h := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().RegisterCustomer(gomock.Any(), gomock.Any()).Return(dto.User{}, errors.New(""))
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(registerCustomerJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.RegisterCustomer(nil).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	}
}

func TestRegisterProvider_Success(t *testing.T) {
	ctrl, mockAuthService, mockValidator, mockverify_jwtGenerator, _, _, h := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().RegisterProvider(gomock.Any(), gomock.Any()).Return(userDto, nil)
	mockAuthService.EXPECT().SendConfirmationLetter(gomock.Any(), token, userDto.Email, userDto.FirstName).Return(nil)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockverify_jwtGenerator.EXPECT().GenerateJWT(userDto.ID, userDto.Email).Return(token, nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(registerProviderJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.RegisterProvider(mockverify_jwtGenerator).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	time.Sleep(500 * time.Millisecond)
}

func TestRegisterProvider_BindError(t *testing.T) {
	ctrl, _, _, _, _, _, h := setup(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(registerProviderJSON))
	rec := httptest.NewRecorder()

	h.RegisterProvider(nil).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgUnsupportedMedia, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}

func TestRegisterProvider_InvalidArguments(t *testing.T) {
	ctrl, _, mockValidator, _, _, _, h := setup(t)
	defer ctrl.Finish()

	mockValidator.EXPECT().Validate(gomock.Any()).Return(rest.NewInvalidArgumentsError(errors.New("")))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(registerProviderJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.RegisterProvider(nil).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInvalidArguments, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}

func TestRegisterProvider_Conflict(t *testing.T) {
	ctrl, mockAuthService, mockValidator, _, _, _, h := setup(t)
	defer ctrl.Finish()

	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockAuthService.EXPECT().RegisterProvider(gomock.Any(), gomock.Any()).Return(dto.User{}, &pgconn.PgError{Code: pgerrcode.UniqueViolation})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(registerProviderJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.RegisterProvider(nil).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, auth.MsgUserAlreadyExists, errResp.Message)
		assert.Equal(t, http.StatusConflict, rec.Code)
	}
}

func TestRegisterProvider_ServiceError(t *testing.T) {
	ctrl, mockAuthService, mockValidator, _, _, _, h := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().RegisterProvider(gomock.Any(), gomock.Any()).Return(dto.User{}, errors.New(""))
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(registerProviderJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.RegisterProvider(nil).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	}
}

func TestResendConfirmationLetter_Success(t *testing.T) {
	ctrl, mockAuthService, mockValidator, mockverify_jwtGenerator, _, _, h := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().GetUserConfirmationDetails(gomock.Any(), gomock.Any()).Return(userConfirmationDetailsDTO, nil)
	mockAuthService.EXPECT().SendConfirmationLetter(gomock.Any(), token, userDto.Email, userDto.FirstName).Return(nil)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockverify_jwtGenerator.EXPECT().GenerateJWT(userDto.ID, userDto.Email).Return(token, nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(emailJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ResendConfirmationLetter(mockverify_jwtGenerator).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestResendConfirmationLetter_BindError(t *testing.T) {
	ctrl, _, _, _, _, _, h := setup(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(emailJSON))
	rec := httptest.NewRecorder()

	h.ResendConfirmationLetter(nil).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgUnsupportedMedia, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}

func TestResendConfirmationLetter_InvalidArguments(t *testing.T) {
	ctrl, _, mockValidator, _, _, _, h := setup(t)
	defer ctrl.Finish()

	mockValidator.EXPECT().Validate(gomock.Any()).Return(rest.NewInvalidArgumentsError(errors.New("")))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(emailJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ResendConfirmationLetter(nil).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInvalidArguments, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}

func TestResendConfirmationLetter_Conflict(t *testing.T) {
	ctrl, mockAuthService, mockValidator, _, _, _, h := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().GetUserConfirmationDetails(gomock.Any(), gomock.Any()).Return(dto.UserConfirmationDetails{UserStatus: true}, nil)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(emailJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ResendConfirmationLetter(nil).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, auth.MsgUserAlreadyActivated, errResp.Message)
		assert.Equal(t, http.StatusConflict, rec.Code)
	}
}

func TestResendConfirmationLetter_NotFound(t *testing.T) {
	ctrl, mockAuthService, mockValidator, mockverify_jwtGenerator, _, _, h := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().GetUserConfirmationDetails(gomock.Any(), gomock.Any()).Return(dto.UserConfirmationDetails{}, pg_error.ErrNotFound)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(emailJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ResendConfirmationLetter(mockverify_jwtGenerator).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, app_error.MsgUserNotFound, errResp.Message)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	}
}

func TestResendConfirmationLetter_Already(t *testing.T) {
	ctrl, mockAuthService, mockValidator, mockverify_jwtGenerator, _, _, h := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().GetUserConfirmationDetails(gomock.Any(), gomock.Any()).Return(dto.UserConfirmationDetails{}, pg_error.ErrNotFound)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(emailJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ResendConfirmationLetter(mockverify_jwtGenerator).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, app_error.MsgUserNotFound, errResp.Message)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	}
}

func TestResendConfirmationLetter_ServiceError(t *testing.T) {
	ctrl, mockAuthService, mockValidator, _, _, _, h := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().GetUserConfirmationDetails(gomock.Any(), gomock.Any()).Return(dto.UserConfirmationDetails{}, errors.New(""))
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(registerProviderJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ResendConfirmationLetter(nil).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	}
}

func TestResendConfirmationLetter_MailError(t *testing.T) {

	ctrl, mockAuthService, mockValidator, mockverify_jwtGenerator, _, _, h := setup(t)
	defer ctrl.Finish()

	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockAuthService.EXPECT().GetUserConfirmationDetails(gomock.Any(), gomock.Any()).Return(userConfirmationDetailsDTO, nil)
	mockverify_jwtGenerator.EXPECT().GenerateJWT(userDto.ID, userDto.Email).Return("", rest.NewUnauthorizedError(errors.New(""), app_error.MsgInvalidToken))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(emailJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ResendConfirmationLetter(mockverify_jwtGenerator).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, app_error.MsgInvalidToken, errResp.Message)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	}
}

func TestLogin_Success(t *testing.T) {
	ctrl, mockAuthService, mockValidator, _, mockAccessGenerator, mockRefreshGenerator, h := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().AuthenticateUser(gomock.Any(), gomock.Any()).Return(UserIdentity, nil)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockAccessGenerator.EXPECT().GenerateJWT(userDto.ID, userDto.Role, userDto.UserStatus).Return(token, nil)
	mockRefreshGenerator.EXPECT().GenerateJWT(userDto.ID).Return(token, nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(loginJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(mockAccessGenerator, mockRefreshGenerator).ServeHTTP(rec, req)

	cookies := rec.Header().Values("Set-Cookie")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, cookies[0], fmt.Sprintf("access_token=%s; HttpOnly; Secure", token))
	assert.Contains(t, cookies[1], fmt.Sprintf("refresh_token=%s; HttpOnly; Secure", token))
}

func TestLogin_BindError(t *testing.T) {
	ctrl, _, _, _, _, _, h := setup(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(loginJSON))
	rec := httptest.NewRecorder()

	h.Login(nil, nil).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgUnsupportedMedia, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}

func TestLogin_InvalidArguments(t *testing.T) {
	ctrl, _, mockValidator, _, _, _, h := setup(t)
	defer ctrl.Finish()

	mockValidator.EXPECT().Validate(gomock.Any()).Return(rest.NewInvalidArgumentsError(errors.New("")))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(loginJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(nil, nil).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInvalidArguments, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}

func TestLogin_AuthError(t *testing.T) {
	ctrl, mockAuthService, mockValidator, _, _, _, h := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().AuthenticateUser(gomock.Any(), gomock.Any()).Return(dto.UserIdentity{}, hasher.ErrPasswordMismatch)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(loginJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(nil, nil).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, auth.MsgIncorrectEmailOrPass, errResp.Message)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	}
}

func TestLogin_ServiceError(t *testing.T) {
	ctrl, mockAuthService, mockValidator, _, _, _, h := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().AuthenticateUser(gomock.Any(), gomock.Any()).Return(dto.UserIdentity{}, errors.New(""))
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(loginJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(nil, nil).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	}
}

func TestLogin_AccessTokenError(t *testing.T) {
	ctrl, mockAuthService, mockValidator, _, mockAccessGenerator, _, h := setup(t)
	defer ctrl.Finish()

	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockAuthService.EXPECT().AuthenticateUser(gomock.Any(), gomock.Any()).Return(UserIdentity, nil)
	mockAccessGenerator.EXPECT().GenerateJWT(userDto.ID, userDto.Role, userDto.UserStatus).Return("", rest.NewInternalServerError(errors.New("")))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(loginJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(mockAccessGenerator, nil).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	}
}

func TestLogin_RefreshTokenError(t *testing.T) {
	ctrl, mockAuthService, mockValidator, _, mockAccessGenerator, mockRefreshGenerator, h := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().AuthenticateUser(gomock.Any(), gomock.Any()).Return(UserIdentity, nil)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockAccessGenerator.EXPECT().GenerateJWT(userDto.ID, userDto.Role, userDto.UserStatus).Return(token, nil)
	mockRefreshGenerator.EXPECT().GenerateJWT(userDto.ID).Return("", rest.NewInternalServerError(errors.New("")))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(loginJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(mockAccessGenerator, mockRefreshGenerator).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	}
}

func TestLogout_Success(t *testing.T) {
	ctrl, _, _, _, _, _, h := setup(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	h.Logout(rec, req)

	cookies := rec.Header().Values("Set-Cookie")

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, cookies[0], "access_token=; Expires=Thu, 01 Jan 1970 00:00:00 GMT; Max-Age=0; HttpOnly")
	assert.Contains(t, cookies[1], "refresh_token=; Expires=Thu, 01 Jan 1970 00:00:00 GMT; Max-Age=0; HttpOnly")
}

func TestRefresh_Success(t *testing.T) {
	ctrl, mockAuthService, _, _, mockAccessGenerator, _, h := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().GetUserRoleAndStatus(gomock.Any(), gomock.Any()).Return(UserRoleAndStatus, nil)
	mockAccessGenerator.EXPECT().GenerateJWT(userDto.ID, userDto.Role, userDto.UserStatus).Return(token, nil)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	ctx := ctx_util.SetJWTId(req.Context(), userDto.ID)
	ctx = ctx_util.SetJWTRole(ctx, enum.UserRole(userDto.Role))
	ctx = ctx_util.SetJWTUserStatus(ctx, userDto.UserStatus)

	h.Refresh(mockAccessGenerator).ServeHTTP(rec, req.WithContext(ctx))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Set-Cookie"), fmt.Sprintf("access_token=%s; HttpOnly; Secure", token))
}

func TestRefresh_NotFound(t *testing.T) {
	ctrl, mockAuthService, _, _, _, _, h := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().GetUserRoleAndStatus(gomock.Any(), gomock.Any()).Return(dto.UserRoleAndStatus{}, pgx.ErrNoRows)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	ctx := ctx_util.SetJWTId(req.Context(), userDto.ID)
	ctx = ctx_util.SetJWTRole(ctx, enum.UserRole(userDto.Role))
	ctx = ctx_util.SetJWTUserStatus(ctx, userDto.UserStatus)

	h.Refresh(nil).ServeHTTP(rec, req.WithContext(ctx))

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, app_error.MsgUserNotFound, errResp.Message)
	}
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.NotContains(t, rec.Header().Get("Set-Cookie"), fmt.Sprintf("access_token=%s; HttpOnly; Secure", token))
}

func TestRefresh_ServiceError(t *testing.T) {
	ctrl, mockAuthService, _, _, _, _, h := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().GetUserRoleAndStatus(gomock.Any(), gomock.Any()).Return(dto.UserRoleAndStatus{}, errors.New(""))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	ctx := ctx_util.SetJWTId(req.Context(), userDto.ID)
	ctx = ctx_util.SetJWTRole(ctx, enum.UserRole(userDto.Role))
	ctx = ctx_util.SetJWTUserStatus(ctx, userDto.UserStatus)

	h.Refresh(nil).ServeHTTP(rec, req.WithContext(ctx))

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
	}
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.NotContains(t, rec.Header().Get("Set-Cookie"), fmt.Sprintf("access_token=%s; HttpOnly; Secure", token))
}

func TestRefresh_JwtIdNotSet(t *testing.T) {
	t.Parallel()

	ctrl, _, _, _, _, _, h := setup(t)
	defer ctrl.Finish()

	handler := h.Refresh(nil)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	}
	assert.NotContains(t, rec.Header().Get("Set-Cookie"), fmt.Sprintf("access_token=%s; HttpOnly; Secure", token))
}

func TestVerifyEmail_Success(t *testing.T) {
	ctrl, mockAuthService, _, mockVerifyJWT, _, _, h := setup(t)
	defer ctrl.Finish()

	mockVerifyJWT.EXPECT().VerifyJWT(gomock.Any()).Return(verifyClaims, nil)
	mockAuthService.EXPECT().VerifyUser(gomock.Any(), gomock.Any(), gomock.Any(), int64(1), verifyClaims.Email).Return(nil)
	mockAuthService.EXPECT().SendVerifiedLetter(verifyClaims.Email).Return(nil)

	q := make(url.Values)
	q.Set("token", token)

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	h.VerifyEmail(mockVerifyJWT).ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	time.Sleep(500 * time.Millisecond)
}

func TestVerifyEmail_InvalidToken(t *testing.T) {
	ctrl, _, _, mockVerifyJWT, _, _, h := setup(t)
	defer ctrl.Finish()

	mockVerifyJWT.EXPECT().VerifyJWT(token).Return(verify_jwt.VerifyClaims{}, rest.NewUnauthorizedError(errors.New(""), app_error.MsgInvalidToken))

	q := make(url.Values)
	q.Set("token", token)
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	h.VerifyEmail(mockVerifyJWT).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, app_error.MsgInvalidToken, errResp.Message)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	}
}

func TestVerifyEmail_ParseIDError(t *testing.T) {
	ctrl, _, _, mockVerifyJWT, _, _, h := setup(t)
	defer ctrl.Finish()

	mockVerifyJWT.EXPECT().VerifyJWT(token).Return(verify_jwt.VerifyClaims{ID: "invalid-id"}, nil)

	q := make(url.Values)
	q.Set("token", token)
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	h.VerifyEmail(mockVerifyJWT).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	}
}

func TestVerifyEmail_AlreadyActivated(t *testing.T) {
	ctrl, mockAuthService, _, mockVerifyJWT, _, _, h := setup(t)
	defer ctrl.Finish()

	mockVerifyJWT.EXPECT().VerifyJWT(token).Return(verifyClaims, nil)
	mockAuthService.EXPECT().VerifyUser(gomock.Any(), gomock.Any(), gomock.Any(), int64(1), verifyClaims.Email).Return(redis.TxFailedErr)

	q := make(url.Values)
	q.Set("token", token)
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	h.VerifyEmail(mockVerifyJWT).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, auth.MsgTokenAlreadyUsed, errResp.Message)
		assert.Equal(t, http.StatusConflict, rec.Code)
	}
}

func TestVerifyEmail_NotFound(t *testing.T) {
	ctrl, mockAuthService, _, mockVerifyJWT, _, _, h := setup(t)
	defer ctrl.Finish()

	mockVerifyJWT.EXPECT().VerifyJWT(token).Return(verifyClaims, nil)
	mockAuthService.EXPECT().VerifyUser(gomock.Any(), gomock.Any(), gomock.Any(), int64(1), verifyClaims.Email).Return(pg_error.ErrNotFound)

	q := make(url.Values)
	q.Set("token", token)
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	h.VerifyEmail(mockVerifyJWT).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, app_error.MsgUserNotFound, errResp.Message)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	}
}

func TestVerifyEmail_ServiceError(t *testing.T) {
	ctrl, mockAuthService, _, mockVerifyJWT, _, _, h := setup(t)
	defer ctrl.Finish()

	mockVerifyJWT.EXPECT().VerifyJWT(token).Return(verifyClaims, nil)
	mockAuthService.EXPECT().VerifyUser(gomock.Any(), gomock.Any(), gomock.Any(), int64(1), verifyClaims.Email).Return(errors.New(""))

	q := make(url.Values)
	q.Set("token", token)
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	h.VerifyEmail(mockVerifyJWT).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	}
}

func TestVerifyEmail_MailError(t *testing.T) {
	ctrl, mockAuthService, _, mockVerifyJWT, _, _, h := setup(t)
	defer ctrl.Finish()

	mockVerifyJWT.EXPECT().VerifyJWT(gomock.Any()).Return(verifyClaims, nil)
	mockAuthService.EXPECT().VerifyUser(gomock.Any(), gomock.Any(), gomock.Any(), int64(1), verifyClaims.Email).Return(nil)
	mockAuthService.EXPECT().SendVerifiedLetter(verifyClaims.Email).Return(errors.New(""))

	q := make(url.Values)
	q.Set("token", token)
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	h.VerifyEmail(mockVerifyJWT).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	time.Sleep(500 * time.Millisecond)
}
