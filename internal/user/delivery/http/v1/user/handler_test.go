package user_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/hexley21/fixup/internal/common/app_error"
	"github.com/hexley21/fixup/internal/common/enum"
	"github.com/hexley21/fixup/internal/common/util/ctx_util"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/user"
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

func setup(t *testing.T) (
	ctrl *gomock.Controller,
	userServiceMock *mockService.MockUserService,
	validatorMock *mockValidator.MockValidator,
	h *user.Handler,
) {
	ctrl = gomock.NewController(t)
	userServiceMock = mockService.NewMockUserService(ctrl)
	validatorMock = mockValidator.NewMockValidator(ctrl)

	logger := std_logger.New()
	jsonManager := std_json.New()

	h = user.NewHandler(
		handler.NewComponents(logger, std_binder.New(jsonManager), validatorMock, json_writer.New(logger, jsonManager)),
		userServiceMock,
	)

	return
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
	ctrl, userServiceMock, _, h := setup(t)
	defer ctrl.Finish()

	userServiceMock.EXPECT().FindUserById(gomock.Any(), int64(1)).Return(userDto, nil)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	h.FindUserById(rec, req.WithContext(ctx_util.SetParamId(req.Context(), 1)))

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestFindUserById_NotFound(t *testing.T) {
	ctrl, userServiceMock, _, h := setup(t)
	defer ctrl.Finish()

	userServiceMock.EXPECT().FindUserById(gomock.Any(), int64(1)).Return(dto.User{}, pg_error.ErrNotFound)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	h.FindUserById(rec, req.WithContext(ctx_util.SetParamId(req.Context(), 1)))
	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, app_error.MsgUserNotFound, errResp.Message)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	}
}

func TestFindUserById_ServiceError(t *testing.T) {
	ctrl, userServiceMock, _, h := setup(t)
	defer ctrl.Finish()

	userServiceMock.EXPECT().FindUserById(gomock.Any(), int64(1)).Return(dto.User{}, rest.NewInternalServerError(errors.New("")))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	h.FindUserById(rec, req.WithContext(ctx_util.SetParamId(req.Context(), 1)))

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	}
}

func TestFindUserById_IdParamNotSet(t *testing.T) {
	ctrl, _, _, h := setup(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	h.FindUserById(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	}
}

func TestUploadProfilePicture_Success(t *testing.T) {
	ctrl, userServiceMock, _, h := setup(t)
	defer ctrl.Finish()

	userServiceMock.EXPECT().SetProfilePicture(gomock.Any(), int64(1), gomock.Any(), "", gomock.Any(), gomock.Any()).Return(nil)

	body, contentType := createMultipartFormData(t, "image", "test.jpg", fileContent)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	h.UploadProfilePicture(rec, req.WithContext(ctx_util.SetParamId(req.Context(), 1)))

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestUploadProfilePicture_MissingHeaders(t *testing.T) {
	ctrl, _, _, h := setup(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	h.UploadProfilePicture(rec, req.WithContext(ctx_util.SetParamId(req.Context(), 1)))

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgFileReadError, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}

func TestUploadProfilePicture_MissingFile(t *testing.T) {
	ctrl, _, _, h := setup(t)
	defer ctrl.Finish()

	_, contentType := createMultipartFormData(t, "image", "test.jpg", fileContent)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	h.UploadProfilePicture(rec, req.WithContext(ctx_util.SetParamId(req.Context(), 1)))

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgFileReadError, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}

func TestUploadProfilePicture_WrongFormData(t *testing.T) {
	ctrl, _, _, h := setup(t)
	defer ctrl.Finish()

	body, contentType := createMultipartFormData(t, "img", "test.jpg", fileContent)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	h.UploadProfilePicture(rec, req.WithContext(ctx_util.SetParamId(req.Context(), 1)))

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgNoFile, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}

func TestUploadProfilePicture_NotFound(t *testing.T) {
	ctrl, userServiceMock, _, h := setup(t)
	defer ctrl.Finish()

	userServiceMock.EXPECT().SetProfilePicture(gomock.Any(), int64(1), gomock.Any(), "", gomock.Any(), gomock.Any()).Return(pg_error.ErrNotFound)

	body, contentType := createMultipartFormData(t, "image", "test.jpg", fileContent)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	h.UploadProfilePicture(rec, req.WithContext(ctx_util.SetParamId(req.Context(), 1)))

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, app_error.MsgUserNotFound, errResp.Message)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	}
}

func TestUploadProfilePicture_ServiceError(t *testing.T) {
	ctrl, userServiceMock, _, h := setup(t)
	defer ctrl.Finish()

	userServiceMock.EXPECT().SetProfilePicture(gomock.Any(), int64(1), gomock.Any(), "", gomock.Any(), gomock.Any()).Return(errors.New(""))

	body, contentType := createMultipartFormData(t, "image", "test.jpg", fileContent)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	h.UploadProfilePicture(rec, req.WithContext(ctx_util.SetParamId(req.Context(), 1)))

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	}
}

func TestUploadProfilePicture_IdParamNotSet(t *testing.T) {
	ctrl, _, _, h := setup(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	h.UploadProfilePicture(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	}
}

func TestUpdateUserData_Success(t *testing.T) {
	ctrl, userServiceMock, validatorMock, h := setup(t)
	defer ctrl.Finish()

	userServiceMock.EXPECT().UpdateUserDataById(gomock.Any(), int64(1), gomock.Any()).Return(userDto, nil)
	validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(updateUserJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.UpdateUserData(rec, req.WithContext(ctx_util.SetParamId(req.Context(), 1)))

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestUpdateUserDataWith_BindError(t *testing.T) {
	ctrl, _, _, h := setup(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(updateUserJSON))
	rec := httptest.NewRecorder()

	h.UpdateUserData(rec, req.WithContext(ctx_util.SetParamId(req.Context(), 1)))

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgUnsupportedMedia, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}

func TestUpdateUserData_InvalidValues(t *testing.T) {
	ctrl, _, validatorMock, h := setup(t)
	defer ctrl.Finish()

	validatorMock.EXPECT().Validate(gomock.Any()).Return(rest.NewInvalidArgumentsError(errors.New("")))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(updateUserJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.UpdateUserData(rec, req.WithContext(ctx_util.SetParamId(req.Context(), 1)))

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInvalidArguments, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}

func TestUpdateUserData_NotFound(t *testing.T) {
	ctrl, userServiceMock, validatorMock, h := setup(t)
	defer ctrl.Finish()

	userServiceMock.EXPECT().UpdateUserDataById(gomock.Any(), int64(1), gomock.Any()).Return(dto.User{}, pg_error.ErrNotFound)
	validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(updateUserJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.UpdateUserData(rec, req.WithContext(ctx_util.SetParamId(req.Context(), 1)))

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, app_error.MsgUserNotFound, errResp.Message)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	}
}

func TestUpdateUserData_NotChanges(t *testing.T) {
	ctrl, userServiceMock, validatorMock, h := setup(t)
	defer ctrl.Finish()

	userServiceMock.EXPECT().UpdateUserDataById(gomock.Any(), int64(1), gomock.Any()).Return(dto.User{}, pgx.ErrNoRows)
	validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(updateUserJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.UpdateUserData(rec, req.WithContext(ctx_util.SetParamId(req.Context(), 1)))

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, app_error.MsgNoChanges, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}

func TestUpdateUserData_ServiceError(t *testing.T) {
	ctrl, userServiceMock, validatorMock, h := setup(t)
	defer ctrl.Finish()

	userServiceMock.EXPECT().UpdateUserDataById(gomock.Any(), int64(1), gomock.Any()).Return(dto.User{}, errors.New(""))
	validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(updateUserJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.UpdateUserData(rec, req.WithContext(ctx_util.SetParamId(req.Context(), 1)))

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	}
}

func TestUpdateUserData_IdParamNotSet(t *testing.T) {
	ctrl, _, _, h := setup(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	h.UpdateUserData(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	}
}

func TestDeleteUser_Success(t *testing.T) {
	ctrl, userServiceMock, _, h := setup(t)
	defer ctrl.Finish()

	userServiceMock.EXPECT().DeleteUserById(gomock.Any(), int64(1)).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.DeleteUser(rec, req.WithContext(ctx_util.SetParamId(req.Context(), 1)))

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestDelete_NotFound(t *testing.T) {
	ctrl, userServiceMock, _, h := setup(t)
	defer ctrl.Finish()

	userServiceMock.EXPECT().DeleteUserById(gomock.Any(), int64(1)).Return(pg_error.ErrNotFound)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.DeleteUser(rec, req.WithContext(ctx_util.SetParamId(req.Context(), 1)))

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, app_error.MsgUserNotFound, errResp.Message)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	}
}

func TestDeleteUser_ServiceError(t *testing.T) {
	ctrl, userServiceMock, _, h := setup(t)
	defer ctrl.Finish()

	userServiceMock.EXPECT().DeleteUserById(gomock.Any(), int64(1)).Return(errors.New(""))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.DeleteUser(rec, req.WithContext(ctx_util.SetParamId(req.Context(), 1)))

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	}
}

func TestTestDeleteUser_IdParamNotSet(t *testing.T) {
	ctrl, _, _, h := setup(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	h.DeleteUser(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	}
}

func TestChangePassword_Success(t *testing.T) {
	ctrl, userServiceMock, validatorMock, h := setup(t)
	defer ctrl.Finish()

	validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)
	userServiceMock.EXPECT().ChangePassword(gomock.Any(), int64(1), gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(changePasswordJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ChangePassword(rec, req.WithContext(ctx_util.SetJWTId(req.Context(), "1")))

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestChangePassword_BindError(t *testing.T) {
	ctrl, _, _, h := setup(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(changePasswordJSON))
	rec := httptest.NewRecorder()

	h.ChangePassword(rec, req.WithContext(ctx_util.SetJWTId(req.Context(), "1")))

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgUnsupportedMedia, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}

func TestChangePassword_InvalidValues(t *testing.T) {
	ctrl, _, validatorMock, h := setup(t)
	defer ctrl.Finish()

	validatorMock.EXPECT().Validate(gomock.Any()).Return(rest.NewInvalidArgumentsError(errors.New("")))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(changePasswordJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ChangePassword(rec, req.WithContext(ctx_util.SetJWTId(req.Context(), "1")))

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInvalidArguments, errResp.Message)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}

func TestChangePassword_NotFound(t *testing.T) {
	ctrl, userServiceMock, validatorMock, h := setup(t)
	defer ctrl.Finish()

	userServiceMock.EXPECT().ChangePassword(gomock.Any(), int64(1), gomock.Any()).Return(pg_error.ErrNotFound)
	validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(changePasswordJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ChangePassword(rec, req.WithContext(ctx_util.SetJWTId(req.Context(), "1")))

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, app_error.MsgUserNotFound, errResp.Message)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	}
}

func TestChangePassword_IncorrectPassword(t *testing.T) {
	ctrl, userServiceMock, validatorMock, h := setup(t)
	defer ctrl.Finish()

	userServiceMock.EXPECT().ChangePassword(gomock.Any(), int64(1), gomock.Any()).Return(hasher.ErrPasswordMismatch)
	validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(changePasswordJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ChangePassword(rec, req.WithContext(ctx_util.SetJWTId(req.Context(), "1")))

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, app_error.MsgIncorrectPassword, errResp.Message)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	}
}

func TestChangePassword_ServiceError(t *testing.T) {
	ctrl, userServiceMock, validatorMock, h := setup(t)
	defer ctrl.Finish()

	userServiceMock.EXPECT().ChangePassword(gomock.Any(), int64(1), gomock.Any()).Return(errors.New(""))
	validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(changePasswordJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ChangePassword(rec, req.WithContext(ctx_util.SetJWTId(req.Context(), "1")))

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	}
}

func TestChangePassword_JwtNotSet(t *testing.T) {
	ctrl, _, _, h := setup(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ChangePassword(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	}
}
