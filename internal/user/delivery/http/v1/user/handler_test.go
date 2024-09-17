package user_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hexley21/fixup/internal/common/rest"
	"github.com/hexley21/fixup/internal/common/util/ctxutil"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/user"
	"github.com/hexley21/fixup/internal/user/enum"
	mock_service "github.com/hexley21/fixup/internal/user/service/mock"
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

func TestFindUserById(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mock_service.NewMockUserService(ctrl)
	mockUserService.EXPECT().FindUserById(ctx, int64(1)).Return(userDto, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/:id")
	ctxutil.SetParamId(c, userDto.ID)

	h := user.NewUserHandler(mockUserService)

	assert.NoError(t, h.FindUserById(c))
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestFindUserByIdOnNonexistentUser(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mock_service.NewMockUserService(ctrl)
	mockUserService.EXPECT().FindUserById(ctx, int64(1)).Return(dto.User{}, pgx.ErrNoRows)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/:id")
	ctxutil.SetParamId(c, userDto.ID)

	h := user.NewUserHandler(mockUserService)

	var errResp *rest.ErrorResponse
	if (assert.ErrorAs(t, h.FindUserById(c), &errResp)) {
		assert.ErrorIs(t, pgx.ErrNoRows, errResp.Cause)
		assert.Equal(t, rest.MsgUserNotFound, errResp.Message)
		assert.Equal(t, http.StatusNotFound, errResp.Status)
	}
}

func TestFindUserByIdWithServiceError(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/:id")
	ctxutil.SetParamId(c, userDto.ID)

	mockUserService := mock_service.NewMockUserService(ctrl)
	mockUserService.EXPECT().FindUserById(ctx, int64(1)).Return(dto.User{}, errors.New(""))

	h := user.NewUserHandler(mockUserService)
 
	var errResp *rest.ErrorResponse
	if (assert.ErrorAs(t, h.FindUserById(c), &errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}

func TestFindUserByIdOnIdParamNotImplemented(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/:id")

	h := user.NewUserHandler(nil)

	var errResp *rest.ErrorResponse
	if (assert.ErrorAs(t, h.FindUserById(c), &errResp)) {
		assert.ErrorIs(t, ctxutil.ErrParamIdNotImplemented, errResp.Cause)
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}

func TestUploadProfilePicture(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mock_service.NewMockUserService(ctrl)
	mockUserService.EXPECT().SetProfilePicture(ctx, int64(1), gomock.Any(), "", gomock.Any(), gomock.Any()).Return(nil)

	body, contentType := createMultipartFormData(t, "image", "test.jpg", fileContent)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", body)	
	req.Header.Set(echo.HeaderContentType, contentType)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/:id")
	ctxutil.SetParamId(c, "1")

	h := user.NewUserHandler(mockUserService)

	assert.NoError(t, h.UploadProfilePicture(c))
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestUploadProfilePictureWithoutMultipart(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)	
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/:id")
	ctxutil.SetParamId(c, "1")

	h := user.NewUserHandler(nil)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.UploadProfilePicture(c), &errResp) {
		assert.ErrorIs(t, errResp.Cause, http.ErrNotMultipart)
		assert.Equal(t, rest.MsgFileReadError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}

func TestUploadProfilePictureWithoutFile(t *testing.T) {
	_, contentType := createMultipartFormData(t, "image", "test.jpg", fileContent)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)	
	req.Header.Set(echo.HeaderContentType, contentType)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/:id")
	ctxutil.SetParamId(c, "1")

	h := user.NewUserHandler(nil)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.UploadProfilePicture(c), &errResp) {
		assert.ErrorIs(t, errResp.Cause, io.EOF)
		assert.Equal(t, rest.MsgFileReadError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}
func TestUploadProfilePictureWithWrongField(t *testing.T) {
	body, contentType := createMultipartFormData(t, "img", "test.jpg", fileContent)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", body)	
	req.Header.Set(echo.HeaderContentType, contentType)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/:id")
	ctxutil.SetParamId(c, "1")

	h := user.NewUserHandler(nil)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.UploadProfilePicture(c), &errResp) {
		assert.NoError(t, errResp.Cause)
		assert.Equal(t, rest.MsgNoFile, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, errResp.Status)
	}
}

func TestUploadProfilePictureToNonexistentUser(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mock_service.NewMockUserService(ctrl)
	mockUserService.EXPECT().SetProfilePicture(ctx, int64(1), gomock.Any(), "", gomock.Any(), gomock.Any()).Return(pgx.ErrNoRows)

	body, contentType := createMultipartFormData(t, "image", "test.jpg", fileContent)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", body)	
	req.Header.Set(echo.HeaderContentType, contentType)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/:id")
	ctxutil.SetParamId(c, "1")

	h := user.NewUserHandler(mockUserService)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.UploadProfilePicture(c), &errResp) {
		assert.ErrorIs(t, errResp.Cause, pgx.ErrNoRows)
		assert.Equal(t, rest.MsgUserNotFound, errResp.Message)
		assert.Equal(t, http.StatusNotFound, errResp.Status)
	}
}

func TestUploadProfilePictureWithServiceError(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mock_service.NewMockUserService(ctrl)
	mockUserService.EXPECT().SetProfilePicture(ctx, int64(1), gomock.Any(), "", gomock.Any(), gomock.Any()).Return(errors.New(""))

	body, contentType := createMultipartFormData(t, "image", "test.jpg", fileContent)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", body)	
	req.Header.Set(echo.HeaderContentType, contentType)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/:id")
	ctxutil.SetParamId(c, "1")

	h := user.NewUserHandler(mockUserService)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, h.UploadProfilePicture(c), &errResp) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}
