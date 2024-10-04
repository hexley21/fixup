package category_type_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/hexley21/fixup/internal/catalog/delivery/http/v1/category_type"
	"github.com/hexley21/fixup/internal/catalog/delivery/http/v1/dto"
	mock_service "github.com/hexley21/fixup/internal/catalog/service/mock"
	"github.com/hexley21/fixup/internal/common/app_error"
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
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var (
	createCategoryTypeJSON = `{"name":"Home"}`

	categoryType = dto.CategoryTypeDTO{
		ID:   "123",
		Name: "Home",
	}

	emptyCategoryTypeArray = []dto.CategoryTypeDTO{}
)

func setup(t *testing.T) (
	ctrl *gomock.Controller,
	mockService *mock_service.MockCategoryTypeService,
	mockValidator *mock_validator.MockValidator,
	h *category_type.CategoryTypeHandler,
) {
	ctrl = gomock.NewController(t)
	mockService = mock_service.NewMockCategoryTypeService(ctrl)
	mockValidator = mock_validator.NewMockValidator(ctrl)

	logger := std_logger.New()
	jsonManager := std_json.New()

	h = category_type.NewHandler(
		handler.NewComponents(logger, std_binder.New(jsonManager), mockValidator, json_writer.New(logger, jsonManager)),
		mockService,
	)

	return ctrl, mockService, mockValidator, h
}

func TestCreateCategoryType_Success(t *testing.T) {
	ctrl, mockService, mockValidator, h := setup(t)
	defer ctrl.Finish()

	mockService.EXPECT().CreateCategoryType(gomock.Any(), gomock.Any()).Return(categoryType, nil)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(createCategoryTypeJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.CreateCategoryType(rec, req)

	var response rest.ApiResponse[dto.CategoryTypeDTO]
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&response)) {
		assert.Equal(t, categoryType, response.Data)
	}
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestCreateCategoryType_EmptyBody(t *testing.T) {
	ctrl, _, _, h := setup(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	h.CreateCategoryType(rec, req)

	var response rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&response)) {
		assert.Equal(t, rest.MsgEmptyBody, response.Message)
	}
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateCategoryType_BindError(t *testing.T) {
	ctrl, _, _, h := setup(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(createCategoryTypeJSON))
	rec := httptest.NewRecorder()

	h.CreateCategoryType(rec, req)

	var response rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&response)) {
		assert.Equal(t, rest.MsgUnsupportedMedia, response.Message)
	}
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateCategoryType_ValidationError(t *testing.T) {
	ctrl, _, mockValidator, h := setup(t)
	defer ctrl.Finish()

	mockValidator.EXPECT().Validate(gomock.Any()).Return(rest.NewInvalidArgumentsError(errors.New("")))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(createCategoryTypeJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.CreateCategoryType(rec, req)

	var response rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&response)) {
		assert.Equal(t, rest.MsgInvalidArguments, response.Message)
	}
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateCategoryType_NameTaken(t *testing.T) {
	ctrl, mockService, mockValidator, h := setup(t)
	defer ctrl.Finish()

	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockService.EXPECT().CreateCategoryType(gomock.Any(), gomock.Any()).Return(dto.CategoryTypeDTO{}, &pgconn.PgError{Code: pgerrcode.UniqueViolation})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(createCategoryTypeJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.CreateCategoryType(rec, req)

	var response rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&response)) {
		assert.Equal(t, app_error.MsgNameAlreadyTaken, response.Message)
	}
	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestCreateCategoryType_ServiceError(t *testing.T) {
	ctrl, mockService, mockValidator, h := setup(t)
	defer ctrl.Finish()

	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockService.EXPECT().CreateCategoryType(gomock.Any(), gomock.Any()).Return(dto.CategoryTypeDTO{}, errors.New(""))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(createCategoryTypeJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.CreateCategoryType(rec, req)

	var response rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&response)) {
		assert.Equal(t, rest.MsgInternalServerError, response.Message)
	}
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetCategoryTypes_Success(t *testing.T) {
	ctrl, mockService, _, h := setup(t)
	defer ctrl.Finish()

	mockService.EXPECT().GetCategoryTypes(gomock.Any(), gomock.Any(), gomock.Any()).Return(emptyCategoryTypeArray, nil)

	q := make(url.Values)
	q.Set("page", "1")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	h.GetCategoryTypes(rec, req)

	var response rest.ApiResponse[[]dto.CategoryTypeDTO]
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&response)) {
		assert.Equal(t, 0, len(response.Data))
	}
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetCategoryTypes_InvalidPage(t *testing.T) {
	ctrl, _, _, h := setup(t)
	defer ctrl.Finish()

	q := make(url.Values)
	q.Set("page", "0")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	h.GetCategoryTypes(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, app_error.MsgInvalidPage, errResp.Message)
	}
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetCategoryTypes_InvalidPerPage(t *testing.T) {
	ctrl, _, _, h := setup(t)
	defer ctrl.Finish()

	q := make(url.Values)
	q.Set("page", "1")
	q.Set("per_page", "0")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	h.GetCategoryTypes(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, app_error.MsgInvalidPerPage, errResp.Message)
	}
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetCategoryTypes_NotFound(t *testing.T) {
	ctrl, mockService, _, h := setup(t)
	defer ctrl.Finish()

	mockService.EXPECT().GetCategoryTypes(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, pgx.ErrNoRows)

	q := make(url.Values)
	q.Set("page", "1")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	h.GetCategoryTypes(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, category_type.MsgCategoryTypeNotFound, errResp.Message)
	}
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetCategoryTypes_ServiceError(t *testing.T) {
	ctrl, mockService, _, h := setup(t)
	defer ctrl.Finish()

	mockService.EXPECT().GetCategoryTypes(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New(""))

	q := make(url.Values)
	q.Set("page", "1")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	h.GetCategoryTypes(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
	}
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetCategoryTypeById_Success(t *testing.T) {
	ctrl, mockService, _, h := setup(t)
	defer ctrl.Finish()

	r := chi.NewRouter()
	r.Get("/{id}", h.GetCategoryTypeById)

	mockService.EXPECT().GetCategoryTypeById(gomock.Any(), gomock.Any()).Return(categoryType, nil)

	req := httptest.NewRequest(http.MethodGet, "/1", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	var response rest.ApiResponse[dto.CategoryTypeDTO]
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&response)) {
		assert.Equal(t, categoryType, response.Data)
	}
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetCategoryTypeById_NotFound(t *testing.T) {
	ctrl, mockService, _, h := setup(t)
	defer ctrl.Finish()

	r := chi.NewRouter()
	r.Get("/{id}", h.GetCategoryTypeById)

	mockService.EXPECT().GetCategoryTypeById(gomock.Any(), gomock.Any()).Return(dto.CategoryTypeDTO{}, pgx.ErrNoRows)

	req := httptest.NewRequest(http.MethodGet, "/1", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, category_type.MsgCategoryTypeNotFound, errResp.Message)
	}
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetCategoryTypeById_ServiceError(t *testing.T) {
	ctrl, mockService, _, h := setup(t)
	defer ctrl.Finish()

	r := chi.NewRouter()
	r.Get("/{id}", h.GetCategoryTypeById)

	mockService.EXPECT().GetCategoryTypeById(gomock.Any(), gomock.Any()).Return(dto.CategoryTypeDTO{}, errors.New(""))

	req := httptest.NewRequest(http.MethodGet, "/1", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
	}
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestPatchCategoryTypeById_Success(t *testing.T) {
	ctrl, mockService, mockValidator, h := setup(t)
	defer ctrl.Finish()

	r := chi.NewRouter()
	r.Get("/{id}", h.PatchCategoryTypeById)

	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockService.EXPECT().UpdateCategoryTypeById(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodGet, "/123", strings.NewReader(createCategoryTypeJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	var response rest.ApiResponse[dto.CategoryTypeDTO]
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&response)) {
		assert.Equal(t, categoryType, response.Data)
	}
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestPatchCategoryTypeById_InvalidId(t *testing.T) {
	ctrl, _, _, h := setup(t)
	defer ctrl.Finish()

	r := chi.NewRouter()
	r.Get("/{id}", h.PatchCategoryTypeById)


	req := httptest.NewRequest(http.MethodGet, "/abc", strings.NewReader(createCategoryTypeJSON))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	var response rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&response)) {
		assert.Equal(t, rest.MsgInvalidId, response.Message)
	}
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPatchCategoryTypeById_BindError(t *testing.T) {
	ctrl, _, _, h := setup(t)
	defer ctrl.Finish()

	r := chi.NewRouter()
	r.Get("/{id}", h.PatchCategoryTypeById)


	req := httptest.NewRequest(http.MethodGet, "/123", strings.NewReader(createCategoryTypeJSON))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	var response rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&response)) {
		assert.Equal(t, rest.MsgUnsupportedMedia, response.Message)
	}
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPatchCategoryTypeById_ValidationError(t *testing.T) {
	ctrl, _, mockValidator, h := setup(t)
	defer ctrl.Finish()

	r := chi.NewRouter()
	r.Get("/{id}", h.PatchCategoryTypeById)

	mockValidator.EXPECT().Validate(gomock.Any()).Return(rest.NewInvalidArgumentsError(errors.New("")))

	req := httptest.NewRequest(http.MethodGet, "/123", strings.NewReader(createCategoryTypeJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	var response rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&response)) {
		assert.Equal(t, rest.MsgInvalidArguments, response.Message)
	}
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPatchCategoryTypeById_NotFound(t *testing.T) {
	ctrl, mockService, mockValidator, h := setup(t)
	defer ctrl.Finish()

	r := chi.NewRouter()
	r.Get("/{id}", h.PatchCategoryTypeById)

	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockService.EXPECT().UpdateCategoryTypeById(gomock.Any(), gomock.Any(), gomock.Any()).Return(pg_error.ErrNotFound)

	req := httptest.NewRequest(http.MethodGet, "/123", strings.NewReader(createCategoryTypeJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	var response rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&response)) {
		assert.Equal(t, category_type.MsgCategoryTypeNotFound, response.Message)
	}
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestPatchCategoryTypeById_AlreadyTaken(t *testing.T) {
	ctrl, mockService, mockValidator, h := setup(t)
	defer ctrl.Finish()

	r := chi.NewRouter()
	r.Get("/{id}", h.PatchCategoryTypeById)

	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockService.EXPECT().UpdateCategoryTypeById(gomock.Any(), gomock.Any(), gomock.Any()).Return(&pgconn.PgError{Code: pgerrcode.UniqueViolation})

	req := httptest.NewRequest(http.MethodGet, "/123", strings.NewReader(createCategoryTypeJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	var response rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&response)) {
		assert.Equal(t, app_error.MsgNameAlreadyTaken, response.Message)
	}
	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestPatchCategoryTypeById_ServiceError(t *testing.T) {
	ctrl, mockService, mockValidator, h := setup(t)
	defer ctrl.Finish()

	r := chi.NewRouter()
	r.Get("/{id}", h.PatchCategoryTypeById)

	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockService.EXPECT().UpdateCategoryTypeById(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New(""))

	req := httptest.NewRequest(http.MethodGet, "/123", strings.NewReader(createCategoryTypeJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	var response rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&response)) {
		assert.Equal(t, rest.MsgInternalServerError, response.Message)
	}
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestDeleteCategoryTypeById_Success(t *testing.T) {
	ctrl, mockService, _, h := setup(t)
	defer ctrl.Finish()

	r := chi.NewRouter()
	r.Get("/{id}", h.DeleteCategoryTypeById)

	mockService.EXPECT().DeleteCategoryTypeById(gomock.Any(), gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodGet, "/123", strings.NewReader(createCategoryTypeJSON))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Empty(t, rec.Body.String())
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestDeleteCategoryTypeById_InvalidId(t *testing.T) {
	ctrl, _, _, h := setup(t)
	defer ctrl.Finish()

	r := chi.NewRouter()
	r.Get("/{id}", h.DeleteCategoryTypeById)


	req := httptest.NewRequest(http.MethodGet, "/abc", strings.NewReader(createCategoryTypeJSON))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	var response rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&response)) {
		assert.Equal(t, rest.MsgInvalidId, response.Message)
	}
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestDeleteCategoryTypeById_NotFound(t *testing.T) {
	ctrl, mockService, _, h := setup(t)
	defer ctrl.Finish()

	r := chi.NewRouter()
	r.Get("/{id}", h.DeleteCategoryTypeById)

	mockService.EXPECT().DeleteCategoryTypeById(gomock.Any(), gomock.Any()).Return(pg_error.ErrNotFound)

	req := httptest.NewRequest(http.MethodGet, "/123", strings.NewReader(createCategoryTypeJSON))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	var response rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&response)) {
		assert.Equal(t, category_type.MsgCategoryTypeNotFound, response.Message)
	}
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestDeleteCategoryTypeById_ServiceError(t *testing.T) {
	ctrl, mockService, _, h := setup(t)
	defer ctrl.Finish()

	r := chi.NewRouter()
	r.Get("/{id}", h.DeleteCategoryTypeById)

	mockService.EXPECT().DeleteCategoryTypeById(gomock.Any(), gomock.Any()).Return(errors.New(""))

	req := httptest.NewRequest(http.MethodGet, "/123", strings.NewReader(createCategoryTypeJSON))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	var response rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&response)) {
		assert.Equal(t, rest.MsgInternalServerError, response.Message)
	}
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
