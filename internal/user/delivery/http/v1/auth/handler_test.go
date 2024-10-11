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
	mockAuthJwt "github.com/hexley21/fixup/internal/common/auth_jwt/mock"
	"github.com/hexley21/fixup/internal/common/enum"
	"github.com/hexley21/fixup/internal/common/util/ctx_util"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/auth"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/dto"
	mockRefreshJwt "github.com/hexley21/fixup/internal/user/jwt/refresh_jwt/mock"
	"github.com/hexley21/fixup/internal/user/jwt/verify_jwt"
	mockVerifyJwt "github.com/hexley21/fixup/internal/user/jwt/verify_jwt/mock"
	"github.com/hexley21/fixup/internal/user/service"
	mockService "github.com/hexley21/fixup/internal/user/service/mock"
	"github.com/hexley21/fixup/pkg/hasher"
	"github.com/hexley21/fixup/pkg/http/binder/std_binder"
	"github.com/hexley21/fixup/pkg/http/handler"
	"github.com/hexley21/fixup/pkg/http/json/std_json"
	"github.com/hexley21/fixup/pkg/http/rest"
	"github.com/hexley21/fixup/pkg/http/writer/json_writer"
	"github.com/hexley21/fixup/pkg/infra/postgres/pg_error"
	"github.com/hexley21/fixup/pkg/logger/std_logger"
	mockValidator "github.com/hexley21/fixup/pkg/validator/mock"
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

	userConfirmationDetailsDTO = service.UserConfirmationDetails{
		ID:         "1",
		UserStatus: false,
		Firstname:  "Larry",
	}

	UserIdentity = service.UserIdentity{
		ID:         "1",
		Role:       string(enum.UserRoleADMIN),
		UserStatus: true,
	}

	UserRoleAndStatus = service.UserRoleAndStatus{
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
	authServiceMock *mockService.MockAuthService,
	validatorMock *mockValidator.MockValidator,
	verifyJWTManager *mockVerifyJwt.MockManager,
	accessJWTGeneratorMock *mockAuthJwt.MockGenerator,
	refreshJWTGeneratorMock *mockRefreshJwt.MockGenerator,
	h *auth.Handler,
) {
	ctrl = gomock.NewController(t)
	authServiceMock = mockService.NewMockAuthService(ctrl)
	validatorMock = mockValidator.NewMockValidator(ctrl)
	verifyJWTManager = mockVerifyJwt.NewMockManager(ctrl)
	accessJWTGeneratorMock = mockAuthJwt.NewMockGenerator(ctrl)
	refreshJWTGeneratorMock = mockRefreshJwt.NewMockGenerator(ctrl)

	logger := std_logger.New()
	jsonManager := std_json.New()

	h = auth.NewHandler(
		handler.NewComponents(logger, std_binder.New(jsonManager), validatorMock, json_writer.New(logger, jsonManager)),
		authServiceMock,
	)

	return
}

func TestRegisterCustomer_Success(t *testing.T) {
	ctrl, authServiceMock, validatorMock, verifyJWTManager, _, _, h := setup(t)
	defer ctrl.Finish()

	authServiceMock.EXPECT().RegisterCustomer(gomock.Any(), gomock.Any()).Return(userDto, nil)
	authServiceMock.EXPECT().SendConfirmationLetter(gomock.Any(), token, userDto.Email, userDto.FirstName).Return(nil)
	validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)
	verifyJWTManager.EXPECT().Generate(userDto.ID, userDto.Email).Return(token, nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(registerCustomerJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.RegisterCustomer(verifyJWTManager).ServeHTTP(rec, req)

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
	ctrl, _, validatorMock, _, _, _, h := setup(t)
	defer ctrl.Finish()

	validatorMock.EXPECT().Validate(gomock.Any()).Return(rest.NewInvalidArgumentsError(nil))

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
	ctrl, authServiceMock, validatorMock, _, _, _, h := setup(t)
	defer ctrl.Finish()

	uniqueViolationErr := &pgconn.PgError{Code: pgerrcode.UniqueViolation}

	authServiceMock.EXPECT().RegisterCustomer(gomock.Any(), gomock.Any()).Return(dto.User{}, uniqueViolationErr)
	validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)

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
	ctrl, authServiceMock, validatorMock, _, _, _, h := setup(t)
	defer ctrl.Finish()

	authServiceMock.EXPECT().RegisterCustomer(gomock.Any(), gomock.Any()).Return(dto.User{}, errors.New(""))
	validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)

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
	ctrl, authServiceMock, validatorMock, verifyJWTManager, _, _, h := setup(t)
	defer ctrl.Finish()

	authServiceMock.EXPECT().RegisterProvider(gomock.Any(), gomock.Any()).Return(userDto, nil)
	authServiceMock.EXPECT().SendConfirmationLetter(gomock.Any(), token, userDto.Email, userDto.FirstName).Return(nil)
	validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)
	verifyJWTManager.EXPECT().Generate(userDto.ID, userDto.Email).Return(token, nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(registerProviderJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.RegisterProvider(verifyJWTManager).ServeHTTP(rec, req)

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
	ctrl, _, validatorMock, _, _, _, h := setup(t)
	defer ctrl.Finish()

	validatorMock.EXPECT().Validate(gomock.Any()).Return(rest.NewInvalidArgumentsError(errors.New("")))

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
	ctrl, authServiceMock, validatorMock, _, _, _, h := setup(t)
	defer ctrl.Finish()

	validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)
	authServiceMock.EXPECT().RegisterProvider(gomock.Any(), gomock.Any()).Return(dto.User{}, &pgconn.PgError{Code: pgerrcode.UniqueViolation})

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
	ctrl, authServiceMock, validatorMock, _, _, _, h := setup(t)
	defer ctrl.Finish()

	authServiceMock.EXPECT().RegisterProvider(gomock.Any(), gomock.Any()).Return(dto.User{}, errors.New(""))
	validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)

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
	ctrl, authServiceMock, validatorMock, verifyJWTManager, _, _, h := setup(t)
	defer ctrl.Finish()

	authServiceMock.EXPECT().GetUserConfirmationDetails(gomock.Any(), gomock.Any()).Return(userConfirmationDetailsDTO, nil)
	authServiceMock.EXPECT().SendConfirmationLetter(gomock.Any(), token, userDto.Email, userDto.FirstName).Return(nil)
	validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)
	verifyJWTManager.EXPECT().Generate(userDto.ID, userDto.Email).Return(token, nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(emailJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ResendConfirmationLetter(verifyJWTManager).ServeHTTP(rec, req)

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
	ctrl, _, validatorMock, _, _, _, h := setup(t)
	defer ctrl.Finish()

	validatorMock.EXPECT().Validate(gomock.Any()).Return(rest.NewInvalidArgumentsError(errors.New("")))

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
	ctrl, authServiceMock, validatorMock, _, _, _, h := setup(t)
	defer ctrl.Finish()

	authServiceMock.EXPECT().GetUserConfirmationDetails(gomock.Any(), gomock.Any()).Return(service.UserConfirmationDetails{UserStatus: true}, nil)
	validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)

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
	ctrl, authServiceMock, validatorMock, verifyJWTManager, _, _, h := setup(t)
	defer ctrl.Finish()

	authServiceMock.EXPECT().GetUserConfirmationDetails(gomock.Any(), gomock.Any()).Return(service.UserConfirmationDetails{}, pg_error.ErrNotFound)
	validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(emailJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ResendConfirmationLetter(verifyJWTManager).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, app_error.MsgUserNotFound, errResp.Message)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	}
}

func TestResendConfirmationLetter_Already(t *testing.T) {
	ctrl, authServiceMock, validatorMock, verifyJWTManager, _, _, h := setup(t)
	defer ctrl.Finish()

	authServiceMock.EXPECT().GetUserConfirmationDetails(gomock.Any(), gomock.Any()).Return(service.UserConfirmationDetails{}, pg_error.ErrNotFound)
	validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(emailJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ResendConfirmationLetter(verifyJWTManager).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, app_error.MsgUserNotFound, errResp.Message)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	}
}

func TestResendConfirmationLetter_ServiceError(t *testing.T) {
	ctrl, authServiceMock, validatorMock, _, _, _, h := setup(t)
	defer ctrl.Finish()

	authServiceMock.EXPECT().GetUserConfirmationDetails(gomock.Any(), gomock.Any()).Return(service.UserConfirmationDetails{}, errors.New(""))
	validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)

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

	ctrl, authServiceMock, validatorMock, verifyJWTManager, _, _, h := setup(t)
	defer ctrl.Finish()

	validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)
	authServiceMock.EXPECT().GetUserConfirmationDetails(gomock.Any(), gomock.Any()).Return(userConfirmationDetailsDTO, nil)
	verifyJWTManager.EXPECT().Generate(userDto.ID, userDto.Email).Return("", rest.NewUnauthorizedError(errors.New(""), app_error.MsgInvalidToken))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(emailJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ResendConfirmationLetter(verifyJWTManager).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, app_error.MsgInvalidToken, errResp.Message)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	}
}

func TestLogin_Success(t *testing.T) {
	ctrl, authServiceMock, validatorMock, _, accessJWTGeneratorMock, refreshJWTGeneratorMock, h := setup(t)
	defer ctrl.Finish()

	authServiceMock.EXPECT().AuthenticateUser(gomock.Any(), gomock.Any()).Return(UserIdentity, nil)
	validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)
	accessJWTGeneratorMock.EXPECT().Generate(userDto.ID, userDto.Role, userDto.UserStatus).Return(token, nil)
	refreshJWTGeneratorMock.EXPECT().Generate(userDto.ID).Return(token, nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(loginJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(accessJWTGeneratorMock, refreshJWTGeneratorMock).ServeHTTP(rec, req)

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
	ctrl, _, validatorMock, _, _, _, h := setup(t)
	defer ctrl.Finish()

	validatorMock.EXPECT().Validate(gomock.Any()).Return(rest.NewInvalidArgumentsError(errors.New("")))

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
	ctrl, authServiceMock, validatorMock, _, _, _, h := setup(t)
	defer ctrl.Finish()

	authServiceMock.EXPECT().AuthenticateUser(gomock.Any(), gomock.Any()).Return(service.UserIdentity{}, hasher.ErrPasswordMismatch)
	validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)

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
	ctrl, authServiceMock, validatorMock, _, _, _, h := setup(t)
	defer ctrl.Finish()

	authServiceMock.EXPECT().AuthenticateUser(gomock.Any(), gomock.Any()).Return(service.UserIdentity{}, errors.New(""))
	validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)

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
	ctrl, authServiceMock, validatorMock, _, accessJWTGeneratorMock, _, h := setup(t)
	defer ctrl.Finish()

	validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)
	authServiceMock.EXPECT().AuthenticateUser(gomock.Any(), gomock.Any()).Return(UserIdentity, nil)
	accessJWTGeneratorMock.EXPECT().Generate(userDto.ID, userDto.Role, userDto.UserStatus).Return("", rest.NewInternalServerError(errors.New("")))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(loginJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(accessJWTGeneratorMock, nil).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	}
}

func TestLogin_RefreshTokenError(t *testing.T) {
	ctrl, authServiceMock, validatorMock, _, accessJWTGeneratorMock, refreshJWTGeneratorMock, h := setup(t)
	defer ctrl.Finish()

	authServiceMock.EXPECT().AuthenticateUser(gomock.Any(), gomock.Any()).Return(UserIdentity, nil)
	validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)
	accessJWTGeneratorMock.EXPECT().Generate(userDto.ID, userDto.Role, userDto.UserStatus).Return(token, nil)
	refreshJWTGeneratorMock.EXPECT().Generate(userDto.ID).Return("", rest.NewInternalServerError(errors.New("")))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(loginJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(accessJWTGeneratorMock, refreshJWTGeneratorMock).ServeHTTP(rec, req)

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
	ctrl, authServiceMock, _, _, accessJWTGeneratorMock, _, h := setup(t)
	defer ctrl.Finish()

	authServiceMock.EXPECT().GetUserRoleAndStatus(gomock.Any(), gomock.Any()).Return(UserRoleAndStatus, nil)
	accessJWTGeneratorMock.EXPECT().Generate(userDto.ID, userDto.Role, userDto.UserStatus).Return(token, nil)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	ctx := ctx_util.SetJWTId(req.Context(), userDto.ID)
	ctx = ctx_util.SetJWTRole(ctx, enum.UserRole(userDto.Role))
	ctx = ctx_util.SetJWTUserStatus(ctx, userDto.UserStatus)

	h.Refresh(accessJWTGeneratorMock).ServeHTTP(rec, req.WithContext(ctx))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Set-Cookie"), fmt.Sprintf("access_token=%s; HttpOnly; Secure", token))
}

func TestRefresh_NotFound(t *testing.T) {
	ctrl, authServiceMock, _, _, _, _, h := setup(t)
	defer ctrl.Finish()

	authServiceMock.EXPECT().GetUserRoleAndStatus(gomock.Any(), gomock.Any()).Return(service.UserRoleAndStatus{}, pgx.ErrNoRows)

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
	ctrl, authServiceMock, _, _, _, _, h := setup(t)
	defer ctrl.Finish()

	authServiceMock.EXPECT().GetUserRoleAndStatus(gomock.Any(), gomock.Any()).Return(service.UserRoleAndStatus{}, errors.New(""))

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

	refreshHandler := h.Refresh(nil)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	refreshHandler.ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	}
	assert.NotContains(t, rec.Header().Get("Set-Cookie"), fmt.Sprintf("access_token=%s; HttpOnly; Secure", token))
}

func TestVerifyEmail_Success(t *testing.T) {
	ctrl, authServiceMock, _, vrfJWTManager, _, _, h := setup(t)
	defer ctrl.Finish()

	vrfJWTManager.EXPECT().Verify(gomock.Any()).Return(verifyClaims, nil)
	authServiceMock.EXPECT().VerifyUser(gomock.Any(), gomock.Any(), gomock.Any(), int64(1)).Return(nil)
	authServiceMock.EXPECT().SendVerifiedLetter(verifyClaims.Email).Return(nil)

	q := make(url.Values)
	q.Set("token", token)

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	h.VerifyEmail(vrfJWTManager).ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	time.Sleep(500 * time.Millisecond)
}

func TestVerifyEmail_InvalidToken(t *testing.T) {
	ctrl, _, _, vrfJWTManager, _, _, h := setup(t)
	defer ctrl.Finish()

	vrfJWTManager.EXPECT().Verify(token).Return(verify_jwt.VerifyClaims{}, rest.NewUnauthorizedError(errors.New(""), app_error.MsgInvalidToken))

	q := make(url.Values)
	q.Set("token", token)
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	h.VerifyEmail(vrfJWTManager).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, app_error.MsgInvalidToken, errResp.Message)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	}
}

func TestVerifyEmail_ParseIDError(t *testing.T) {
	ctrl, _, _, vrfJWTManager, _, _, h := setup(t)
	defer ctrl.Finish()

	vrfJWTManager.EXPECT().Verify(token).Return(verify_jwt.VerifyClaims{ID: "invalid-id"}, nil)

	q := make(url.Values)
	q.Set("token", token)
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	h.VerifyEmail(vrfJWTManager).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	}
}

func TestVerifyEmail_AlreadyActivated(t *testing.T) {
	ctrl, authServiceMock, _, vrfJWTManager, _, _, h := setup(t)
	defer ctrl.Finish()

	vrfJWTManager.EXPECT().Verify(token).Return(verifyClaims, nil)
	authServiceMock.EXPECT().VerifyUser(gomock.Any(), gomock.Any(), gomock.Any(), int64(1)).Return(redis.TxFailedErr)

	q := make(url.Values)
	q.Set("token", token)
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	h.VerifyEmail(vrfJWTManager).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, auth.MsgTokenAlreadyUsed, errResp.Message)
		assert.Equal(t, http.StatusConflict, rec.Code)
	}
}

func TestVerifyEmail_NotFound(t *testing.T) {
	ctrl, authServiceMock, _, vrfJWTManager, _, _, h := setup(t)
	defer ctrl.Finish()

	vrfJWTManager.EXPECT().Verify(token).Return(verifyClaims, nil)
	authServiceMock.EXPECT().VerifyUser(gomock.Any(), gomock.Any(), gomock.Any(), int64(1)).Return(pg_error.ErrNotFound)

	q := make(url.Values)
	q.Set("token", token)
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	h.VerifyEmail(vrfJWTManager).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, app_error.MsgUserNotFound, errResp.Message)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	}
}

func TestVerifyEmail_ServiceError(t *testing.T) {
	ctrl, authServiceMock, _, vrfJWTManager, _, _, h := setup(t)
	defer ctrl.Finish()

	vrfJWTManager.EXPECT().Verify(token).Return(verifyClaims, nil)
	authServiceMock.EXPECT().VerifyUser(gomock.Any(), gomock.Any(), gomock.Any(), int64(1)).Return(errors.New(""))

	q := make(url.Values)
	q.Set("token", token)
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	h.VerifyEmail(vrfJWTManager).ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	}
}

func TestVerifyEmail_MailError(t *testing.T) {
	ctrl, authServiceMock, _, vrfJWTManager, _, _, h := setup(t)
	defer ctrl.Finish()

	vrfJWTManager.EXPECT().Verify(gomock.Any()).Return(verifyClaims, nil)
	authServiceMock.EXPECT().VerifyUser(gomock.Any(), gomock.Any(), gomock.Any(), int64(1)).Return(nil)
	authServiceMock.EXPECT().SendVerifiedLetter(verifyClaims.Email).Return(errors.New(""))

	q := make(url.Values)
	q.Set("token", token)
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	h.VerifyEmail(vrfJWTManager).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	time.Sleep(500 * time.Millisecond)
}
