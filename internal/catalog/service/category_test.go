package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hexley21/fixup/internal/catalog/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/catalog/entity"
	mock_repository "github.com/hexley21/fixup/internal/catalog/repository/mock"
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
	mockCategoryRepo *mock_repository.MockCategoryRepository,
) {
	ctrl = gomock.NewController(t)
	ctx = context.Background()

	mockCategoryRepo = mock_repository.NewMockCategoryRepository(ctrl)
	svc = service.NewCategoryService(mockCategoryRepo, 50, 100)

	return
}

func TestCreateCategory_Success(t *testing.T) {
	ctrl, ctx, svc, mockCategoryRepo := setupCategory(t)
	defer ctrl.Finish()

	mockCategoryRepo.EXPECT().CreateCategory(ctx, gomock.Any()).Return(categoryEntity, nil)

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
	ctrl, ctx, svc, mockCategoryRepo := setupCategory(t)
	defer ctrl.Finish()

	mockCategoryRepo.EXPECT().CreateCategory(ctx, gomock.Any()).Return(entity.Category{}, errors.New(""))

	categoryDTO, err := svc.CreateCategory(ctx, createCategoryDTO)

	assert.Error(t, err)
	assert.Empty(t, categoryDTO)
}

func TestDeleteCategoryById_Success(t *testing.T) {
	ctrl, ctx, svc, mockCategoryRepo := setupCategory(t)
	defer ctrl.Finish()

	mockCategoryRepo.EXPECT().DeleteCategoryById(ctx, id).Return(nil)

	assert.NoError(t, svc.DeleteCategoryById(ctx, id))
}

func TestDeleteCategoryById_RepositoryError(t *testing.T) {
	ctrl, ctx, svc, mockCategoryRepo := setupCategory(t)
	defer ctrl.Finish()

	mockCategoryRepo.EXPECT().DeleteCategoryById(ctx, id).Return(errors.New(""))

	assert.Error(t, svc.DeleteCategoryById(ctx, id))
}

func TestGetCategoryById_Success(t *testing.T) {
	ctrl, ctx, svc, mockCategoryRepo := setupCategory(t)
	defer ctrl.Finish()

	mockCategoryRepo.EXPECT().GetCategoryById(ctx, id).Return(categoryEntity, nil)

	categoryDTO, err := svc.GetCategoryById(ctx, id)

	assert.NoError(t, err)
	assert.Equal(t, strId, categoryDTO.ID)
	assert.Equal(t, strId, categoryDTO.TypeID)
	assert.Equal(t, categoryName, categoryDTO.Name)
}

func TestGetCategoryById_RepositoryError(t *testing.T) {
	ctrl, ctx, svc, mockCategoryRepo := setupCategory(t)
	defer ctrl.Finish()

	mockCategoryRepo.EXPECT().GetCategoryById(ctx, id).Return(entity.Category{}, errors.New(""))

	categoryDTO, err := svc.GetCategoryById(ctx, id)

	assert.Error(t, err)
	assert.Empty(t, categoryDTO)
}

func TestGetCategories_Success(t *testing.T) {
	ctrl, ctx, svc, mockCategoryRepo := setupCategory(t)
	defer ctrl.Finish()

	mockCategoryRepo.EXPECT().GetCategories(ctx, per_page*(page-1), per_page).Return([]entity.Category{categoryEntity, categoryEntity}, nil)

	dtos, err := svc.GetCategories(ctx, page, per_page)
	assert.NoError(t, err)
	assert.Len(t, dtos, int(per_page))
}

func TestGetCategories_RepositoryError(t *testing.T) {
	ctrl, ctx, svc, mockCategoryRepo := setupCategory(t)
	defer ctrl.Finish()

	mockCategoryRepo.EXPECT().GetCategories(ctx, per_page*(page-1), per_page).Return(nil, errors.New(""))

	dtos, err := svc.GetCategories(ctx, page, per_page)
	assert.Error(t, err)
	assert.Empty(t, dtos)
}

func TestGetCategoriesByTypeId_Success(t *testing.T) {
	ctrl, ctx, svc, mockCategoryRepo := setupCategory(t)
	defer ctrl.Finish()

	mockCategoryRepo.EXPECT().GetCategoriesByTypeId(ctx, id, per_page*(page-1), per_page).Return([]entity.Category{categoryEntity, categoryEntity}, nil)

	dtos, err := svc.GetCategoriesByTypeId(ctx, id, page, per_page)
	assert.NoError(t, err)
	assert.Len(t, dtos, 2)
}

func TestGetCategoriesByTypeId_RepositoryError(t *testing.T) {
	ctrl, ctx, svc, mockCategoryRepo := setupCategory(t)
	defer ctrl.Finish()

	mockCategoryRepo.EXPECT().GetCategoriesByTypeId(ctx, id, per_page*(page-1), per_page).Return(nil, errors.New(""))

	dtos, err := svc.GetCategoriesByTypeId(ctx, id, page, per_page)
	assert.Error(t, err)
	assert.Empty(t, dtos)
}

func TestUpdateCategoryById_Success(t *testing.T) {
	ctrl, ctx, svc, mockCategoryRepo := setupCategory(t)
	defer ctrl.Finish()

	mockCategoryRepo.EXPECT().UpdateCategoryById(ctx, gomock.Any()).Return(categoryEntity, nil)

	category, err := svc.UpdateCategoryById(ctx, id, patchCategoryDTO)
	assert.NoError(t, err)
	assert.Equal(t, categoryName, category.Name)
	assert.Equal(t, strId, category.TypeID)
	assert.Equal(t, strId, category.ID)
}

func TestUpdateCategoryById_RepositoryError(t *testing.T) {
	ctrl, ctx, svc, mockCategoryRepo := setupCategory(t)
	defer ctrl.Finish()

	mockCategoryRepo.EXPECT().UpdateCategoryById(ctx, gomock.Any()).Return(entity.Category{}, errors.New(""))

	category, err := svc.UpdateCategoryById(ctx, id, patchCategoryDTO)
	assert.Error(t, err)
	assert.Empty(t, category)
}
