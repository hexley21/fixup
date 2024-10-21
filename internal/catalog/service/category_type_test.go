package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hexley21/fixup/internal/catalog/repository"
	mock_repository "github.com/hexley21/fixup/internal/catalog/repository/mock"
	"github.com/hexley21/fixup/internal/catalog/service"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

const (
	categoryTypeName = "Home"

	id     int32 = 1
	limit  int64 = 2
	offset int64 = 0
)

var (
	categoryTypeModel = repository.CategoryTypeModel{
		ID:   id,
		Name: categoryTypeName,
	}

	categoryTypeModels = []repository.CategoryTypeModel{
		{ID: 1, Name: "Category 1"},
		{ID: 2, Name: "Category 2"},
	}
)

func setupCategoryType(t *testing.T) (
	ctrl *gomock.Controller,
	ctx context.Context,
	svc service.CategoryTypeService,
	mockCategoryTypeRepo *mock_repository.MockCategoryTypeRepository,
) {
	ctrl = gomock.NewController(t)
	ctx = context.Background()

	mockCategoryTypeRepo = mock_repository.NewMockCategoryTypeRepository(ctrl)
	svc = service.NewCategoryTypeService(mockCategoryTypeRepo)

	return
}

func TestCreateCategoryType_Success(t *testing.T) {
	ctrl, ctx, svc, mockCategoryTypeRepo := setupCategoryType(t)
	defer ctrl.Finish()

	mockCategoryTypeRepo.EXPECT().Create(ctx, categoryTypeName).Return(categoryTypeModel, nil)

	result, err := svc.Create(ctx, categoryName)

	if assert.NoError(t, err) {
		assert.Equal(t, categoryTypeModel.ID, result.ID)
		assert.Equal(t, categoryTypeModel.Name, result.Name)
	}
}

func TestCreateCategoryType_RepositoryError(t *testing.T) {
	ctrl, ctx, svc, mockCategoryTypeRepo := setupCategoryType(t)
	defer ctrl.Finish()

	mockCategoryTypeRepo.EXPECT().Create(ctx, categoryTypeName).Return(repository.CategoryTypeModel{}, errors.New(""))

	result, err := svc.Create(ctx, categoryTypeName)

	if assert.EqualError(t, err, "") {
		assert.Empty(t, result.ID)
		assert.Empty(t, result.Name)
	}
}

func TestDeleteCategoryTypeById_Success(t *testing.T) {
	ctrl, ctx, svc, mockRepo := setupCategoryType(t)
	defer ctrl.Finish()

	mockRepo.EXPECT().Delete(ctx, id).Return(true, nil)

	err := svc.Delete(ctx, id)
	assert.NoError(t, err)
}

func TestDeleteCategoryTypeById_RepositoryError(t *testing.T) {
	ctrl, ctx, svc, mockRepo := setupCategoryType(t)
	defer ctrl.Finish()

	mockRepo.EXPECT().Delete(ctx, id).Return(false, errors.New(""))

	err := svc.Delete(ctx, id)
	assert.Error(t, err)
}

func TestDeleteCategoryTypeById_NotFound(t *testing.T) {
	ctrl, ctx, svc, mockRepo := setupCategoryType(t)
	defer ctrl.Finish()

	mockRepo.EXPECT().Delete(ctx, id).Return(false, nil)

	err := svc.Delete(ctx, id)
	assert.ErrorIs(t, err, service.ErrCategoryTypeNotFound)
}

func TestGetCategoryTypeById_Success(t *testing.T) {
	ctrl, ctx, svc, mockRepo := setupCategoryType(t)
	defer ctrl.Finish()

	mockRepo.EXPECT().Get(ctx, id).Return(categoryTypeModel, nil)

	result, err := svc.Get(ctx, id)

	assert.NoError(t, err)
	assert.Equal(t, categoryTypeModel.ID, result.ID)
	assert.Equal(t, categoryTypeModel.Name, result.Name)
}

func TestGetCategoryTypeById_NotFound(t *testing.T) {
	ctrl, ctx, svc, mockRepo := setupCategoryType(t)
	defer ctrl.Finish()

	mockRepo.EXPECT().Get(ctx, id).Return(repository.CategoryTypeModel{}, pgx.ErrNoRows)

	result, err := svc.Get(ctx, id)

	if assert.ErrorIs(t, err, service.ErrCategoryTypeNotFound) {
		assert.Empty(t, result.ID)
		assert.Empty(t, result.Name)
	}
}

func TestGetCategoryTypeById_RepositoryError(t *testing.T) {
	ctrl, ctx, svc, mockRepo := setupCategoryType(t)
	defer ctrl.Finish()

	mockRepo.EXPECT().Get(ctx, id).Return(repository.CategoryTypeModel{}, errors.New(""))

	result, err := svc.Get(ctx, id)

	if assert.EqualError(t, err, "") {
		assert.Empty(t, result.ID)
		assert.Empty(t, result.Name)
	}
}

func TestGetCategoryTypes_Success(t *testing.T) {
	ctrl, ctx, svc, mockRepo := setupCategoryType(t)
	defer ctrl.Finish()

	mockRepo.EXPECT().List(ctx, limit, offset).Return(categoryTypeModels, nil)

	result, err := svc.List(ctx, limit, offset)

	if assert.NoError(t, err) {
		for i := range len(result) {
			assert.Equal(t, categoryTypeModels[i].ID, result[i].ID)
			assert.Equal(t, categoryTypeModels[i].Name, result[i].Name)
		}
	}
}

func TestGetCategoryTypes_RepositoryError(t *testing.T) {
	ctrl, ctx, svc, mockRepo := setupCategoryType(t)
	defer ctrl.Finish()

	mockRepo.EXPECT().List(ctx, limit, offset).Return(nil, errors.New(""))

	result, err := svc.List(ctx, limit, offset)

	if assert.EqualError(t, err, "") {
		assert.Empty(t, result)
	}
}

func TestUpdateCategoryTypeById_Success(t *testing.T) {
	ctrl, ctx, svc, mockRepo := setupCategoryType(t)
	defer ctrl.Finish()

	mockRepo.EXPECT().Update(ctx, id, categoryTypeName).Return(true, nil)

	assert.NoError(t, svc.Update(ctx, id, categoryTypeName))
}

func TestUpdateCategoryTypeById_RepositoryError(t *testing.T) {
	ctrl, ctx, svc, mockRepo := setupCategoryType(t)
	defer ctrl.Finish()

	mockRepo.EXPECT().Update(ctx, id, categoryTypeName).Return(false, errors.New(""))

	assert.Error(t, svc.Update(ctx, id, categoryTypeName))
}

func TestUpdateCategoryTypeById_NotFound(t *testing.T) {
	ctrl, ctx, svc, mockRepo := setupCategoryType(t)
	defer ctrl.Finish()

	mockRepo.EXPECT().Update(ctx, id, categoryTypeName).Return(false, nil)

	assert.ErrorIs(t, svc.Update(ctx, id, categoryTypeName), service.ErrCategoryTypeNotFound)
}
