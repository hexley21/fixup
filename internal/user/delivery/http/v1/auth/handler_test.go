package auth_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/hexley21/fixup/internal/common/rest"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/auth"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/user/enum"
	"github.com/hexley21/fixup/internal/user/service"
	mock_service "github.com/hexley21/fixup/internal/user/service/mock"
	mock_verifier "github.com/hexley21/fixup/internal/user/service/verifier/mock"
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

	userConfirmationDetails = dto.UserConfirmationDetails{
        ID:        "1",
		UserStatus: true,
        Firstname: "Larry",
    }

	registerCustomerJSON = `{"email": "larry@page.com", "phone_number": "995111222333", "first_name": "Larry", "last_name": "Page", "password": "larrypage123"}`
	registerProviderJSON = `{"email": "larry@page.com", "phone_number": "995111222333", "first_name": "Larry", "last_name": "Page", "password": "larrypage123", "personal_id_number": "1234567890"}`
	emailJSON = `{"email": "larry@page.com"}`

	verificationToken = "Ehx0DNg86zL6QCB8gMZxzkm0fPt3ObwhQzKAu22bnVYZvVe84GAAh8jFp5Cf47R5Ync"
)

func setup(t *testing.T) (*gomock.Controller, *mock_service.MockAuthService, *mock_validator.MockValidator ,*mock_verifier.MockJwtGenerator, *auth.AuthHandler, *echo.Echo) {
    ctrl := gomock.NewController(t)
    mockAuthService := mock_service.NewMockAuthService(ctrl)
	mockValidator := mock_validator.NewMockValidator(ctrl)
	mockVerifierGenerator := mock_verifier.NewMockJwtGenerator(ctrl)

	h := auth.NewAuthHandler(mockAuthService)

	e := echo.New()
	e.Validator = mockValidator

    return ctrl, mockAuthService, mockValidator, mockVerifierGenerator, h, e
}

func TestRegisterCustomer_Success(t *testing.T) {
	ctx := context.Background()

	ctrl, mockAuthService, mockValidator, mockVerifierGenerator, h, e := setup(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().RegisterCustomer(ctx, gomock.Any()).Return(userDto, nil)
	mockAuthService.EXPECT().SendConfirmationLetter(verificationToken, userDto.Email, userDto.FirstName).Return(nil)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockVerifierGenerator.EXPECT().GenerateJWT(userDto.ID, userDto.Email).Return(verificationToken, nil)

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
	ctrl, _, mockValidator, _, h, e := setup(t)
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

	ctrl, mockAuthService, mockValidator, _, h, e := setup(t)
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

	ctrl, mockAuthService, mockValidator, _, h, e := setup(t)
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
    ctrl, mockAuthService, mockValidator, mockVerifierGenerator, h, e := setup(t)
    defer ctrl.Finish()

    mockAuthService.EXPECT().RegisterProvider(ctx, gomock.Any()).Return(userDto, nil)
	mockAuthService.EXPECT().SendConfirmationLetter(verificationToken, userDto.Email, userDto.FirstName).Return(nil)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockVerifierGenerator.EXPECT().GenerateJWT(userDto.ID, userDto.Email).Return(verificationToken, nil)

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
	ctrl, _, mockValidator, _, h, e := setup(t)
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

	ctrl, mockAuthService, mockValidator, _, h, e := setup(t)
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

	ctrl, mockAuthService, mockValidator, _, h, e := setup(t)
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
    ctrl, mockAuthService, mockValidator, mockVerifierGenerator, h, e := setup(t)
    defer ctrl.Finish()

    mockAuthService.EXPECT().GetUserConfirmationDetails(ctx, gomock.Any()).Return(userConfirmationDetails, nil)
    mockAuthService.EXPECT().SendConfirmationLetter(verificationToken, userDto.Email, userDto.FirstName).Return(nil)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockVerifierGenerator.EXPECT().GenerateJWT(userDto.ID, userDto.Email).Return(verificationToken, nil)

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
	ctrl, _, mockValidator, _, h, e := setup(t)
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

	ctrl, mockAuthService, mockValidator, _, h, e := setup(t)
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

	ctrl, mockAuthService, mockValidator, _, h, e := setup(t)
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

	ctrl, mockAuthService, mockValidator, _, h, e := setup(t)
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
    ctrl, mockAuthService, mockValidator, mockVerifierGenerator, h, e := setup(t)
    defer ctrl.Finish()

    mockAuthService.EXPECT().GetUserConfirmationDetails(ctx, gomock.Any()).Return(userConfirmationDetails, nil)
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
