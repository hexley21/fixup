package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hexley21/fixup/internal/catalog/domain"
	"github.com/hexley21/fixup/internal/catalog/repository"
	mock_repository "github.com/hexley21/fixup/internal/catalog/repository/mock"
	"github.com/hexley21/fixup/internal/catalog/service"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

const categoryName = "Home"

var (
	categoryModel = repository.CategoryModel{ID: id, Name: categoryName, TypeID: id}
	categoryInfoVO = domain.CategoryInfo{Name: categoryName, TypeID: id}
)

func setupCategory(t *testing.T) (
	ctrl *gomock.Controller,
	ctx context.Context,
	svc service.CategoryService,
	mockCategoryRepository *mock_repository.MockCategoryRepository,
) {
	ctrl = gomock.NewController(t)
	ctx = context.Background()

	mockCategoryRepository = mock_repository.NewMockCategoryRepository(ctrl)
	svc = service.NewCategoryService(mockCategoryRepository)

	return
}

func TestCreateCategory_Success(t *testing.T) {
	ctrl, ctx, svc, mockCategoryRepository := setupCategory(t)
	defer ctrl.Finish()

	mockCategoryRepository.EXPECT().Create(ctx, gomock.Any()).Return(categoryModel, nil)

	categoryEntity, err := svc.Create(ctx, categoryInfoVO)
	assert.NoError(t, err)
	assert.Equal(t, id, categoryEntity.ID)
	assert.Equal(t, id, categoryEntity.Info.TypeID)
	assert.Equal(t, categoryName, categoryEntity.Info.Name)
}

func TestCreateCategory_RepositoryError(t *testing.T) {
	ctrl, ctx, svc, mockCategoryRepository := setupCategory(t)
	defer ctrl.Finish()

	mockCategoryRepository.EXPECT().Create(ctx, gomock.Any()).Return(repository.CategoryModel{}, errors.New(""))

	categoryEntity, err := svc.Create(ctx, categoryInfoVO)

	assert.Error(t, err)
	assert.Empty(t, categoryEntity)
}

func TestDeleteCategoryById_Success(t *testing.T) {
	ctrl, ctx, svc, mockCategoryRepository := setupCategory(t)
	defer ctrl.Finish()

	mockCategoryRepository.EXPECT().Delete(ctx, id).Return(true, nil)

	assert.NoError(t, svc.Delete(ctx, id))
}

func TestDeleteCategoryById_RepositoryError(t *testing.T) {
	ctrl, ctx, svc, mockCategoryRepository := setupCategory(t)
	defer ctrl.Finish()

	mockCategoryRepository.EXPECT().Delete(ctx, id).Return(false, errors.New(""))

	assert.Error(t, svc.Delete(ctx, id))
}

func TestDeleteCategoryById_NotFound(t *testing.T) {
	ctrl, ctx, svc, mockCategoryRepository := setupCategory(t)
	defer ctrl.Finish()

	mockCategoryRepository.EXPECT().Delete(ctx, id).Return(false, nil)

	assert.ErrorIs(t, svc.Delete(ctx, id), service.ErrCategoryNotFound)
}

func TestGetCategoryById_Success(t *testing.T) {
	ctrl, ctx, svc, mockCategoryRepository := setupCategory(t)
	defer ctrl.Finish()

	mockCategoryRepository.EXPECT().Get(ctx, id).Return(categoryModel, nil)

	categoryEntity, err := svc.Get(ctx, id)

	assert.NoError(t, err)
	assert.Equal(t, id, categoryEntity.ID)
	assert.Equal(t, id, categoryEntity.Info.TypeID)
	assert.Equal(t, categoryName, categoryEntity.Info.Name)
}

func TestGetCategoryById_RepositoryError(t *testing.T) {
	ctrl, ctx, svc, mockCategoryRepository := setupCategory(t)
	defer ctrl.Finish()

	mockCategoryRepository.EXPECT().Get(ctx, id).Return(repository.CategoryModel{}, errors.New(""))

	categoryEntity, err := svc.Get(ctx, id)

	assert.Error(t, err)
	assert.Empty(t, categoryEntity)
}

func TestGetCategoryById_NotFound(t *testing.T) {
	ctrl, ctx, svc, mockCategoryRepository := setupCategory(t)
	defer ctrl.Finish()

	mockCategoryRepository.EXPECT().Get(ctx, id).Return(repository.CategoryModel{}, pgx.ErrNoRows)

	categoryEntity, err := svc.Get(ctx, id)

	assert.ErrorIs(t, err, service.ErrCategoryNotFound)
	assert.Empty(t, categoryEntity)
}

func TestListCategories_Success(t *testing.T) {
	ctrl, ctx, svc, mockCategoryRepository := setupCategory(t)
	defer ctrl.Finish()

	mockCategoryRepository.EXPECT().List(ctx, limit, offset).Return([]repository.CategoryModel{categoryModel, categoryModel}, nil)

	categoriesDTO, err := svc.List(ctx, limit, offset)
	assert.NoError(t, err)
	assert.Len(t, categoriesDTO, int(limit))
}

func TestListCategories_RepositoryError(t *testing.T) {
	ctrl, ctx, svc, mockCategoryRepository := setupCategory(t)
	defer ctrl.Finish()

	mockCategoryRepository.EXPECT().List(ctx, limit, offset).Return(nil, errors.New(""))

	categoriesDTO, err := svc.List(ctx, limit ,offset)
	assert.Error(t, err)
	assert.Empty(t, categoriesDTO)
}

func TestListCategoriesByTypeId_Success(t *testing.T) {
	ctrl, ctx, svc, mockCategoryRepository := setupCategory(t)
	defer ctrl.Finish()

	mockCategoryRepository.EXPECT().ListByTypeId(ctx, id, limit, offset).Return([]repository.CategoryModel{categoryModel, categoryModel}, nil)

	categoriesDTO, err := svc.ListByTypeId(ctx, id, limit, offset)
	assert.NoError(t, err)
	assert.Len(t, categoriesDTO, 2)
}

func TestListCategoriesByTypeId_RepositoryError(t *testing.T) {
	ctrl, ctx, svc, mockCategoryRepository := setupCategory(t)
	defer ctrl.Finish()

	mockCategoryRepository.EXPECT().ListByTypeId(ctx, id, limit, offset).Return(nil, errors.New(""))

	categoriesDTO, err := svc.ListByTypeId(ctx, id, limit, offset)
	assert.Error(t, err)
	assert.Empty(t, categoriesDTO)
}

func TestUpdateCategoryById_Success(t *testing.T) {
	ctrl, ctx, svc, mockCategoryRepository := setupCategory(t)
	defer ctrl.Finish()

	mockCategoryRepository.EXPECT().Update(ctx, id, categoryInfoVO).Return(categoryModel, nil)

	categoryEntity, err := svc.Update(ctx, id, categoryInfoVO)
	assert.NoError(t, err)
	assert.Equal(t, categoryName, categoryEntity.Info.Name)
	assert.Equal(t, id, categoryEntity.Info.TypeID)
	assert.Equal(t, id, categoryEntity.ID)
}

func TestUpdateCategoryById_RepositoryError(t *testing.T) {
	ctrl, ctx, svc, mockCategoryRepository := setupCategory(t)
	defer ctrl.Finish()

	mockCategoryRepository.EXPECT().Update(ctx, id, categoryInfoVO).Return(repository.CategoryModel{}, errors.New(""))

	categoryEntity, err := svc.Update(ctx, id, categoryInfoVO)
	assert.Error(t, err)
	assert.Empty(t, categoryEntity)
}

func TestUpdateCategoryById_NotFound(t *testing.T) {
	ctrl, ctx, svc, mockCategoryRepository := setupCategory(t)
	defer ctrl.Finish()

	mockCategoryRepository.EXPECT().Update(ctx, id, categoryInfoVO).Return(repository.CategoryModel{}, pgx.ErrNoRows)

	categoryEntity, err := svc.Update(ctx, id, categoryInfoVO)
	assert.ErrorIs(t, err, service.ErrCategoryNotFound)
	assert.Empty(t, categoryEntity)
}
