package user_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
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
	userDTO = dto.User{
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

func TestFindUserById(t *testing.T) {
	ctrl, userServiceMock, _, h := setup(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		mockSetup     func()
		expectedCode  int
		expectedError string
	}{
		{
			name: "Success",
			mockSetup: func() {
				userServiceMock.EXPECT().FindUserById(gomock.Any(), int64(1)).Return(userDTO, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Not Found",
			mockSetup: func() {
				userServiceMock.EXPECT().FindUserById(gomock.Any(), int64(1)).Return(dto.User{}, pg_error.ErrNotFound)
			},
			expectedCode:  http.StatusNotFound,
			expectedError: app_error.MsgUserNotFound,
		},
		{
			name: "Service Error",
			mockSetup: func() {
				userServiceMock.EXPECT().FindUserById(gomock.Any(), int64(1)).Return(dto.User{}, rest.NewInternalServerError(errors.New("")))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: rest.MsgInternalServerError,
		},
		{
			name:          "Id Param Not Set",
			mockSetup:     func() {},
			expectedCode:  http.StatusInternalServerError,
			expectedError: rest.MsgInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			var req *http.Request
			if tt.name == "Id Param Not Set" {
				req = httptest.NewRequest(http.MethodPost, "/", nil)
			} else {
				req = httptest.NewRequest(http.MethodPost, "/", nil)
				req = req.WithContext(ctx_util.SetParamId(req.Context(), 1))
			}

			rec := httptest.NewRecorder()

			h.FindUserById(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code)

			if tt.expectedError != "" {
				var errResp rest.ErrorResponse
				if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
					assert.Equal(t, tt.expectedError, errResp.Message)
				}
			}
		})
	}
}

func TestUploadProfilePicture(t *testing.T) {
	ctrl, userServiceMock, _, h := setup(t)
	defer ctrl.Finish()

	type testCase struct {
		name          string
		mockSetup     func(tt *testCase)
		body          io.Reader
		contentType   string
		expectedCode  int
		expectedError string
	}

	tests := []testCase{
		{
			name: "Success",
			mockSetup: func(tt *testCase) {
				tt.body, tt.contentType = createMultipartFormData(t, "image", "test.jpg", fileContent)
				userServiceMock.EXPECT().SetProfilePicture(gomock.Any(), int64(1), gomock.Any(), "", gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedCode: http.StatusNoContent,
		},
		{
			name:          "Missing Headers",
			mockSetup:     func(tt *testCase) {},
			body:          nil,
			expectedCode:  http.StatusBadRequest,
			expectedError: rest.MsgFileReadError,
		},
		{
			name: "Missing File",
			mockSetup: func(tt *testCase) {
				_, tt.contentType = createMultipartFormData(t, "image", "test.jpg", fileContent)
			},
			body:          nil,
			expectedCode:  http.StatusBadRequest,
			expectedError: rest.MsgFileReadError,
		},
		{
			name: "Wrong Form Data",
			mockSetup: func(tt *testCase) {
				tt.body, tt.contentType = createMultipartFormData(t, "img", "test.jpg", fileContent)
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: rest.MsgNoFile,
		},
		{
			name: "Not Found",
			mockSetup: func(tt *testCase) {
				tt.body, tt.contentType = createMultipartFormData(t, "image", "test.jpg", fileContent)
				userServiceMock.EXPECT().SetProfilePicture(gomock.Any(), int64(1), gomock.Any(), "", gomock.Any(), gomock.Any()).Return(pg_error.ErrNotFound)
			},
			expectedCode:  http.StatusNotFound,
			expectedError: app_error.MsgUserNotFound,
		},
		{
			name: "Service Error",
			mockSetup: func(tt *testCase) {
				tt.body, tt.contentType = createMultipartFormData(t, "image", "test.jpg", fileContent)
				userServiceMock.EXPECT().SetProfilePicture(gomock.Any(), int64(1), gomock.Any(), "", gomock.Any(), gomock.Any()).Return(errors.New(""))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: rest.MsgInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(&tt)

			req := httptest.NewRequest(http.MethodPost, "/", tt.body)
			if tt.body != nil {
				req.Header.Set("Content-Type", tt.contentType)
			}
			rec := httptest.NewRecorder()

			h.UploadProfilePicture(rec, req.WithContext(ctx_util.SetParamId(req.Context(), 1)))

			assert.Equal(t, tt.expectedCode, rec.Code)

			if tt.expectedError != "" {
				var errResp rest.ErrorResponse
				if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
					assert.Equal(t, tt.expectedError, errResp.Message)
				}
			}
		})
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

func TestUpdateUserData(t *testing.T) {
	ctrl, userServiceMock, validatorMock, h := setup(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		mockSetup     func()
		body          io.Reader
		expectedCode  int
		expectedError string
	}{
		{
			name: "Success",
			mockSetup: func() {
				userServiceMock.EXPECT().UpdateUserDataById(gomock.Any(), int64(1), gomock.Any()).Return(userDTO, nil)
				validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)
			},
			body:         strings.NewReader(updateUserJSON),
			expectedCode: http.StatusOK,
		},
		{
			name:          "Bind Error",
			mockSetup:     func() {},
			body:          strings.NewReader(updateUserJSON),
			expectedCode:  http.StatusBadRequest,
			expectedError: rest.MsgUnsupportedMedia,
		},
		{
			name: "Invalid Values",
			mockSetup: func() {
				validatorMock.EXPECT().Validate(gomock.Any()).Return(rest.NewInvalidArgumentsError(errors.New("")))
			},
			body:          strings.NewReader(updateUserJSON),
			expectedCode:  http.StatusBadRequest,
			expectedError: rest.MsgInvalidArguments,
		},
		{
			name: "Not Found",
			mockSetup: func() {
				userServiceMock.EXPECT().UpdateUserDataById(gomock.Any(), int64(1), gomock.Any()).Return(dto.User{}, pg_error.ErrNotFound)
				validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)
			},
			body:          strings.NewReader(updateUserJSON),
			expectedCode:  http.StatusNotFound,
			expectedError: app_error.MsgUserNotFound,
		},
		{
			name: "No Changes",
			mockSetup: func() {
				userServiceMock.EXPECT().UpdateUserDataById(gomock.Any(), int64(1), gomock.Any()).Return(dto.User{}, pgx.ErrNoRows)
				validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)
			},
			body:          strings.NewReader(updateUserJSON),
			expectedCode:  http.StatusBadRequest,
			expectedError: app_error.MsgNoChanges,
		},
		{
			name: "Service Error",
			mockSetup: func() {
				userServiceMock.EXPECT().UpdateUserDataById(gomock.Any(), int64(1), gomock.Any()).Return(dto.User{}, errors.New(""))
				validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)
			},
			body:          strings.NewReader(updateUserJSON),
			expectedCode:  http.StatusInternalServerError,
			expectedError: rest.MsgInternalServerError,
		},
		{
			name:          "Id Param Not Set",
			mockSetup:     func() {},
			body:          nil,
			expectedCode:  http.StatusInternalServerError,
			expectedError: rest.MsgInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest(http.MethodPost, "/", tt.body)
			if tt.name != "Bind Error" {
				req.Header.Set("Content-Type", "application/json")
			}
			rec := httptest.NewRecorder()

			if tt.name == "Id Param Not Set" {
				h.UpdateUserData(rec, req)
			} else {
				h.UpdateUserData(rec, req.WithContext(ctx_util.SetParamId(req.Context(), 1)))
			}

			assert.Equal(t, tt.expectedCode, rec.Code)

			if tt.expectedError != "" {
				var errResp rest.ErrorResponse
				if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
					assert.Equal(t, tt.expectedError, errResp.Message)
				}
			}
		})
	}
}

func TestDeleteUser(t *testing.T) {
	ctrl, userServiceMock, _, h := setup(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		mockSetup     func()
		expectedCode  int
		expectedError string
	}{
		{
			name: "Success",
			mockSetup: func() {
				userServiceMock.EXPECT().DeleteUserById(gomock.Any(), int64(1)).Return(nil)
			},
			expectedCode: http.StatusNoContent,
		},
		{
			name: "Not Found",
			mockSetup: func() {
				userServiceMock.EXPECT().DeleteUserById(gomock.Any(), int64(1)).Return(pg_error.ErrNotFound)
			},
			expectedCode:  http.StatusNotFound,
			expectedError: app_error.MsgUserNotFound,
		},
		{
			name: "Service Error",
			mockSetup: func() {
				userServiceMock.EXPECT().DeleteUserById(gomock.Any(), int64(1)).Return(errors.New(""))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: rest.MsgInternalServerError,
		},
		{
			name:          "Id Param Not Set",
			mockSetup:     func() {},
			expectedCode:  http.StatusInternalServerError,
			expectedError: rest.MsgInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest(http.MethodPost, "/", nil)
			if tt.name != "Id Param Not Set" {
				req.Header.Set("Content-Type", "application/json")
				req = req.WithContext(ctx_util.SetParamId(req.Context(), 1))
			}
			rec := httptest.NewRecorder()

			h.DeleteUser(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code)

			if tt.expectedError != "" {
				var errResp rest.ErrorResponse
				if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
					assert.Equal(t, tt.expectedError, errResp.Message)
				}
			}
		})
	}
}

func TestChangePassword(t *testing.T) {
	ctrl, userServiceMock, validatorMock, h := setup(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		mockSetup     func()
		body          io.Reader
		expectedCode  int
		expectedError string
	}{
		{
			name: "Success",
			mockSetup: func() {
				validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)
				userServiceMock.EXPECT().ChangePassword(gomock.Any(), int64(1), gomock.Any()).Return(nil)
			},
			expectedCode: http.StatusNoContent,
		},
		{
			name:          "Bind Error",
			mockSetup:     func() {},
			expectedCode:  http.StatusBadRequest,
			expectedError: rest.MsgUnsupportedMedia,
		},
		{
			name: "Invalid Values",
			mockSetup: func() {
				validatorMock.EXPECT().Validate(gomock.Any()).Return(rest.NewInvalidArgumentsError(errors.New("")))
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: rest.MsgInvalidArguments,
		},
		{
			name: "Not Found",
			mockSetup: func() {
				validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)
				userServiceMock.EXPECT().ChangePassword(gomock.Any(), int64(1), gomock.Any()).Return(pg_error.ErrNotFound)
			},
			expectedCode:  http.StatusNotFound,
			expectedError: app_error.MsgUserNotFound,
		},
		{
			name: "Incorrect Password",
			mockSetup: func() {
				validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)
				userServiceMock.EXPECT().ChangePassword(gomock.Any(), int64(1), gomock.Any()).Return(hasher.ErrPasswordMismatch)
			},
			expectedCode:  http.StatusUnauthorized,
			expectedError: app_error.MsgIncorrectPassword,
		},
		{
			name: "Service Error",
			mockSetup: func() {
				validatorMock.EXPECT().Validate(gomock.Any()).Return(nil)
				userServiceMock.EXPECT().ChangePassword(gomock.Any(), int64(1), gomock.Any()).Return(errors.New(""))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: rest.MsgInternalServerError,
		},
		{
			name:          "JWT Not Set",
			mockSetup:     func() {},
			expectedCode:  http.StatusInternalServerError,
			expectedError: rest.MsgInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(changePasswordJSON))
			if tt.name != "JWT Not Set" {
				req = req.WithContext(ctx_util.SetJWTId(req.Context(), "1"))
			}
			if tt.name != "Bind Error" {
				req.Header.Set("Content-Type", "application/json")
			}
			rec := httptest.NewRecorder()

			h.ChangePassword(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code)

			if tt.expectedError != "" {
				var errResp rest.ErrorResponse
				if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
					assert.Equal(t, tt.expectedError, errResp.Message)
				}
			}
		})
	}
}
