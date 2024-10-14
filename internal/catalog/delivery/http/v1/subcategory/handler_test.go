package subcategory_test

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/hexley21/fixup/internal/catalog/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/catalog/delivery/http/v1/dto/mapper"
	"github.com/hexley21/fixup/internal/catalog/delivery/http/v1/subcategory"
	"github.com/hexley21/fixup/internal/catalog/entity"
	"github.com/hexley21/fixup/internal/catalog/service"
	mock_service "github.com/hexley21/fixup/internal/catalog/service/mock"
	"github.com/hexley21/fixup/pkg/http/binder/std_binder"
	"github.com/hexley21/fixup/pkg/http/handler"
	"github.com/hexley21/fixup/pkg/http/json/std_json"
	"github.com/hexley21/fixup/pkg/http/rest"
	"github.com/hexley21/fixup/pkg/http/writer/json_writer"
	"github.com/hexley21/fixup/pkg/logger/std_logger"
	mock_validator "github.com/hexley21/fixup/pkg/validator/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

const (
	id int32 = 1
	page    int32 = 0
	perPage int32 = 10
)

var (
	subcategoryInfoEntity = entity.SubcategoryInfo{
		Name:       "Leakage",
		CategoryID: id,
	}
	subcategoryEntity = entity.Subcategory{
		ID:              id,
		SubcategoryInfo: subcategoryInfoEntity,
	}

	subcategoryDTO     = mapper.SubcategoryToDTO(subcategoryEntity)
)

func setup(t *testing.T) (
	ctrl *gomock.Controller,
	mockSubcategoryService *mock_service.MockSubcategory,
	mockValidator *mock_validator.MockValidator,
	h *subcategory.Handler,
) {
	ctrl = gomock.NewController(t)
	mockSubcategoryService = mock_service.NewMockSubcategory(ctrl)
	mockValidator = mock_validator.NewMockValidator(ctrl)

	logger := std_logger.New()
	jsonManager := std_json.New()

	h = subcategory.NewHandler(
		handler.NewComponents(logger, std_binder.New(jsonManager), mockValidator, json_writer.New(logger, jsonManager)),
		mockSubcategoryService,
		50,
		100,
	)

	return
}

func TestGet(t *testing.T) {
	ctrl, serviceMock, _, h := setup(t)
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
				serviceMock.EXPECT().Get(gomock.Any(), id).Return(subcategoryEntity, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Not Found",
			mockSetup: func() {
				serviceMock.EXPECT().Get(gomock.Any(), id).Return(entity.Subcategory{}, service.ErrSubcategoryNotFound)
			},
			expectedCode:  http.StatusNotFound,
			expectedError: service.ErrSubcategoryNotFound.Error(),
		},
		{
			name: "Service Error",
			mockSetup: func() {
				serviceMock.EXPECT().Get(gomock.Any(), id).Return(entity.Subcategory{}, rest.NewInternalServerError(errors.New("")))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: rest.MsgInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			r := chi.NewRouter()
			r.Get("/{subcategory_id}", h.Get)

			req := httptest.NewRequest(http.MethodGet, "/1", nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

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

func TestList(t *testing.T) {
	ctrl, serviceMock, _, h := setup(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		mockSetup     func()
		expectedCode  int
		expectedError string
		expectedData  []dto.Subcategory
	}{
		{
			name: "Success",
			mockSetup: func() {
				serviceMock.EXPECT().List(gomock.Any(), page, perPage).Return([]entity.Subcategory{
					subcategoryEntity,
					subcategoryEntity,
				}, nil)
			},
			expectedCode: http.StatusOK,
			expectedData: []dto.Subcategory{
				subcategoryDTO,
				subcategoryDTO,
			},
		},
		{
			name: "Service Error",
			mockSetup: func() {
				serviceMock.EXPECT().List(gomock.Any(), page, perPage).Return(nil, errors.New("internal error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: rest.MsgInternalServerError,
		},
		{
			name: "No Subcategories",
			mockSetup: func() {
				serviceMock.EXPECT().List(gomock.Any(), page, perPage).Return([]entity.Subcategory{}, nil)
			},
			expectedCode: http.StatusOK,
			expectedData: []dto.Subcategory{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			r := chi.NewRouter()
			r.Get("/", h.List)

			q := make(url.Values)
			q.Set("page", "1")
			q.Set("per_page", "10")
			req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code)

			log.Print(rec.Body.String())

			if tt.expectedError != "" {
				var errResp rest.ErrorResponse
				if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
					assert.Equal(t, tt.expectedError, errResp.Message)
				}
			} else {
				var response rest.ApiResponse[[]dto.Subcategory]
				if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&response)) {
					assert.Equal(t, tt.expectedData, response.Data)
				}
			}
		})
	}
}

func TestListByCategoryId(t *testing.T) {
	ctrl, serviceMock, _, h := setup(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		mockSetup     func()
		expectedCode  int
		expectedError string
		expectedData  []dto.Subcategory
	}{
		{
			name: "Success",
			mockSetup: func() {
				serviceMock.EXPECT().ListByCategoryId(gomock.Any(), id, page, perPage).Return([]entity.Subcategory{
					subcategoryEntity,
					subcategoryEntity,
				}, nil)
			},
			expectedCode: http.StatusOK,
			expectedData: []dto.Subcategory{
				subcategoryDTO,
				subcategoryDTO,
			},
		},
		{
			name:          "Invalid Category ID",
			mockSetup:     func() {},
			expectedCode:  http.StatusBadRequest,
			expectedError: rest.MsgInvalidId,
		},
		{
			name: "Service Error",
			mockSetup: func() {
				serviceMock.EXPECT().ListByCategoryId(gomock.Any(), id, page, perPage).Return(nil, errors.New("internal error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: rest.MsgInternalServerError,
		},
		{
			name: "No Subcategories",
			mockSetup: func() {
				serviceMock.EXPECT().ListByCategoryId(gomock.Any(), id, page, perPage).Return([]entity.Subcategory{}, nil)
			},
			expectedCode: http.StatusOK,
			expectedData: []dto.Subcategory{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			r := chi.NewRouter()
			r.Get("/{category_id}", h.ListByCategoryId)

			var u string
			if tt.name == "Invalid Category ID" {
				u = "/abc"
			} else {
				q := make(url.Values)
				q.Set("page", "1")
				q.Set("per_page", "10")

				u = "/1?" + q.Encode()
			}
			req := httptest.NewRequest(http.MethodGet, u, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code)

			if tt.expectedError != "" {
				var errResp rest.ErrorResponse
				if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
					assert.Equal(t, tt.expectedError, errResp.Message)
				}
			} else {
				var response rest.ApiResponse[[]dto.Subcategory]
				if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&response)) {
					assert.Equal(t, tt.expectedData, response.Data)
				}
			}
		})
	}
}

func TestListByTypeId(t *testing.T) {
	ctrl, serviceMock, _, h := setup(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		mockSetup     func()
		expectedCode  int
		expectedError string
		expectedData  []dto.Subcategory
	}{
		{
			name: "Success",
			mockSetup: func() {
				serviceMock.EXPECT().ListByTypeId(gomock.Any(), id, page, perPage).Return([]entity.Subcategory{
					subcategoryEntity,
					subcategoryEntity,
				}, nil)
			},
			expectedCode: http.StatusOK,
			expectedData: []dto.Subcategory{
				subcategoryDTO,
				subcategoryDTO,
			},
		},
		{
			name:          "Invalid Category ID",
			mockSetup:     func() {},
			expectedCode:  http.StatusBadRequest,
			expectedError: rest.MsgInvalidId,
		},
		{
			name: "Service Error",
			mockSetup: func() {
				serviceMock.EXPECT().ListByTypeId(gomock.Any(), id, page, perPage).Return(nil, errors.New("internal error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: rest.MsgInternalServerError,
		},
		{
			name: "No Subcategories",
			mockSetup: func() {
				serviceMock.EXPECT().ListByTypeId(gomock.Any(), id, page, perPage).Return([]entity.Subcategory{}, nil)
			},
			expectedCode: http.StatusOK,
			expectedData: []dto.Subcategory{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			r := chi.NewRouter()
			r.Get("/{type_id}", h.ListByTypeId)

			var u string
			if tt.name == "Invalid Category ID" {
				u = "/abc"
			} else {
				q := make(url.Values)
				q.Set("page", "1")
				q.Set("per_page", "10")

				u = "/1?" + q.Encode()
			}
			req := httptest.NewRequest(http.MethodGet, u, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code)

			if tt.expectedError != "" {
				var errResp rest.ErrorResponse
				if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
					assert.Equal(t, tt.expectedError, errResp.Message)
				}
			} else {
				var response rest.ApiResponse[[]dto.Subcategory]
				if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&response)) {
					assert.Equal(t, tt.expectedData, response.Data)
				}
			}
		})
	}
}

// TODO: Add update and delete tests