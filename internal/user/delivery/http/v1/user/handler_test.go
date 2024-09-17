package user_test

import (
	"bytes"
	"context"
	"errors"
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

	var err *rest.ErrorResponse
	if (assert.ErrorAs(t, h.FindUserById(c), &err)) {
		assert.ErrorIs(t, pgx.ErrNoRows, err.Cause)
		assert.Equal(t, rest.MsgUserNotFound, err.Message)
		assert.Equal(t, http.StatusNotFound, err.Status)
	}
}

func TestFindUserByIdOnRepositoryError(t *testing.T) {
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
 
	var err *rest.ErrorResponse
	if (assert.ErrorAs(t, h.FindUserById(c), &err)) {
		assert.Equal(t, rest.MsgInternalServerError, err.Message)
		assert.Equal(t, http.StatusInternalServerError, err.Status)
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

	var err *rest.ErrorResponse
	if (assert.ErrorAs(t, h.FindUserById(c), &err)) {
		assert.ErrorIs(t, ctxutil.ErrParamIdNotImplemented, err.Cause)
		assert.Equal(t, rest.MsgInternalServerError, err.Message)
		assert.Equal(t, http.StatusInternalServerError, err.Status)
	}
}
