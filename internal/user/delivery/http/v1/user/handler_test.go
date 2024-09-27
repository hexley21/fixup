package user_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/hexley21/fixup/internal/common/app_error"
	"github.com/hexley21/fixup/internal/common/util/ctx_util"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/user"
	"github.com/hexley21/fixup/internal/user/enum"
	mock_service "github.com/hexley21/fixup/internal/user/service/mock"
	"github.com/hexley21/fixup/pkg/hasher"
	"github.com/hexley21/fixup/pkg/http/binder/std_binder"
	"github.com/hexley21/fixup/pkg/http/json/std_json"
	"github.com/hexley21/fixup/pkg/http/rest"
	"github.com/hexley21/fixup/pkg/http/writer/json_writer"
	"github.com/hexley21/fixup/pkg/infra/postgres/pg_error"
	"github.com/hexley21/fixup/pkg/logger/std_logger"
	mock_validator "github.com/hexley21/fixup/pkg/validator/mock"
	"github.com/jackc/pgx/v5"
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

func setup(t *testing.T) (*gomock.Controller, *mock_service.MockUserService, *mock_validator.MockValidator, *user.HandlerFactory) {
	ctrl := gomock.NewController(t)
	mockUserService := mock_service.NewMockUserService(ctrl)
	mockValidator := mock_validator.NewMockValidator(ctrl)

	logger := std_logger.New()
	jsonManager := std_json.New()

	f := user.NewFactory(
		logger,
		std_binder.New(jsonManager),
		mockValidator,
		json_writer.New(logger, jsonManager),
		mockUserService,
	)

	return ctrl, mockUserService, mockValidator, f
}

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

	ctrl, mockUserService, _, f := setup(t)
	defer ctrl.Finish()

	mockUserService.EXPECT().FindUserById(ctx, int64(1)).Return(userDto, nil)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	f.FindUserById(rec, req.WithContext(ctx_util.SetParamId(req.Context(), 1)))

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestFindUserById_NotFound(t *testing.T) {
	ctx := context.Background()

	ctrl, mockUserService, _, f := setup(t)
	defer ctrl.Finish()

	mockUserService.EXPECT().FindUserById(ctx, int64(1)).Return(dto.User{}, pg_error.ErrNotFound)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	f.FindUserById(rec, req.WithContext(ctx_util.SetParamId(req.Context(), 1)))
	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, app_error.MsgUserNotFound, errResp.Message)
		assert.Equal(t, http.StatusNotFound, errResp.Status)
	}
}

func TestFindUserById_ServiceError(t *testing.T) {
	ctx := context.Background()

	ctrl, mockUserService, _, f := setup(t)
	defer ctrl.Finish()

	mockUserService.EXPECT().FindUserById(ctx, int64(1)).Return(dto.User{}, rest.NewInternalServerError(errors.New("")))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	f.FindUserById(rec, req.WithContext(ctx_util.SetParamId(req.Context(), 1)))

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}

func TestFindUserById_IdParamNotSet(t *testing.T) {
	ctrl, _, _, f := setup(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	f.FindUserById(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}

func TestUploadProfilePicture_Success(t *testing.T) {
	ctx := context.Background()

	ctrl, mockUserService, _, f := setup(t)
	defer ctrl.Finish()

	mockUserService.EXPECT().SetProfilePicture(ctx, int64(1), gomock.Any(), "", gomock.Any(), gomock.Any()).Return(nil)

	body, contentType := createMultipartFormData(t, "image", "test.jpg", fileContent)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	f.UploadProfilePicture(rec, req.WithContext(ctx_util.SetParamId(req.Context(), 1)))

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestUploadProfilePicture_MissingHeaders(t *testing.T) {
	ctrl, _, _, f := setup(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	f.UploadProfilePicture(rec, req.WithContext(ctx_util.SetParamId(req.Context(), 1)))

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgFileReadError, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, errResp.Status)
	}
}

func TestUploadProfilePicture_MissingFile(t *testing.T) {
	ctrl, _, _, f := setup(t)
	defer ctrl.Finish()

	_, contentType := createMultipartFormData(t, "image", "test.jpg", fileContent)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	f.UploadProfilePicture(rec, req.WithContext(ctx_util.SetParamId(req.Context(), 1)))

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgFileReadError, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, errResp.Status)
	}
}

func TestUploadProfilePicture_WrongFormData(t *testing.T) {
	ctrl, _, _, f := setup(t)
	defer ctrl.Finish()

	body, contentType := createMultipartFormData(t, "img", "test.jpg", fileContent)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	f.UploadProfilePicture(rec, req.WithContext(ctx_util.SetParamId(req.Context(), 1)))

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgNoFile, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, errResp.Status)
	}
}

func TestUploadProfilePicture_NotFound(t *testing.T) {
	ctx := context.Background()

	ctrl, mockUserService, _, f := setup(t)
	defer ctrl.Finish()

	mockUserService.EXPECT().SetProfilePicture(ctx, int64(1), gomock.Any(), "", gomock.Any(), gomock.Any()).Return(pg_error.ErrNotFound)

	body, contentType := createMultipartFormData(t, "image", "test.jpg", fileContent)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()
	
	f.UploadProfilePicture(rec, req.WithContext(ctx_util.SetParamId(req.Context(), 1)))

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, app_error.MsgUserNotFound, errResp.Message)
		assert.Equal(t, http.StatusNotFound, errResp.Status)
	}
}

func TestUploadProfilePicture_ServiceError(t *testing.T) {
	ctx := context.Background()

	ctrl, mockUserService, _, f := setup(t)
	defer ctrl.Finish()

	mockUserService.EXPECT().SetProfilePicture(ctx, int64(1), gomock.Any(), "", gomock.Any(), gomock.Any()).Return(errors.New(""))

	body, contentType := createMultipartFormData(t, "image", "test.jpg", fileContent)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	
	f.UploadProfilePicture(rec, req.WithContext(ctx_util.SetParamId(req.Context(), 1)))

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}

func TestUploadProfilePicture_IdParamNotSet(t *testing.T) {
	ctrl, _, _, f := setup(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	f.UploadProfilePicture(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}

func TestUpdateUserData_Success(t *testing.T) {
	ctx := context.Background()

	ctrl, mockUserService, mockValidator, f := setup(t)
	defer ctrl.Finish()

	mockUserService.EXPECT().UpdateUserDataById(ctx, int64(1), gomock.Any()).Return(userDto, nil)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(updateUserJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	f.UpdateUserData(rec, req.WithContext(ctx_util.SetParamId(req.Context(), 1)))

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestUpdateUserDataWith_BindError(t *testing.T) {
	ctrl, _, _, f := setup(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(updateUserJSON))
	rec := httptest.NewRecorder()
	
	f.UpdateUserData(rec, req.WithContext(ctx_util.SetParamId(req.Context(), 1)))

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgUnsupportedMedia, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, errResp.Status)
	}
}

func TestUpdateUserData_InvalidValues(t *testing.T) {
	ctrl, _, mockValidator, f := setup(t)
	defer ctrl.Finish()

	mockValidator.EXPECT().Validate(gomock.Any()).Return(rest.NewInvalidArgumentsError(errors.New("")))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(updateUserJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	f.UpdateUserData(rec, req.WithContext(ctx_util.SetParamId(req.Context(), 1)))

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInvalidArguments, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, errResp.Status)
	}
}

func TestUpdateUserData_NotFound(t *testing.T) {
	ctx := context.Background()

	ctrl, mockUserService, mockValidator, f := setup(t)
	defer ctrl.Finish()

	mockUserService.EXPECT().UpdateUserDataById(ctx, int64(1), gomock.Any()).Return(dto.User{}, pg_error.ErrNotFound)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(updateUserJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	f.UpdateUserData(rec, req.WithContext(ctx_util.SetParamId(req.Context(), 1)))

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, app_error.MsgUserNotFound, errResp.Message)
		assert.Equal(t, http.StatusNotFound, errResp.Status)
	}
}

func TestUpdateUserData_NotChanges(t *testing.T) {
	ctx := context.Background()

	ctrl, mockUserService, mockValidator, f := setup(t)
	defer ctrl.Finish()

	mockUserService.EXPECT().UpdateUserDataById(ctx, int64(1), gomock.Any()).Return(dto.User{}, pgx.ErrNoRows)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(updateUserJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	f.UpdateUserData(rec, req.WithContext(ctx_util.SetParamId(req.Context(), 1)))

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, app_error.MsgNoChanges, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, errResp.Status)
	}
}

func TestUpdateUserData_ServiceError(t *testing.T) {
	ctx := context.Background()

	ctrl, mockUserService, mockValidator, f := setup(t)
	defer ctrl.Finish()

	mockUserService.EXPECT().UpdateUserDataById(ctx, int64(1), gomock.Any()).Return(dto.User{}, errors.New(""))
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(updateUserJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	f.UpdateUserData(rec, req.WithContext(ctx_util.SetParamId(req.Context(), 1)))

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}

func TestUpdateUserData_IdParamNotSet(t *testing.T) {
	ctrl, _, _, f := setup(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	f.UpdateUserData(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}

func TestDeleteUser_Success(t *testing.T) {
	ctx := context.Background()

	ctrl, mockUserService, _, f := setup(t)
	defer ctrl.Finish()

	mockUserService.EXPECT().DeleteUserById(ctx, int64(1)).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	f.DeleteUser(rec, req.WithContext(ctx_util.SetParamId(req.Context(), 1)))

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestDelete_NotFound(t *testing.T) {
	ctx := context.Background()

	ctrl, mockUserService, _, f := setup(t)
	defer ctrl.Finish()

	mockUserService.EXPECT().DeleteUserById(ctx, int64(1)).Return(pg_error.ErrNotFound)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	f.DeleteUser(rec, req.WithContext(ctx_util.SetParamId(req.Context(), 1)))

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, app_error.MsgUserNotFound, errResp.Message)
		assert.Equal(t, http.StatusNotFound, errResp.Status)
	}
}

func TestDeleteUser_ServiceError(t *testing.T) {
	ctx := context.Background()

	ctrl, mockUserService, _, f := setup(t)
	defer ctrl.Finish()

	mockUserService.EXPECT().DeleteUserById(ctx, int64(1)).Return(errors.New(""))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	f.DeleteUser(rec, req.WithContext(ctx_util.SetParamId(req.Context(), 1)))

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}

func TestTestDeleteUser_IdParamNotSet(t *testing.T) {
	ctrl, _, _, f := setup(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	f.DeleteUser(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}

func TestChangePassword_Success(t *testing.T) {
	ctx := context.Background()

	ctrl, mockUserService, mockValidator, f := setup(t)
	defer ctrl.Finish()

	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockUserService.EXPECT().ChangePassword(ctx, int64(1), gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(changePasswordJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	f.ChangePassword(rec, req.WithContext(ctx_util.SetJWTId(req.Context(), "1")))

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestChangePassword_BindError(t *testing.T) {
	ctrl, _, _, f := setup(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(changePasswordJSON))
	rec := httptest.NewRecorder()

	f.ChangePassword(rec, req.WithContext(ctx_util.SetJWTId(req.Context(), "1")))

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgUnsupportedMedia, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, errResp.Status)
	}
}

func TestChangePassword_InvalidValues(t *testing.T) {
	ctrl, _, mockValidator, f := setup(t)
	defer ctrl.Finish()

	mockValidator.EXPECT().Validate(gomock.Any()).Return(rest.NewInvalidArgumentsError(errors.New("")))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(changePasswordJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	
	f.ChangePassword(rec, req.WithContext(ctx_util.SetJWTId(req.Context(), "1")))

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInvalidArguments, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, errResp.Status)
	}
}

func TestChangePassword_NotFound(t *testing.T) {
	ctx := context.Background()

	ctrl, mockUserService, mockValidator, f := setup(t)
	defer ctrl.Finish()

	mockUserService.EXPECT().ChangePassword(ctx, int64(1), gomock.Any()).Return(pg_error.ErrNotFound)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(changePasswordJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	
	f.ChangePassword(rec, req.WithContext(ctx_util.SetJWTId(req.Context(), "1")))

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, app_error.MsgUserNotFound, errResp.Message)
		assert.Equal(t, http.StatusNotFound, errResp.Status)
	}
}

func TestChangePassword_IncorrectPassword(t *testing.T) {
	ctx := context.Background()

	ctrl, mockUserService, mockValidator, f := setup(t)
	defer ctrl.Finish()

	mockUserService.EXPECT().ChangePassword(ctx, int64(1), gomock.Any()).Return(hasher.ErrPasswordMismatch)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(changePasswordJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	f.ChangePassword(rec, req.WithContext(ctx_util.SetJWTId(req.Context(), "1")))

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, app_error.MsgIncorrectPassword, errResp.Message)
		assert.Equal(t, http.StatusUnauthorized, errResp.Status)
	}
}

func TestChangePassword_ServiceError(t *testing.T) {
	ctx := context.Background()

	ctrl, mockUserService, mockValidator, f := setup(t)
	defer ctrl.Finish()

	mockUserService.EXPECT().ChangePassword(ctx, int64(1), gomock.Any()).Return(errors.New(""))
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(changePasswordJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	f.ChangePassword(rec, req.WithContext(ctx_util.SetJWTId(req.Context(), "1")))

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}

func TestChangePassword_JwtNotSet(t *testing.T) {
	ctrl, _, _, f := setup(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	f.ChangePassword(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}
