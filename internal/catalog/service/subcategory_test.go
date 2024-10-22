package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hexley21/fixup/internal/catalog/domain"
	"github.com/hexley21/fixup/internal/catalog/repository"
	mock_repository "github.com/hexley21/fixup/internal/catalog/repository/mock"
	"github.com/hexley21/fixup/internal/catalog/service"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func setupSubcategory(t *testing.T) (
	ctrl *gomock.Controller,
	ctx context.Context,
	mockSubcategoryRepo *mock_repository.MockSubcategory,
	svc service.SubcategoryService,

) {
	ctrl = gomock.NewController(t)
	ctx = context.Background()

	mockSubcategoryRepo = mock_repository.NewMockSubcategory(ctrl)
	svc = service.NewSubcategoryService(mockSubcategoryRepo)

	return
}

func TestGetSubcategory(t *testing.T) {
	ctrl, ctx, mockSubcategoryRepo, svc := setupSubcategory(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		id            int32
		mockReturn    repository.SubcategoryModel
		mockError     error
		expectedError error
	}{
		{
			name:       "Success",
			id:         1,
			mockReturn: repository.SubcategoryModel{ID: 1, CategoryID: 1, Name: "Test Subcategory"},
		},
		{
			name:          "NotFound",
			id:            2,
			mockReturn:    repository.SubcategoryModel{},
			mockError:     pgx.ErrNoRows,
			expectedError: service.ErrSubcategoryNotFound,
		},
		{
			name:          "OtherError",
			id:            3,
			mockReturn:    repository.SubcategoryModel{},
			mockError:     errors.New("some error"),
			expectedError: errors.New("failed to fetch subcategory: some error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSubcategoryRepo.EXPECT().Get(ctx, tt.id).Return(tt.mockReturn, tt.mockError)

			result, err := svc.Get(ctx, tt.id)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, tt.expectedError, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockReturn.ID, result.ID)
				assert.Equal(t, tt.mockReturn.CategoryID, result.Info.CategoryID)
				assert.Equal(t, tt.mockReturn.Name, result.Info.Name)
			}
		})
	}
}

func TestListSubcategories(t *testing.T) {
	ctrl, ctx, mockSubcategoryRepo, svc := setupSubcategory(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		limit         int64
		offset        int64
		mockReturn    []repository.SubcategoryModel
		mockError     error
		expectedError error
	}{
		{
			name:       "Success",
			limit:      10,
			offset:     0,
			mockReturn: []repository.SubcategoryModel{{ID: 1, CategoryID: 1, Name: "Test Subcategory"}},
		},
		{
			name:          "EmptyList",
			limit:         10,
			offset:        0,
			mockReturn:    []repository.SubcategoryModel{},
			expectedError: nil,
		},
		{
			name:          "OtherError",
			limit:         10,
			offset:        0,
			mockReturn:    nil,
			mockError:     errors.New("some error"),
			expectedError: errors.New("failed to fetch subcategories: some error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSubcategoryRepo.EXPECT().List(ctx, tt.limit, tt.offset).Return(tt.mockReturn, tt.mockError)

			result, err := svc.List(ctx, tt.limit, tt.offset)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, tt.expectedError, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.mockReturn), len(result))
			}
		})
	}
}

func TestListByCategoryId(t *testing.T) {
	ctrl, ctx, mockSubcategoryRepo, svc := setupSubcategory(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		categoryID    int32
		limit         int64
		offset        int64
		mockReturn    []repository.SubcategoryModel
		mockError     error
		expectedError error
	}{
		{
			name:       "Success",
			categoryID: 1,
			limit:      10,
			offset:     0,
			mockReturn: []repository.SubcategoryModel{{ID: 1, CategoryID: 1, Name: "Test Subcategory"}},
		},
		{
			name:       "EmptyList",
			categoryID: 1,
			limit:      10,
			offset:     0,
			mockReturn: []repository.SubcategoryModel{},
		},
		{
			name:          "OtherError",
			categoryID:    1,
			limit:         10,
			offset:        0,
			mockReturn:    nil,
			mockError:     errors.New("some error"),
			expectedError: errors.New("failed to fetch subcategories: some error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSubcategoryRepo.EXPECT().ListByCategoryId(ctx, tt.categoryID, tt.limit, tt.offset).Return(tt.mockReturn, tt.mockError)

			result, err := svc.ListByCategoryId(ctx, tt.categoryID, tt.limit, tt.offset)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, tt.expectedError, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.mockReturn), len(result))
			}
		})
	}
}

func TestListByTypeId(t *testing.T) {
	ctrl, ctx, mockSubcategoryRepo, svc := setupSubcategory(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		typeID        int32
		limit         int64
		offset        int64
		mockReturn    []repository.SubcategoryModel
		mockError     error
		expectedError error
	}{
		{
			name:       "Success",
			typeID:     1,
			limit:      10,
			offset:     0,
			mockReturn: []repository.SubcategoryModel{{ID: 1, CategoryID: 1, Name: "Test Subcategory"}},
		},
		{
			name:       "EmptyList",
			typeID:     1,
			limit:      10,
			offset:     0,
			mockReturn: []repository.SubcategoryModel{},
		},
		{
			name:          "OtherError",
			typeID:        1,
			limit:         10,
			offset:        0,
			mockReturn:    nil,
			mockError:     errors.New("some error"),
			expectedError: errors.New("failed to fetch subcategories: some error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSubcategoryRepo.EXPECT().ListByTypeId(ctx, tt.typeID, tt.limit, tt.offset).Return(tt.mockReturn, tt.mockError)

			result, err := svc.ListByTypeId(ctx, tt.typeID, tt.limit, tt.offset)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, tt.expectedError, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.mockReturn), len(result))
			}
		})
	}
}

func TestCreateSubcategory(t *testing.T) {
	ctrl, ctx, mockSubcategoryRepo, svc := setupSubcategory(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		info          domain.SubcategoryInfo
		mockReturnID  int32
		mockError     error
		expectedID    int32
		expectedError error
	}{
		{
			name:         "Success",
			info:         domain.SubcategoryInfo{Name: "Test Subcategory"},
			mockReturnID: 1,
			expectedID:   1,
		},
		{
			name:          "NameTaken",
			info:          domain.SubcategoryInfo{Name: "Duplicate Subcategory"},
			mockError:     &pgconn.PgError{Code: pgerrcode.RaiseException},
			expectedError: service.ErrSubcategoryNameTaken,
		},
		{
			name:          "OtherError",
			info:          domain.SubcategoryInfo{Name: "Another Subcategory"},
			mockError:     errors.New("some error"),
			expectedError: errors.New("failed to create subcategory: some error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSubcategoryRepo.EXPECT().Create(ctx, tt.info).Return(tt.mockReturnID, tt.mockError)

			resultID, err := svc.Create(ctx, tt.info)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, tt.expectedError, err.Error())
				assert.Equal(t, int32(0), resultID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, resultID)
			}
		})
	}
}

func TestUpdateSubcategory(t *testing.T) {
	ctrl, ctx, mockSubcategoryRepo, svc := setupSubcategory(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		id            int32
		info          domain.SubcategoryInfo
		mockReturn    repository.SubcategoryModel
		mockError     error
		expectedError error
	}{
		{
			name:       "Success",
			id:         1,
			info:       domain.SubcategoryInfo{Name: "Updated Subcategory"},
			mockReturn: repository.SubcategoryModel{ID: 1, CategoryID: 1, Name: "Updated Subcategory"},
		},
		{
			name:          "NotFound",
			id:            2,
			info:          domain.SubcategoryInfo{Name: "Non-existent Subcategory"},
			mockError:     pgx.ErrNoRows,
			expectedError: service.ErrSubcategoryNotFound,
		},
		{
			name:          "NameTaken",
			id:            3,
			info:          domain.SubcategoryInfo{Name: "Duplicate Subcategory"},
			mockError:     &pgconn.PgError{Code: pgerrcode.RaiseException},
			expectedError: service.ErrSubcategoryNameTaken,
		},
		{
			name:          "OtherError",
			id:            4,
			info:          domain.SubcategoryInfo{Name: "Another Subcategory"},
			mockError:     errors.New("some error"),
			expectedError: errors.New("failed to update subcategory: some error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSubcategoryRepo.EXPECT().Update(ctx, tt.id, tt.info).Return(tt.mockReturn, tt.mockError)

			result, err := svc.Update(ctx, tt.id, tt.info)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, tt.expectedError, err.Error())
				assert.Equal(t, domain.Subcategory{}, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockReturn.CategoryID, result.Info.CategoryID)
				assert.Equal(t, tt.mockReturn.ID, result.ID)
				assert.Equal(t, tt.mockReturn.Name, result.Info.Name)
			}
		})
	}
}

func TestDeleteSubcategory(t *testing.T) {
	ctrl, ctx, mockSubcategoryRepo, svc := setupSubcategory(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		id            int32
		mockError     error
		mockOk        bool
		expectedError error
	}{
		{
			name:          "Success",
			id:            1,
			mockError:     nil,
			expectedError: nil,
			mockOk:        true,
		},
		{
			name:          "NotFound",
			id:            2,
			mockError:     nil,
			mockOk:        false,
			expectedError: service.ErrSubcategoryNotFound,
		},
		{
			name:          "OtherError",
			id:            3,
			mockError:     errors.New("some error"),
			mockOk:        true,
			expectedError: errors.New("failed to delete subcategory: some error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSubcategoryRepo.EXPECT().Delete(ctx, tt.id).Return(tt.mockOk, tt.mockError)

			err := svc.Delete(ctx, tt.id)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, tt.expectedError, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
