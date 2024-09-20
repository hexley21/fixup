package user_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/hexley21/fixup/internal/common/rest"
	"github.com/hexley21/fixup/internal/common/util/ctxutil"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/user"
	"github.com/hexley21/fixup/internal/user/enum"
	mock_service "github.com/hexley21/fixup/internal/user/service/mock"
	"github.com/hexley21/fixup/pkg/hasher"
	mock_validator "github.com/hexley21/fixup/pkg/validator/mock"
	"github.com/jackc/pgx/v5"
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

	updateUserJSON     = `{"email": "larry@page.com","first_name": "Larry","last_name": "Page","phone_number": "995112233"}`
	changePasswordJSON = `{"old_password": "larrypage123", "new_password": "pagelarry321"}`

	fileContent = []byte("fake file content")
)

func createMultipartFormData(t *testing.T, fieldName, fileName string, fileContent []byte) (*bytes.Buffer, string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		t.Fatal(err)
	}

	_, err = part.Write(fileContent)
	if err != nil {
		t.Fatal(err)
	}

	err = writer.Close()
	if err != nil {
		t.Fatal(err)
	}

	return body, writer.FormDataContentType()
}

func TestFindUserById_Success(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mock_service.NewMockUserService(ctrl)
	mockUserService.EXPECT().FindUserById(ctx, int64(1)).Return(userDto, nil)

	h := user.NewUserHandler(mockUserService)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetPath("/:id")
	ctxutil.SetParamId(c, userDto.ID)

	assert.NoError(t, h.FindUserById(c))
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestFindUserById_NotFound(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mock_service.NewMockUserService(ctrl)
	mockUserService.EXPECT().FindUserById(ctx, int64(1)).Return(dto.User{}, pgx.ErrNoRows)

	h := user.NewUserHandler(mockUserService)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetPath("/:id")
	ctxutil.SetParamId(c, userDto.ID)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.FindUserById(c), &errResp) {
		assert.ErrorIs(t, pgx.ErrNoRows, errResp.Cause)
		assert.Equal(t, rest.MsgUserNotFound, errResp.Message)
		assert.Equal(t, http.StatusNotFound, errResp.Status)
	}
}

func TestFindUserById_ServiceError(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mock_service.NewMockUserService(ctrl)
	mockUserService.EXPECT().FindUserById(ctx, int64(1)).Return(dto.User{}, errors.New(""))

	h := user.NewUserHandler(mockUserService)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetPath("/:id")
	ctxutil.SetParamId(c, userDto.ID)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.FindUserById(c), &errResp) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}

func TestFindUserById_IdParamNotImplemented(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	h := user.NewUserHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetPath("/:id")

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.FindUserById(c), &errResp) {
		assert.ErrorIs(t, ctxutil.ErrParamIdNotImplemented.Cause, errResp.Cause)
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}

func TestUploadProfilePicture_Success(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mock_service.NewMockUserService(ctrl)
	mockUserService.EXPECT().SetProfilePicture(ctx, int64(1), gomock.Any(), "", gomock.Any(), gomock.Any()).Return(nil)

	h := user.NewUserHandler(mockUserService)

	body, contentType := createMultipartFormData(t, "image", "test.jpg", fileContent)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set(echo.HeaderContentType, contentType)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetPath("/:id")
	ctxutil.SetParamId(c, "1")

	assert.NoError(t, h.UploadProfilePicture(c))
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestUploadProfilePicture_MissingHeaders(t *testing.T) {
	h := user.NewUserHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	
	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetPath("/:id")
	ctxutil.SetParamId(c, "1")

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.UploadProfilePicture(c), &errResp) {
		assert.ErrorIs(t, errResp.Cause, http.ErrNotMultipart)
		assert.Equal(t, rest.MsgFileReadError, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, errResp.Status)
	}
}

func TestUploadProfilePicture_MissingFile(t *testing.T) {
	h := user.NewUserHandler(nil)

	_, contentType := createMultipartFormData(t, "image", "test.jpg", fileContent)
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(echo.HeaderContentType, contentType)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetPath("/:id")
	ctxutil.SetParamId(c, "1")


	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.UploadProfilePicture(c), &errResp) {
		assert.ErrorIs(t, errResp.Cause, io.EOF)
		assert.Equal(t, rest.MsgFileReadError, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, errResp.Status)
	}
}

func TestUploadProfilePicture_WrongFormData(t *testing.T) {
	h := user.NewUserHandler(nil)


	body, contentType := createMultipartFormData(t, "img", "test.jpg", fileContent)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set(echo.HeaderContentType, contentType)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetPath("/:id")
	ctxutil.SetParamId(c, "1")

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.UploadProfilePicture(c), &errResp) {
		assert.NoError(t, errResp.Cause)
		assert.Equal(t, rest.MsgNoFile, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, errResp.Status)
	}
}

func TestUploadProfilePicture_NotFound(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mock_service.NewMockUserService(ctrl)
	mockUserService.EXPECT().SetProfilePicture(ctx, int64(1), gomock.Any(), "", gomock.Any(), gomock.Any()).Return(pgx.ErrNoRows)

	h := user.NewUserHandler(mockUserService)

	body, contentType := createMultipartFormData(t, "image", "test.jpg", fileContent)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set(echo.HeaderContentType, contentType)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetPath("/:id")
	ctxutil.SetParamId(c, "1")

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.UploadProfilePicture(c), &errResp) {
		assert.ErrorIs(t, errResp.Cause, pgx.ErrNoRows)
		assert.Equal(t, rest.MsgUserNotFound, errResp.Message)
		assert.Equal(t, http.StatusNotFound, errResp.Status)
	}
}

func TestUploadProfilePicture_ServiceError(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mock_service.NewMockUserService(ctrl)
	mockUserService.EXPECT().SetProfilePicture(ctx, int64(1), gomock.Any(), "", gomock.Any(), gomock.Any()).Return(errors.New(""))

	h := user.NewUserHandler(mockUserService)

	body, contentType := createMultipartFormData(t, "image", "test.jpg", fileContent)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set(echo.HeaderContentType, contentType)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetPath("/:id")
	ctxutil.SetParamId(c, "1")

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.UploadProfilePicture(c), &errResp) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}

func TestUploadProfilePicture_IdParamNotImplemented(t *testing.T) {
	h := user.NewUserHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetPath("/:id")

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.UploadProfilePicture(c), &errResp) {
		assert.ErrorIs(t, ctxutil.ErrParamIdNotImplemented.Cause, errResp.Cause)
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}

func TestUpdateUserData_Success(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mock_service.NewMockUserService(ctrl)
	mockValidator := mock_validator.NewMockValidator(ctrl)

	mockUserService.EXPECT().UpdateUserDataById(ctx, int64(1), gomock.Any()).Return(userDto, nil)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	h := user.NewUserHandler(mockUserService)
	
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(updateUserJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	
	e := echo.New()
	e.Validator = mockValidator

	c := e.NewContext(req, rec)
	c.SetPath("/:id")
	ctxutil.SetParamId(c, "1")

	assert.NoError(t, h.UpdateUserData(c))
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestUpdateUserDataWith_BindError(t *testing.T) {
	h := user.NewUserHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(updateUserJSON))
	rec := httptest.NewRecorder()
	
	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetPath("/:id")
	ctxutil.SetParamId(c, "1")

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.UpdateUserData(c), &errResp) {
		assert.Equal(t, rest.MsgInvalidArguments, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, errResp.Status)
	}
}

func TestUpdateUserData_InvalidValues(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockValidator := mock_validator.NewMockValidator(ctrl)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(errors.New(""))
	
	h := user.NewUserHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(updateUserJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	
	e := echo.New()
	e.Validator = mockValidator

	c := e.NewContext(req, rec)
	c.SetPath("/:id")
	ctxutil.SetParamId(c, "1")

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.UpdateUserData(c), &errResp) {
		assert.Equal(t, rest.MsgInvalidArguments, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, errResp.Status)
	}
}

func TestUpdateUserData_NotFound(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mock_service.NewMockUserService(ctrl)
	mockValidator := mock_validator.NewMockValidator(ctrl)

	mockUserService.EXPECT().UpdateUserDataById(ctx, int64(1), gomock.Any()).Return(dto.User{}, pgx.ErrNoRows)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	h := user.NewUserHandler(mockUserService)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(updateUserJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e := echo.New()
	e.Validator = mockValidator

	c := e.NewContext(req, rec)
	c.SetPath("/:id")
	ctxutil.SetParamId(c, "1")

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.UpdateUserData(c), &errResp) {
		assert.ErrorIs(t, errResp.Cause, pgx.ErrNoRows)
		assert.Equal(t, rest.MsgUserNotFound, errResp.Message)
		assert.Equal(t, http.StatusNotFound, errResp.Status)
	}
}

func TestUpdateUserData_ServiceError(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mock_service.NewMockUserService(ctrl)
	mockValidator := mock_validator.NewMockValidator(ctrl)

	mockUserService.EXPECT().UpdateUserDataById(ctx, int64(1), gomock.Any()).Return(dto.User{}, errors.New(""))
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	h := user.NewUserHandler(mockUserService)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(updateUserJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e := echo.New()
	e.Validator = mockValidator

	c := e.NewContext(req, rec)
	c.SetPath("/:id")
	ctxutil.SetParamId(c, "1")

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.UpdateUserData(c), &errResp) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}

func TestUpdateUserData_IdParamNotImplemented(t *testing.T) {
	h := user.NewUserHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetPath("/:id")

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.UpdateUserData(c), &errResp) {
		assert.ErrorIs(t, ctxutil.ErrParamIdNotImplemented.Cause, errResp.Cause)
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}

func TestDeleteUser_Success(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mock_service.NewMockUserService(ctrl)
	mockUserService.EXPECT().DeleteUserById(ctx, int64(1)).Return(nil)

	h := user.NewUserHandler(mockUserService)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetPath("/:id")
	ctxutil.SetParamId(c, "1")

	assert.NoError(t, h.DeleteUser(c))
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestDelete_NotFound(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mock_service.NewMockUserService(ctrl)
	mockUserService.EXPECT().DeleteUserById(ctx, int64(1)).Return(pgx.ErrNoRows)

	h := user.NewUserHandler(mockUserService)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetPath("/:id")
	ctxutil.SetParamId(c, "1")


	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.DeleteUser(c), &errResp) {
		assert.ErrorIs(t, pgx.ErrNoRows, errResp.Cause)
		assert.Equal(t, rest.MsgUserNotFound, errResp.Message)
		assert.Equal(t, http.StatusNotFound, errResp.Status)
	}
}

func TestDeleteUser_ServiceError(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mock_service.NewMockUserService(ctrl)
	mockUserService.EXPECT().DeleteUserById(ctx, int64(1)).Return(errors.New(""))

	h := user.NewUserHandler(mockUserService)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetPath("/:id")
	ctxutil.SetParamId(c, "1")

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.DeleteUser(c), &errResp) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}

func TestTestDeleteUser_IdParamNotImplemented(t *testing.T) {
	h := user.NewUserHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetPath("/:id")

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.DeleteUser(c), &errResp) {
		assert.ErrorIs(t, ctxutil.ErrParamIdNotImplemented.Cause, errResp.Cause)
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}

func TestChangePassword_Success(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mock_service.NewMockUserService(ctrl)
	mockValidator := mock_validator.NewMockValidator(ctrl)

	mockUserService.EXPECT().ChangePassword(ctx, int64(1), gomock.Any()).Return(nil)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	h := user.NewUserHandler(mockUserService)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	e := echo.New()
	e.Validator = mockValidator
	c := e.NewContext(req, rec)
	ctxutil.SetJwtId(c, "1")

	assert.NoError(t, h.ChangePassword(c))
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestChangePassword_BindError(t *testing.T) {
	h := user.NewUserHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(changePasswordJSON))
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)
	ctxutil.SetJwtId(c, "1")


	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.ChangePassword(c), &errResp) {
		assert.Equal(t, rest.MsgInvalidArguments, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, errResp.Status)
	}
}

func TestChangePassword_InvalidValues(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockValidator := mock_validator.NewMockValidator(ctrl)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(errors.New(""))

	h := user.NewUserHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(changePasswordJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e := echo.New()
	e.Validator = mockValidator
	c := e.NewContext(req, rec)
	ctxutil.SetJwtId(c, "1")

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.ChangePassword(c), &errResp) {
		assert.Equal(t, rest.MsgInvalidArguments, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, errResp.Status)
	}
}

func TestChangePassword_NotFound(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mock_service.NewMockUserService(ctrl)
	mockValidator := mock_validator.NewMockValidator(ctrl)

	mockUserService.EXPECT().ChangePassword(ctx, int64(1), gomock.Any()).Return(pgx.ErrNoRows)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	h := user.NewUserHandler(mockUserService)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	e := echo.New()
	e.Validator = mockValidator
	c := e.NewContext(req, rec)
	ctxutil.SetJwtId(c, "1")

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.ChangePassword(c), &errResp) {
		assert.Equal(t, pgx.ErrNoRows, errResp.Cause)
		assert.Equal(t, rest.MsgUserNotFound, errResp.Message)
		assert.Equal(t, http.StatusNotFound, errResp.Status)
	}
}

func TestChangePassword_IncorrectPassword(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mock_service.NewMockUserService(ctrl)
	mockValidator := mock_validator.NewMockValidator(ctrl)

	mockUserService.EXPECT().ChangePassword(ctx, int64(1), gomock.Any()).Return(hasher.ErrPasswordMismatch)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	h := user.NewUserHandler(mockUserService)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	
	e := echo.New()
	e.Validator = mockValidator
	c := e.NewContext(req, rec)
	ctxutil.SetJwtId(c, "1")

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.ChangePassword(c), &errResp) {
		assert.Equal(t, hasher.ErrPasswordMismatch, errResp.Cause)
		assert.Equal(t, rest.MsgIncorrectPassword, errResp.Message)
		assert.Equal(t, http.StatusUnauthorized, errResp.Status)
	}
}

func TestChangePassword_ServiceError(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mock_service.NewMockUserService(ctrl)
	mockValidator := mock_validator.NewMockValidator(ctrl)

	mockUserService.EXPECT().ChangePassword(ctx, int64(1), gomock.Any()).Return(errors.New(""))
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	h := user.NewUserHandler(mockUserService)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	
	e := echo.New()
	e.Validator = mockValidator
	c := e.NewContext(req, rec)
	ctxutil.SetJwtId(c, "1")

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.ChangePassword(c), &errResp) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}

func TestChangePassword_JwtNotImplement(t *testing.T) {
	h := user.NewUserHandler(nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.ChangePassword(c), &errResp) {
		assert.ErrorIs(t, ctxutil.ErrJwtNotImplemented.Cause, errResp.Cause)
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}
