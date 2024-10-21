package auth_test

// TODO: Refactor auth handler tests
// import (
// 	"encoding/json"
// 	"errors"
// 	"fmt"
// 	"net/http"
// 	"net/http/httptest"
// 	"net/url"
// 	"strings"
// 	"testing"
// 	"time"

// 	"github.com/golang-jwt/jwt/v5"
// 	"github.com/hexley21/fixup/internal/common/app_error"
// 	mockAuthJwt "github.com/hexley21/fixup/internal/common/auth_jwt/mock"
// 	"github.com/hexley21/fixup/internal/common/enum"
// 	"github.com/hexley21/fixup/internal/common/util/ctx_util"
// 	"github.com/hexley21/fixup/internal/user/delivery/http/v1/auth"
// 	"github.com/hexley21/fixup/internal/user/delivery/http/v1/dto"
// 	mockRefreshJwt "github.com/hexley21/fixup/internal/user/jwt/refresh_jwt/mock"
// 	"github.com/hexley21/fixup/internal/user/jwt/verify_jwt"
// 	mockVerifyJwt "github.com/hexley21/fixup/internal/user/jwt/verify_jwt/mock"
// 	"github.com/hexley21/fixup/internal/user/service"
// 	mockService "github.com/hexley21/fixup/internal/user/service/mock"
// 	"github.com/hexley21/fixup/pkg/hasher"
// 	"github.com/hexley21/fixup/pkg/http/binder/std_binder"
// 	"github.com/hexley21/fixup/pkg/http/handler"
// 	"github.com/hexley21/fixup/pkg/http/json/std_json"
// 	"github.com/hexley21/fixup/pkg/http/rest"
// 	"github.com/hexley21/fixup/pkg/http/writer/json_writer"
// 	"github.com/hexley21/fixup/pkg/infra/postgres/pg_error"
// 	"github.com/hexley21/fixup/pkg/logger/std_logger"
// 	mockValidator "github.com/hexley21/fixup/pkg/validator/mock"
// 	"github.com/jackc/pgerrcode"
// 	"github.com/jackc/pgx/v5"
// 	"github.com/jackc/pgx/v5/pgconn"
// 	"github.com/redis/go-redis/v9"
// 	"github.com/stretchr/testify/assert"
// 	"go.uber.org/mock/gomock"
// )

// type testCase struct {
// 	name          string
// 	mockSetup     func()
// 	input         string
// 	skipHeader    bool
// 	expectedCode  int
// 	expectedError string
// 	cookieCheck   bool
// }

// var (
// 	userDTO = dto.User{
// 		ID:          "1",
// 		FirstName:   "Larry",
// 		LastName:    "Page",
// 		PhoneNumber: "995111222333",
// 		Email:       "larry@page.com",
// 		PictureUrl:  "larrypage.png",
// 		Role:        string(enum.UserRoleADMIN),
// 		Active:      true,
// 		CreatedAt:   time.Now(),
// 	}

// 	userConfirmationDetailsDTO = service.UserConfirmationDetails{
// 		ID:        "1",
// 		Active:    false,
// 		Firstname: "Larry",
// 	}

// 	UserIdentity = service.UserIdentity{
// 		ID:     "1",
// 		Role:   string(enum.UserRoleADMIN),
// 		Active: true,
// 	}

// 	UserRoleAndStatus = service.UserRoleAndStatus{
// 		Role:   string(enum.UserRoleADMIN),
// 		Active: true,
// 	}

// 	verifyClaims = verify_jwt.VerifyClaims{
// 		ID:    "1",
// 		Email: "larry@page.com",
// 		RegisteredClaims: jwt.RegisteredClaims{
// 			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
// 		},
// 	}

// 	registerCustomerJSON = `{"email": "larry@page.com", "phone_number": "995111222333", "first_name": "Larry", "last_name": "Page", "password": "larrypage123"}`
// 	registerProviderJSON = `{"email": "larry@page.com", "phone_number": "995111222333", "first_name": "Larry", "last_name": "Page", "password": "larrypage123", "personal_id_number": "1234567890"}`
// 	emailJSON            = `{"email": "larry@page.com"}`
// 	loginJSON            = `{"email": "larry@page.com", "password": "larry@page.com"}`

// 	token = "Ehx0DNg86zL"
// )

// func setup(t *testing.T) (
// 	ctrl *gomock.Controller,
// 	authServiceMock *mockService.MockAuthService,
// 	validatorMock *mockValidator.MockValidator,
// 	verifyJWTManagerMock *mockVerifyJwt.MockManager,
// 	accessJWTGeneratorMock *mockAuthJwt.MockGenerator,
// 	refreshJWTGeneratorMock *mockRefreshJwt.MockGenerator,
// 	h *auth.Handler,
// ) {
// 	ctrl = gomock.NewController(t)
// 	authServiceMock = mockService.NewMockAuthService(ctrl)
// 	validatorMock = mockValidator.NewMockValidator(ctrl)
// 	verifyJWTManagerMock = mockVerifyJwt.NewMockManager(ctrl)
// 	accessJWTGeneratorMock = mockAuthJwt.NewMockGenerator(ctrl)
// 	refreshJWTGeneratorMock = mockRefreshJwt.NewMockGenerator(ctrl)

// 	logger := std_logger.New()
// 	jsonManager := std_json.New()

// 	h = auth.NewHandler(
// 		handler.NewComponents(logger, std_binder.New(jsonManager), validatorMock, json_writer.New(logger, jsonManager)),
// 		authServiceMock,
// 	)

// 	return
// }

// func TestRegisterCustomer(t *testing.T) {
// 	ctrl, authServiceMock, validatorMock, verifyJWTManagerMock, _, _, h := setup(t)
// 	defer ctrl.Finish()

// 	tests := []testCase{
// 		{
// 			name: "Success",
// 			mockSetup: func() {
// 				authServiceMock.EXPECT().RegisterCustomer(gomock.Any(), gomock.Any()).Return(userDTO, nil)
// 				authServiceMock.EXPECT().SendConfirmationLetter(gomock.Any(), token, userDTO.Email, userDTO.FirstName).Return(nil)
// 				validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)
// 				verifyJWTManagerMock.EXPECT().Generate(userDTO.ID, userDTO.Email).Return(token, nil)
// 			},
// 			input:        registerCustomerJSON,
// 			expectedCode: http.StatusCreated,
// 		},
// 		{
// 			name:          "Bind Error",
// 			mockSetup:     func() {},
// 			input:         `invalid-json`,
// 			skipHeader:    true,
// 			expectedCode:  http.StatusBadRequest,
// 			expectedError: rest.MsgUnsupportedMedia,
// 		},
// 		{
// 			name: "Invalid Arguments",
// 			mockSetup: func() {
// 				validatorMock.EXPECT().Validate(gomock.Any()).Return(rest.NewInvalidArgumentsError(nil))
// 			},
// 			input:         registerCustomerJSON,
// 			expectedCode:  http.StatusBadRequest,
// 			expectedError: rest.MsgInvalidArguments,
// 		},
// 		{
// 			name: "User Already Exists",
// 			mockSetup: func() {
// 				authServiceMock.EXPECT().RegisterCustomer(gomock.Any(), gomock.Any()).Return(dto.User{}, &pgconn.PgError{Code: pgerrcode.UniqueViolation})
// 				validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)
// 			},
// 			input:         registerCustomerJSON,
// 			expectedCode:  http.StatusConflict,
// 			expectedError: auth.MsgUserAlreadyExists,
// 		},
// 		{
// 			name: "Internal Error",
// 			mockSetup: func() {
// 				authServiceMock.EXPECT().RegisterCustomer(gomock.Any(), gomock.Any()).Return(dto.User{}, errors.New(""))
// 				validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)
// 			},
// 			input:         registerCustomerJSON,
// 			expectedCode:  http.StatusInternalServerError,
// 			expectedError: rest.MsgInternalServerError,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			tt.mockSetup()

// 			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.input))
// 			if !tt.skipHeader {
// 				req.Header.Set("Content-Type", "application/json")
// 			}
// 			rec := httptest.NewRecorder()

// 			h.RegisterCustomer(verifyJWTManagerMock).ServeHTTP(rec, req)

// 			if tt.expectedError != "" {
// 				var errResp rest.ErrorResponse
// 				if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
// 					assert.Equal(t, tt.expectedError, errResp.Message)
// 				}
// 			}

// 			assert.Equal(t, tt.expectedCode, rec.Code)
// 			if tt.expectedCode == http.StatusOK {
// 				time.Sleep(time.Second / 2)
// 			}
// 		})
// 	}
// }

// func TestRegisterProvider(t *testing.T) {
// 	ctrl, authServiceMock, validatorMock, verifyJWTManagerMock, _, _, h := setup(t)
// 	defer ctrl.Finish()

// 	tests := []testCase{
// 		{
// 			name: "Success",
// 			mockSetup: func() {
// 				authServiceMock.EXPECT().RegisterProvider(gomock.Any(), gomock.Any()).Return(userDTO, nil)
// 				authServiceMock.EXPECT().SendConfirmationLetter(gomock.Any(), token, userDTO.Email, userDTO.FirstName).Return(nil)
// 				validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)
// 				verifyJWTManagerMock.EXPECT().Generate(userDTO.ID, userDTO.Email).Return(token, nil)
// 			},
// 			input:        registerProviderJSON,
// 			expectedCode: http.StatusCreated,
// 		},
// 		{
// 			name:          "Bind Error",
// 			mockSetup:     func() {},
// 			input:         `invalid-json`,
// 			skipHeader:    true,
// 			expectedCode:  http.StatusBadRequest,
// 			expectedError: rest.MsgUnsupportedMedia,
// 		},
// 		{
// 			name: "Invalid Arguments",
// 			mockSetup: func() {
// 				validatorMock.EXPECT().Validate(gomock.Any()).Return(rest.NewInvalidArgumentsError(errors.New("")))
// 			},
// 			input:         registerProviderJSON,
// 			expectedCode:  http.StatusBadRequest,
// 			expectedError: rest.MsgInvalidArguments,
// 		},
// 		{
// 			name: "User Already Exists",
// 			mockSetup: func() {
// 				validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)
// 				authServiceMock.EXPECT().RegisterProvider(gomock.Any(), gomock.Any()).Return(dto.User{}, &pgconn.PgError{Code: pgerrcode.UniqueViolation})
// 			},
// 			input:         registerProviderJSON,
// 			expectedCode:  http.StatusConflict,
// 			expectedError: auth.MsgUserAlreadyExists,
// 		},
// 		{
// 			name: "Service Error",
// 			mockSetup: func() {
// 				authServiceMock.EXPECT().RegisterProvider(gomock.Any(), gomock.Any()).Return(dto.User{}, errors.New(""))
// 				validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)
// 			},
// 			input:         registerProviderJSON,
// 			expectedCode:  http.StatusInternalServerError,
// 			expectedError: rest.MsgInternalServerError,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			tt.mockSetup()

// 			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.input))
// 			if !tt.skipHeader {
// 				req.Header.Set("Content-Type", "application/json")
// 			}
// 			rec := httptest.NewRecorder()

// 			h.RegisterProvider(verifyJWTManagerMock).ServeHTTP(rec, req)

// 			if tt.expectedError != "" {
// 				var errResp rest.ErrorResponse
// 				if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
// 					assert.Equal(t, tt.expectedError, errResp.Message)
// 				}
// 			}

// 			assert.Equal(t, tt.expectedCode, rec.Code)
// 			if tt.expectedCode == http.StatusOK {
// 				time.Sleep(time.Second / 2)
// 			}
// 		})
// 	}
// }

// func TestResendConfirmationLetter(t *testing.T) {
// 	ctrl, authServiceMock, validatorMock, verifyJWTManagerMock, _, _, h := setup(t)
// 	defer ctrl.Finish()

// 	tests := []struct {
// 		name          string
// 		mockSetup     func()
// 		input         string
// 		skipHeader    bool
// 		expectedCode  int
// 		expectedError string
// 	}{
// 		{
// 			name: "Success",
// 			mockSetup: func() {
// 				authServiceMock.EXPECT().GetUserConfirmationDetails(gomock.Any(), gomock.Any()).Return(userConfirmationDetailsDTO, nil)
// 				authServiceMock.EXPECT().SendConfirmationLetter(gomock.Any(), token, userDTO.Email, userDTO.FirstName).Return(nil)
// 				validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)
// 				verifyJWTManagerMock.EXPECT().Generate(userDTO.ID, userDTO.Email).Return(token, nil)

// 			},
// 			input:        emailJSON,
// 			expectedCode: http.StatusNoContent,
// 		},
// 		{
// 			name:          "Bind Error",
// 			mockSetup:     func() {},
// 			input:         `invalid-json`,
// 			skipHeader:    true,
// 			expectedCode:  http.StatusBadRequest,
// 			expectedError: rest.MsgUnsupportedMedia,
// 		},
// 		{
// 			name: "Invalid Arguments",
// 			mockSetup: func() {
// 				validatorMock.EXPECT().Validate(gomock.Any()).Return(rest.NewInvalidArgumentsError(errors.New("validation error")))
// 			},
// 			input:         emailJSON,
// 			expectedCode:  http.StatusBadRequest,
// 			expectedError: rest.MsgInvalidArguments,
// 		},
// 		{
// 			name: "User Already Activated",
// 			mockSetup: func() {
// 				authServiceMock.EXPECT().GetUserConfirmationDetails(gomock.Any(), gomock.Any()).Return(service.UserConfirmationDetails{Active: true}, nil)
// 				validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)
// 			},
// 			input:         emailJSON,
// 			expectedCode:  http.StatusConflict,
// 			expectedError: auth.MsgUserAlreadyActivated,
// 		},
// 		{
// 			name: "Not Found",
// 			mockSetup: func() {
// 				authServiceMock.EXPECT().GetUserConfirmationDetails(gomock.Any(), gomock.Any()).Return(service.UserConfirmationDetails{}, pg_error.ErrNotFound)
// 				validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)
// 			},
// 			input:         emailJSON,
// 			expectedCode:  http.StatusNotFound,
// 			expectedError: app_error.MsgUserNotFound,
// 		},
// 		{
// 			name: "Service Error",
// 			mockSetup: func() {
// 				authServiceMock.EXPECT().GetUserConfirmationDetails(gomock.Any(), gomock.Any()).Return(service.UserConfirmationDetails{}, errors.New("service error"))
// 				validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)
// 			},
// 			input:         emailJSON,
// 			expectedCode:  http.StatusInternalServerError,
// 			expectedError: rest.MsgInternalServerError,
// 		},
// 		{
// 			name: "Mail Error",
// 			mockSetup: func() {
// 				validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)
// 				authServiceMock.EXPECT().GetUserConfirmationDetails(gomock.Any(), gomock.Any()).Return(userConfirmationDetailsDTO, nil)
// 				verifyJWTManagerMock.EXPECT().Generate(userDTO.ID, userDTO.Email).Return("", rest.NewUnauthorizedError(errors.New(""), app_error.MsgInvalidToken))
// 			},
// 			input:         emailJSON,
// 			expectedCode:  http.StatusUnauthorized,
// 			expectedError: app_error.MsgInvalidToken,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			tt.mockSetup()

// 			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.input))
// 			if !tt.skipHeader {
// 				req.Header.Set("Content-Type", "application/json")
// 			}
// 			rec := httptest.NewRecorder()

// 			h.ResendConfirmationLetter(verifyJWTManagerMock).ServeHTTP(rec, req)

// 			if tt.expectedError != "" {
// 				var errResp rest.ErrorResponse
// 				if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
// 					assert.Equal(t, tt.expectedError, errResp.Message)
// 				}
// 			}

// 			assert.Equal(t, tt.expectedCode, rec.Code)
// 			if tt.expectedCode == http.StatusOK {
// 				time.Sleep(time.Second / 2)
// 			}
// 		})
// 	}
// }
// func TestLogin(t *testing.T) {
// 	ctrl, authServiceMock, validatorMock, _, accessJWTGeneratorMock, refreshJWTGeneratorMock, h := setup(t)
// 	defer ctrl.Finish()

// 	tests := []testCase{
// 		{
// 			name: "Success",
// 			mockSetup: func() {
// 				authServiceMock.EXPECT().AuthenticateUser(gomock.Any(), gomock.Any()).Return(UserIdentity, nil)
// 				validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)
// 				accessJWTGeneratorMock.EXPECT().Generate(userDTO.ID, userDTO.Role, userDTO.Active).Return(token, nil)
// 				refreshJWTGeneratorMock.EXPECT().Generate(userDTO.ID).Return(token, nil)
// 			},
// 			input:        loginJSON,
// 			expectedCode: http.StatusOK,
// 			cookieCheck:  true,
// 		},
// 		{
// 			name:          "Bind Error",
// 			mockSetup:     func() {},
// 			input:         `invalid-json`,
// 			skipHeader:    true,
// 			expectedCode:  http.StatusBadRequest,
// 			expectedError: rest.MsgUnsupportedMedia,
// 		},
// 		{
// 			name: "Invalid Arguments",
// 			mockSetup: func() {
// 				validatorMock.EXPECT().Validate(gomock.Any()).Return(rest.NewInvalidArgumentsError(errors.New("")))
// 			},
// 			input:         loginJSON,
// 			expectedCode:  http.StatusBadRequest,
// 			expectedError: rest.MsgInvalidArguments,
// 		},
// 		{
// 			name: "Auth Error",
// 			mockSetup: func() {
// 				authServiceMock.EXPECT().AuthenticateUser(gomock.Any(), gomock.Any()).Return(service.UserIdentity{}, hasher.ErrPasswordMismatch)
// 				validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)
// 			},
// 			input:         loginJSON,
// 			expectedCode:  http.StatusUnauthorized,
// 			expectedError: auth.MsgIncorrectEmailOrPass,
// 		},
// 		{
// 			name: "Service Error",
// 			mockSetup: func() {
// 				authServiceMock.EXPECT().AuthenticateUser(gomock.Any(), gomock.Any()).Return(service.UserIdentity{}, errors.New(""))
// 				validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)
// 			},
// 			input:         loginJSON,
// 			expectedCode:  http.StatusInternalServerError,
// 			expectedError: rest.MsgInternalServerError,
// 		},
// 		{
// 			name: "Access Token Error",
// 			mockSetup: func() {
// 				validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)
// 				authServiceMock.EXPECT().AuthenticateUser(gomock.Any(), gomock.Any()).Return(UserIdentity, nil)
// 				accessJWTGeneratorMock.EXPECT().Generate(userDTO.ID, userDTO.Role, userDTO.Active).Return("", rest.NewInternalServerError(errors.New("")))
// 			},
// 			input:         loginJSON,
// 			expectedCode:  http.StatusInternalServerError,
// 			expectedError: rest.MsgInternalServerError,
// 		},
// 		{
// 			name: "Refresh Token Error",
// 			mockSetup: func() {
// 				authServiceMock.EXPECT().AuthenticateUser(gomock.Any(), gomock.Any()).Return(UserIdentity, nil)
// 				validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)
// 				accessJWTGeneratorMock.EXPECT().Generate(userDTO.ID, userDTO.Role, userDTO.Active).Return(token, nil)
// 				refreshJWTGeneratorMock.EXPECT().Generate(userDTO.ID).Return("", rest.NewInternalServerError(errors.New("")))
// 			},
// 			input:         loginJSON,
// 			expectedCode:  http.StatusInternalServerError,
// 			expectedError: rest.MsgInternalServerError,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			tt.mockSetup()

// 			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.input))
// 			if !tt.skipHeader {
// 				req.Header.Set("Content-Type", "application/json")
// 			}
// 			rec := httptest.NewRecorder()

// 			h.Login(accessJWTGeneratorMock, refreshJWTGeneratorMock).ServeHTTP(rec, req)

// 			if tt.expectedError != "" {
// 				var errResp rest.ErrorResponse
// 				if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
// 					assert.Equal(t, tt.expectedError, errResp.Message)
// 				}
// 			}

// 			assert.Equal(t, tt.expectedCode, rec.Code)

// 			if tt.cookieCheck {
// 				cookies := rec.Header().Values("Set-Cookie")
// 				assert.Contains(t, cookies[0], fmt.Sprintf("access_token=%s; HttpOnly; Secure", token))
// 				assert.Contains(t, cookies[1], fmt.Sprintf("refresh_token=%s; HttpOnly; Secure", token))
// 			}
// 		})
// 	}
// }

// func TestLogout_Success(t *testing.T) {
// 	ctrl, _, _, _, _, _, h := setup(t)
// 	defer ctrl.Finish()

// 	req := httptest.NewRequest(http.MethodPost, "/", nil)
// 	rec := httptest.NewRecorder()

// 	h.Logout(rec, req)

// 	cookies := rec.Header().Values("Set-Cookie")

// 	assert.Equal(t, http.StatusOK, rec.Code)
// 	assert.Contains(t, cookies[0], "access_token=; Expires=Thu, 01 Jan 1970 00:00:00 GMT; Max-Age=0; HttpOnly")
// 	assert.Contains(t, cookies[1], "refresh_token=; Expires=Thu, 01 Jan 1970 00:00:00 GMT; Max-Age=0; HttpOnly")
// }

// func TestRefresh(t *testing.T) {
// 	ctrl, authServiceMock, _, _, accessJWTGeneratorMock, _, h := setup(t)
// 	defer ctrl.Finish()

// 	tests := []testCase{
// 		{
// 			name: "Success",
// 			mockSetup: func() {
// 				authServiceMock.EXPECT().GetUserRoleAndStatus(gomock.Any(), gomock.Any()).Return(UserRoleAndStatus, nil)
// 				accessJWTGeneratorMock.EXPECT().Generate(userDTO.ID, userDTO.Role, userDTO.Active).Return(token, nil)
// 			},
// 			expectedCode: http.StatusOK,
// 			cookieCheck:  true,
// 		},
// 		{
// 			name: "Not Found",
// 			mockSetup: func() {
// 				authServiceMock.EXPECT().GetUserRoleAndStatus(gomock.Any(), gomock.Any()).Return(service.UserRoleAndStatus{}, pgx.ErrNoRows)
// 			},
// 			expectedCode:  http.StatusNotFound,
// 			expectedError: app_error.MsgUserNotFound,
// 		},
// 		{
// 			name: "Service Error",
// 			mockSetup: func() {
// 				authServiceMock.EXPECT().GetUserRoleAndStatus(gomock.Any(), gomock.Any()).Return(service.UserRoleAndStatus{}, errors.New(""))
// 			},
// 			expectedCode:  http.StatusInternalServerError,
// 			expectedError: rest.MsgInternalServerError,
// 		},
// 		{
// 			name:          "JwtId Not Set",
// 			mockSetup:     func() {},
// 			expectedCode:  http.StatusInternalServerError,
// 			expectedError: rest.MsgInternalServerError,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			tt.mockSetup()

// 			req := httptest.NewRequest(http.MethodPost, "/", nil)
// 			rec := httptest.NewRecorder()

// 			if tt.name != "JwtId Not Set" {
// 				ctx := ctx_util.SetJWTId(req.Context(), userDTO.ID)
// 				ctx = ctx_util.SetJWTRole(ctx, enum.UserRole(userDTO.Role))
// 				ctx = ctx_util.SetJWTActive(ctx, userDTO.Active)
// 				req = req.WithContext(ctx)
// 			}

// 			h.Refresh(accessJWTGeneratorMock).ServeHTTP(rec, req)

// 			assert.Equal(t, tt.expectedCode, rec.Code)

// 			if tt.expectedError != "" {
// 				var errResp rest.ErrorResponse
// 				if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
// 					assert.Equal(t, tt.expectedError, errResp.Message)
// 				}
// 			}

// 			if tt.expectedCode != http.StatusOK {
// 				assert.NotContains(t, rec.Header().Get("Set-Cookie"), fmt.Sprintf("access_token=%s; HttpOnly; Secure", token))
// 			} else {
// 				assert.Contains(t, rec.Header().Get("Set-Cookie"), fmt.Sprintf("access_token=%s; HttpOnly; Secure", token))
// 			}
// 		})
// 	}
// }

// func TestVerifyEmail(t *testing.T) {
// 	ctrl, authServiceMock, _, vrfJWTManager, _, _, h := setup(t)
// 	defer ctrl.Finish()

// 	tests := []testCase{
// 		{
// 			name: "Success",
// 			mockSetup: func() {
// 				vrfJWTManager.EXPECT().Verify(gomock.Any()).Return(verifyClaims, nil)
// 				authServiceMock.EXPECT().VerifyUser(gomock.Any(), gomock.Any(), gomock.Any(), int64(1)).Return(nil)
// 				authServiceMock.EXPECT().SendVerifiedLetter(verifyClaims.Email).Return(nil)
// 			},
// 			expectedCode: http.StatusOK,
// 		},
// 		{
// 			name: "Invalid Token",
// 			mockSetup: func() {
// 				vrfJWTManager.EXPECT().Verify(token).Return(verify_jwt.VerifyClaims{}, rest.NewUnauthorizedError(errors.New(""), app_error.MsgInvalidToken))
// 			},
// 			expectedCode:  http.StatusUnauthorized,
// 			expectedError: app_error.MsgInvalidToken,
// 		},
// 		{
// 			name: "Parse ID Error",
// 			mockSetup: func() {
// 				vrfJWTManager.EXPECT().Verify(token).Return(verify_jwt.VerifyClaims{ID: "invalid-id"}, nil)
// 			},
// 			expectedCode:  http.StatusInternalServerError,
// 			expectedError: rest.MsgInternalServerError,
// 		},
// 		{
// 			name: "Already Activated",
// 			mockSetup: func() {
// 				vrfJWTManager.EXPECT().Verify(token).Return(verifyClaims, nil)
// 				authServiceMock.EXPECT().VerifyUser(gomock.Any(), gomock.Any(), gomock.Any(), int64(1)).Return(redis.TxFailedErr)
// 			},
// 			expectedCode:  http.StatusConflict,
// 			expectedError: auth.MsgTokenAlreadyUsed,
// 		},
// 		{
// 			name: "Not Found",
// 			mockSetup: func() {
// 				vrfJWTManager.EXPECT().Verify(token).Return(verifyClaims, nil)
// 				authServiceMock.EXPECT().VerifyUser(gomock.Any(), gomock.Any(), gomock.Any(), int64(1)).Return(pg_error.ErrNotFound)
// 			},
// 			expectedCode:  http.StatusNotFound,
// 			expectedError: app_error.MsgUserNotFound,
// 		},
// 		{
// 			name: "Service Error",
// 			mockSetup: func() {
// 				vrfJWTManager.EXPECT().Verify(token).Return(verifyClaims, nil)
// 				authServiceMock.EXPECT().VerifyUser(gomock.Any(), gomock.Any(), gomock.Any(), int64(1)).Return(errors.New(""))
// 			},
// 			expectedCode:  http.StatusInternalServerError,
// 			expectedError: rest.MsgInternalServerError,
// 		},
// 		{
// 			name: "Mail Error",
// 			mockSetup: func() {
// 				vrfJWTManager.EXPECT().Verify(gomock.Any()).Return(verifyClaims, nil)
// 				authServiceMock.EXPECT().VerifyUser(gomock.Any(), gomock.Any(), gomock.Any(), int64(1)).Return(nil)
// 				authServiceMock.EXPECT().SendVerifiedLetter(verifyClaims.Email).Return(errors.New(""))
// 			},
// 			expectedCode: http.StatusOK,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			tt.mockSetup()

// 			q := make(url.Values)
// 			q.Set("token", token)
// 			req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
// 			rec := httptest.NewRecorder()

// 			h.VerifyEmail(vrfJWTManager).ServeHTTP(rec, req)

// 			assert.Equal(t, tt.expectedCode, rec.Code)

// 			if tt.expectedError != "" {
// 				var errResp rest.ErrorResponse
// 				if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
// 					assert.Equal(t, tt.expectedError, errResp.Message)
// 				}
// 			}

// 			if tt.expectedCode == http.StatusOK {
// 				time.Sleep(time.Second / 2)
// 			}
// 		})
// 	}
// }
