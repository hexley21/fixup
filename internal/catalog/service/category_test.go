package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hexley21/fixup/internal/catalog/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/catalog/entity"
	mockRepository "github.com/hexley21/fixup/internal/catalog/repository/mock"
	"github.com/hexley21/fixup/internal/catalog/service"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

const (
	categoryName = "Home"
)

var (
	categoryEntity = entity.Category{ID: id, Name: categoryName, TypeID: id}

	createCategoryDTO = dto.CreateCategoryDTO{Name: categoryName, TypeID: strId}
	patchCategoryDTO  = dto.PatchCategoryDTO{Name: categoryName, TypeID: strId}
)

func setupCategory(t *testing.T) (
	ctrl *gomock.Controller,
	ctx context.Context,
	svc service.CategoryService,
	categoryRepoMock *mockRepository.MockCategoryRepository,
) {
	ctrl = gomock.NewController(t)
	ctx = context.Background()

	categoryRepoMock = mockRepository.NewMockCategoryRepository(ctrl)
	svc = service.NewCategoryService(categoryRepoMock, 50, 100)

	return
}

func TestCreateCategory_Success(t *testing.T) {
	ctrl, ctx, svc, categoryRepoMock := setupCategory(t)
	defer ctrl.Finish()

	categoryRepoMock.EXPECT().CreateCategory(ctx, gomock.Any()).Return(categoryEntity, nil)

	categoryDTO, err := svc.CreateCategory(ctx, createCategoryDTO)
	assert.NoError(t, err)
	assert.Equal(t, strId, categoryDTO.ID)
	assert.Equal(t, strId, categoryDTO.TypeID)
	assert.Equal(t, categoryName, categoryDTO.Name)
}

func TestCreateCategory_InvalidId(t *testing.T) {
	ctrl, ctx, svc, _ := setupCategory(t)
	defer ctrl.Finish()

	categoryDTO, err := svc.CreateCategory(ctx, dto.CreateCategoryDTO{TypeID: "abc"})

	assert.Error(t, err)
	assert.Empty(t, categoryDTO)
}

func TestCreateCategory_RepositoryError(t *testing.T) {
	ctrl, ctx, svc, categoryRepoMock := setupCategory(t)
	defer ctrl.Finish()

	categoryRepoMock.EXPECT().CreateCategory(ctx, gomock.Any()).Return(entity.Category{}, errors.New(""))

	categoryDTO, err := svc.CreateCategory(ctx, createCategoryDTO)

	assert.Error(t, err)
	assert.Empty(t, categoryDTO)
}

func TestDeleteCategoryById_Success(t *testing.T) {
	ctrl, ctx, svc, categoryRepoMock := setupCategory(t)
	defer ctrl.Finish()

	categoryRepoMock.EXPECT().DeleteCategoryById(ctx, id).Return(nil)

	assert.NoError(t, svc.DeleteCategoryById(ctx, id))
}

func TestDeleteCategoryById_RepositoryError(t *testing.T) {
	ctrl, ctx, svc, categoryRepoMock := setupCategory(t)
	defer ctrl.Finish()

	categoryRepoMock.EXPECT().DeleteCategoryById(ctx, id).Return(errors.New(""))

	assert.Error(t, svc.DeleteCategoryById(ctx, id))
}

func TestGetCategoryById_Success(t *testing.T) {
	ctrl, ctx, svc, categoryRepoMock := setupCategory(t)
	defer ctrl.Finish()

	categoryRepoMock.EXPECT().GetCategoryById(ctx, id).Return(categoryEntity, nil)

	categoryDTO, err := svc.GetCategoryById(ctx, id)

	assert.NoError(t, err)
	assert.Equal(t, strId, categoryDTO.ID)
	assert.Equal(t, strId, categoryDTO.TypeID)
	assert.Equal(t, categoryName, categoryDTO.Name)
}

func TestGetCategoryById_RepositoryError(t *testing.T) {
	ctrl, ctx, svc, categoryRepoMock := setupCategory(t)
	defer ctrl.Finish()

	categoryRepoMock.EXPECT().GetCategoryById(ctx, id).Return(entity.Category{}, errors.New(""))

	categoryDTO, err := svc.GetCategoryById(ctx, id)

	assert.Error(t, err)
	assert.Empty(t, categoryDTO)
}

func TestGetCategories_Success(t *testing.T) {
	ctrl, ctx, svc, categoryRepoMock := setupCategory(t)
	defer ctrl.Finish()

	categoryRepoMock.EXPECT().GetCategories(ctx, perPage*(page-1), perPage).Return([]entity.Category{categoryEntity, categoryEntity}, nil)

	categoriesDTO, err := svc.GetCategories(ctx, page, perPage)
	assert.NoError(t, err)
	assert.Len(t, categoriesDTO, int(perPage))
}

func TestGetCategories_RepositoryError(t *testing.T) {
	ctrl, ctx, svc, categoryRepoMock := setupCategory(t)
	defer ctrl.Finish()

	categoryRepoMock.EXPECT().GetCategories(ctx, perPage*(page-1), perPage).Return(nil, errors.New(""))

	categoriesDTO, err := svc.GetCategories(ctx, page, perPage)
	assert.Error(t, err)
	assert.Empty(t, categoriesDTO)
}

func TestGetCategoriesByTypeId_Success(t *testing.T) {
	ctrl, ctx, svc, categoryRepoMock := setupCategory(t)
	defer ctrl.Finish()

	categoryRepoMock.EXPECT().GetCategoriesByTypeId(ctx, id, perPage*(page-1), perPage).Return([]entity.Category{categoryEntity, categoryEntity}, nil)

	categoriesDTO, err := svc.GetCategoriesByTypeId(ctx, id, page, perPage)
	assert.NoError(t, err)
	assert.Len(t, categoriesDTO, 2)
}

func TestGetCategoriesByTypeId_RepositoryError(t *testing.T) {
	ctrl, ctx, svc, categoryRepoMock := setupCategory(t)
	defer ctrl.Finish()

	categoryRepoMock.EXPECT().GetCategoriesByTypeId(ctx, id, perPage*(page-1), perPage).Return(nil, errors.New(""))

	categoriesDTO, err := svc.GetCategoriesByTypeId(ctx, id, page, perPage)
	assert.Error(t, err)
	assert.Empty(t, categoriesDTO)
}

func TestUpdateCategoryById_Success(t *testing.T) {
	ctrl, ctx, svc, categoryRepoMock := setupCategory(t)
	defer ctrl.Finish()

	categoryRepoMock.EXPECT().UpdateCategoryById(ctx, gomock.Any()).Return(categoryEntity, nil)

	categoryDTO, err := svc.UpdateCategoryById(ctx, id, patchCategoryDTO)
	assert.NoError(t, err)
	assert.Equal(t, categoryName, categoryDTO.Name)
	assert.Equal(t, strId, categoryDTO.TypeID)
	assert.Equal(t, strId, categoryDTO.ID)
}

func TestUpdateCategoryById_RepositoryError(t *testing.T) {
	ctrl, ctx, svc, categoryRepoMock := setupCategory(t)
	defer ctrl.Finish()

	categoryRepoMock.EXPECT().UpdateCategoryById(ctx, gomock.Any()).Return(entity.Category{}, errors.New(""))

	categoryDTO, err := svc.UpdateCategoryById(ctx, id, patchCategoryDTO)
	assert.Error(t, err)
	assert.Empty(t, categoryDTO)
}
