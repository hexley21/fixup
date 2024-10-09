package category_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/hexley21/fixup/internal/catalog/delivery/http/v1/category"
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

const (
	createCategoryJSON = `{"name":"Home", "typeId":"123"}`
	patchCategoryJSON = `{"name":"Home", "typeId":"123"}`
)

var (
	categoryDTO = dto.CategoryDTO{
		ID:     "123",
		Name:   "Home",
		TypeID: "123",
	}

	categoriesDTO = []dto.CategoryDTO{categoryDTO, categoryDTO}
)

func setup(t *testing.T) (
	ctrl *gomock.Controller,
	mockService *mock_service.MockCategoryService,
	mockValidator *mock_validator.MockValidator,
	h *category.Handler,
) {
	ctrl = gomock.NewController(t)
	mockService = mock_service.NewMockCategoryService(ctrl)
	mockValidator = mock_validator.NewMockValidator(ctrl)

	logger := std_logger.New()
	jsonManager := std_json.New()

	h = category.NewHandler(
		handler.NewComponents(logger, std_binder.New(jsonManager), mockValidator, json_writer.New(logger, jsonManager)),
		mockService,
	)

	return ctrl, mockService, mockValidator, h
}

func TestCreateCategory_Success(t *testing.T) {
	ctrl, mockService, mockValidator, h := setup(t)
	defer ctrl.Finish()

	mockService.EXPECT().CreateCategory(gomock.Any(), gomock.Any()).Return(categoryDTO, nil)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(createCategoryJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.CreateCategory(rec, req)

	var response rest.ApiResponse[dto.CategoryDTO]
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&response)) {
		assert.Equal(t, categoryDTO, response.Data)
	}
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestCreateCategory_EmptyBody(t *testing.T) {
	ctrl, _, _, h := setup(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	h.CreateCategory(rec, req)

	var response rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&response)) {
		assert.Equal(t, rest.MsgEmptyBody, response.Message)
	}
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateCategory_BindError(t *testing.T) {
	ctrl, _, _, h := setup(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(createCategoryJSON))
	rec := httptest.NewRecorder()

	h.CreateCategory(rec, req)

	var response rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&response)) {
		assert.Equal(t, rest.MsgUnsupportedMedia, response.Message)
	}
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateCategory_ValidationError(t *testing.T) {
	ctrl, _, mockValidator, h := setup(t)
	defer ctrl.Finish()

	mockValidator.EXPECT().Validate(gomock.Any()).Return(rest.NewInvalidArgumentsError(errors.New("")))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(createCategoryJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.CreateCategory(rec, req)

	var response rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&response)) {
		assert.Equal(t, rest.MsgInvalidArguments, response.Message)
	}
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateCategory_NameTaken(t *testing.T) {
	ctrl, mockService, mockValidator, h := setup(t)
	defer ctrl.Finish()

	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockService.EXPECT().CreateCategory(gomock.Any(), gomock.Any()).Return(dto.CategoryDTO{}, &pgconn.PgError{Code: pgerrcode.RaiseException})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(createCategoryJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.CreateCategory(rec, req)

	var response rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&response)) {
		assert.Equal(t, app_error.MsgNameAlreadyTaken, response.Message)
	}
	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestCreateCategory_ServiceError(t *testing.T) {
	ctrl, mockService, mockValidator, h := setup(t)
	defer ctrl.Finish()

	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockService.EXPECT().CreateCategory(gomock.Any(), gomock.Any()).Return(dto.CategoryDTO{}, errors.New(""))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(createCategoryJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.CreateCategory(rec, req)

	var response rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&response)) {
		assert.Equal(t, rest.MsgInternalServerError, response.Message)
	}
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetCategories_Success(t *testing.T) {
	ctrl, mockService, _, h := setup(t)
	defer ctrl.Finish()

	mockService.EXPECT().GetCategories(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)

	q := make(url.Values)
	q.Set("page", "1")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	h.GetCategoryies(rec, req)

	var response rest.ApiResponse[[]dto.CategoryTypeDTO]
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&response)) {
		assert.Equal(t, 0, len(response.Data))
	}
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetCategories_InvalidPage(t *testing.T) {
	ctrl, _, _, h := setup(t)
	defer ctrl.Finish()

	q := make(url.Values)
	q.Set("page", "0")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	h.GetCategoryies(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, app_error.MsgInvalidPage, errResp.Message)
	}
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetCategories_InvalidPerPage(t *testing.T) {
	ctrl, _, _, h := setup(t)
	defer ctrl.Finish()

	q := make(url.Values)
	q.Set("page", "1")
	q.Set("per_page", "0")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	h.GetCategoryies(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, app_error.MsgInvalidPerPage, errResp.Message)
	}
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetCategories_NotFound(t *testing.T) {
	ctrl, mockService, _, h := setup(t)
	defer ctrl.Finish()

	mockService.EXPECT().GetCategories(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, pgx.ErrNoRows)

	q := make(url.Values)
	q.Set("page", "1")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	h.GetCategoryies(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, category.MsgCategoryNotFound, errResp.Message)
	}
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetCategories_ServiceError(t *testing.T) {
	ctrl, mockService, _, h := setup(t)
	defer ctrl.Finish()

	mockService.EXPECT().GetCategories(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New(""))

	q := make(url.Values)
	q.Set("page", "1")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	h.GetCategoryies(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
	}
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetCategoriesByTypeId_Success(t *testing.T) {
	ctrl, mockService, _, h := setup(t)
	defer ctrl.Finish()

	r := chi.NewRouter()
	r.Get("/{id}", h.GetCategoryiesByTypeId)

	q := make(url.Values)
	q.Set("page", "1")

	mockService.EXPECT().GetCategoriesByTypeId(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(categoriesDTO, nil)

	req := httptest.NewRequest(http.MethodGet, "/1?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	var response rest.ApiResponse[[]dto.CategoryDTO]
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&response)) {
		assert.NotEmpty(t, response.Data)
	}
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetCategoriesByTypeId_InvalidId(t *testing.T) {
	ctrl, _, _, h := setup(t)
	defer ctrl.Finish()

	r := chi.NewRouter()
	r.Get("/{id}", h.GetCategoryiesByTypeId)

	req := httptest.NewRequest(http.MethodGet, "/ABC", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	var response rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&response)) {
		assert.Equal(t, rest.MsgInvalidId, response.Message)
	}
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetCategoriesByTypeId_InvalidPage(t *testing.T) {
	ctrl, _, _, h := setup(t)
	defer ctrl.Finish()

	r := chi.NewRouter()
	r.Get("/{id}", h.GetCategoryiesByTypeId)

	q := make(url.Values)
	q.Set("page", "0")

	req := httptest.NewRequest(http.MethodGet, "/1?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, app_error.MsgInvalidPage, errResp.Message)
	}
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetCategoriesByTypeId_InvalidPerPage(t *testing.T) {
	ctrl, _, _, h := setup(t)
	defer ctrl.Finish()

	r := chi.NewRouter()
	r.Get("/{id}", h.GetCategoryiesByTypeId)

	q := make(url.Values)
	q.Set("page", "1")
	q.Set("per_page", "0")

	req := httptest.NewRequest(http.MethodGet, "/1?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, app_error.MsgInvalidPerPage, errResp.Message)
	}
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetCategoriesByTypeId_NotFound(t *testing.T) {
	ctrl, mockService, _, h := setup(t)
	defer ctrl.Finish()

	r := chi.NewRouter()
	r.Get("/{id}", h.GetCategoryiesByTypeId)

	q := make(url.Values)
	q.Set("page", "1")

	mockService.EXPECT().GetCategoriesByTypeId(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, pgx.ErrNoRows)

	req := httptest.NewRequest(http.MethodGet, "/1?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, category.MsgCategoryNotFound, errResp.Message)
	}
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetCategoriesByTypeId_ServiceError(t *testing.T) {
	ctrl, mockService, _, h := setup(t)
	defer ctrl.Finish()

	r := chi.NewRouter()
	r.Get("/{id}", h.GetCategoryiesByTypeId)

	q := make(url.Values)
	q.Set("page", "1")

	mockService.EXPECT().GetCategoriesByTypeId(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New(""))

	req := httptest.NewRequest(http.MethodGet, "/1?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
	}
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetCategoryById_Success(t *testing.T) {
	ctrl, mockService, _, h := setup(t)
	defer ctrl.Finish()

	r := chi.NewRouter()
	r.Get("/{id}", h.GetCategoryById)

	mockService.EXPECT().GetCategoryById(gomock.Any(), gomock.Any()).Return(categoryDTO, nil)

	req := httptest.NewRequest(http.MethodGet, "/1", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	var response rest.ApiResponse[dto.CategoryDTO]
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&response)) {
		assert.NotEmpty(t, response.Data)
	}
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetCategoryTypeById_InvalidID(t *testing.T) {
	ctrl, _, _, h := setup(t)
	defer ctrl.Finish()

	r := chi.NewRouter()
	r.Get("/{id}", h.GetCategoryById)

	req := httptest.NewRequest(http.MethodGet, "/ABC", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInvalidId, errResp.Message)
	}
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetCategoryTypeById_NotFound(t *testing.T) {
	ctrl, mockService, _, h := setup(t)
	defer ctrl.Finish()

	r := chi.NewRouter()
	r.Get("/{id}", h.GetCategoryById)

	mockService.EXPECT().GetCategoryById(gomock.Any(), gomock.Any()).Return(dto.CategoryDTO{}, pgx.ErrNoRows)

	req := httptest.NewRequest(http.MethodGet, "/1", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, category.MsgCategoryNotFound, errResp.Message)
	}
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetCategoryById_ServiceError(t *testing.T) {
	ctrl, mockService, _, h := setup(t)
	defer ctrl.Finish()

	r := chi.NewRouter()
	r.Get("/{id}", h.GetCategoryById)

	mockService.EXPECT().GetCategoryById(gomock.Any(), gomock.Any()).Return(dto.CategoryDTO{}, errors.New(""))

	req := httptest.NewRequest(http.MethodGet, "/1", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
	}
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestPatchCategoryById_Success(t *testing.T) {
	ctrl, mockService, mockValidator, h := setup(t)
	defer ctrl.Finish()

	r := chi.NewRouter()
	r.Patch("/{id}", h.PatchCategoryById)

	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockService.EXPECT().UpdateCategoryById(gomock.Any(), gomock.Any(), gomock.Any()).Return(categoryDTO, nil)

	req := httptest.NewRequest(http.MethodPatch, "/123", strings.NewReader(patchCategoryJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	var response rest.ApiResponse[dto.CategoryDTO]
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&response)) {
		assert.Equal(t, categoryDTO, response.Data)
	}
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestPatchCategoryTypeById_InvalidId(t *testing.T) {
	ctrl, _, _, h := setup(t)
	defer ctrl.Finish()

	r := chi.NewRouter()
	r.Patch("/{id}", h.PatchCategoryById)

	req := httptest.NewRequest(http.MethodPatch, "/ABC", strings.NewReader(patchCategoryJSON))
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
	r.Patch("/{id}", h.PatchCategoryById)


	req := httptest.NewRequest(http.MethodPatch, "/123", strings.NewReader(patchCategoryJSON))
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
	r.Patch("/{id}", h.PatchCategoryById)

	mockValidator.EXPECT().Validate(gomock.Any()).Return(rest.NewInvalidArgumentsError(errors.New("")))

	req := httptest.NewRequest(http.MethodPatch, "/123", strings.NewReader(patchCategoryJSON))
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
	r.Patch("/{id}", h.PatchCategoryById)

	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockService.EXPECT().UpdateCategoryById(gomock.Any(), gomock.Any(), gomock.Any()).Return(dto.CategoryDTO{}, pgx.ErrNoRows)

	req := httptest.NewRequest(http.MethodPatch, "/123", strings.NewReader(patchCategoryJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	var response rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&response)) {
		assert.Equal(t, category.MsgCategoryNotFound, response.Message)
	}
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestPatchCategoryTypeById_AlreadyTaken(t *testing.T) {
	ctrl, mockService, mockValidator, h := setup(t)
	defer ctrl.Finish()

	r := chi.NewRouter()
	r.Patch("/{id}", h.PatchCategoryById)

	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockService.EXPECT().UpdateCategoryById(gomock.Any(), gomock.Any(), gomock.Any()).Return(dto.CategoryDTO{}, &pgconn.PgError{Code: pgerrcode.RaiseException})

	req := httptest.NewRequest(http.MethodPatch, "/123", strings.NewReader(patchCategoryJSON))
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
	r.Patch("/{id}", h.PatchCategoryById)

	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)
	mockService.EXPECT().UpdateCategoryById(gomock.Any(), gomock.Any(), gomock.Any()).Return(dto.CategoryDTO{}, errors.New(""))


	req := httptest.NewRequest(http.MethodPatch, "/123", strings.NewReader(patchCategoryJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	var response rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&response)) {
		assert.Equal(t, rest.MsgInternalServerError, response.Message)
	}
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}


func TestDeleteCategoryById_Success(t *testing.T) {
	ctrl, mockService, _, h := setup(t)
	defer ctrl.Finish()

	r := chi.NewRouter()
	r.Delete("/{id}", h.DeleteCategoryById)

	mockService.EXPECT().DeleteCategoryById(gomock.Any(), gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/123", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Empty(t, rec.Body.String())
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestDeleteCategoryTypeById_InvalidId(t *testing.T) {
	ctrl, _, _, h := setup(t)
	defer ctrl.Finish()

	r := chi.NewRouter()
	r.Delete("/{id}", h.DeleteCategoryById)


	req := httptest.NewRequest(http.MethodDelete, "/ABC", nil)
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
	r.Delete("/{id}", h.DeleteCategoryById)

	mockService.EXPECT().DeleteCategoryById(gomock.Any(), gomock.Any()).Return(pg_error.ErrNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/123", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	var response rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&response)) {
		assert.Equal(t, category.MsgCategoryNotFound, response.Message)
	}
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestDeleteCategoryTypeById_ServiceError(t *testing.T) {
	ctrl, mockService, _, h := setup(t)
	defer ctrl.Finish()

	r := chi.NewRouter()
	r.Delete("/{id}", h.DeleteCategoryById)

	mockService.EXPECT().DeleteCategoryById(gomock.Any(), gomock.Any()).Return(errors.New(""))

	req := httptest.NewRequest(http.MethodDelete, "/123", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	var response rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&response)) {
		assert.Equal(t, rest.MsgInternalServerError, response.Message)
	}
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
