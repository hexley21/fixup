package service_test

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"github.com/hexley21/fixup/internal/catalog/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/catalog/entity"
	"github.com/hexley21/fixup/internal/catalog/repository"
	mock_repository "github.com/hexley21/fixup/internal/catalog/repository/mock"
	"github.com/hexley21/fixup/internal/catalog/service"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

const (
	categoryName = "Home"
	id = int32(1)

	page     int32 = 1
    per_page int32 = 2
)
var (
	categoryTypeEntity = entity.CategoryType{
		ID: id,
		Name: categoryName,
	}
	createCategoryTypeDTO = dto.CreateCategoryTypeDTO{
		Name: categoryName,
	}

	categoryTypeEntities = []entity.CategoryType{
        {ID: 1, Name: "Category 1"},
        {ID: 2, Name: "Category 2"},
    }

	patchCategoryTypeDTO = dto.PatchCategoryTypeDTO{
		Name: categoryName,
	}
)

func setup(t *testing.T) (
	ctrl *gomock.Controller,
	ctx context.Context,
	svc service.CategoryTypeService,
	mockCategoryTypeRepo *mock_repository.MockCategoryTypeRepository,
){
	ctrl = gomock.NewController(t)
	ctx = context.Background()

	mockCategoryTypeRepo = mock_repository.NewMockCategoryTypeRepository(ctrl)
	svc = service.NewCategoryTypeService(mockCategoryTypeRepo)

	return
}

func TestCreateCategoryType_Success(t *testing.T) {
    ctrl, ctx, svc, mockCategoryTypeRepo := setup(t)
	defer ctrl.Finish()

    mockCategoryTypeRepo.EXPECT().CreateCategoryType(ctx, categoryName).Return(categoryTypeEntity, nil)

    result, err := svc.CreateCategoryType(ctx, createCategoryTypeDTO)

    assert.Equal(t, strconv.Itoa(int(categoryTypeEntity.ID)), result.ID)
	assert.Equal(t, categoryTypeEntity.Name, result.Name)
    assert.NoError(t, err)
}

func TestCreateCategoryType_RepositoryError(t *testing.T) {
    ctrl, ctx, svc, mockCategoryTypeRepo := setup(t)
    defer ctrl.Finish()

    mockCategoryTypeRepo.EXPECT().CreateCategoryType(ctx, categoryName).Return(entity.CategoryType{}, errors.New(""))

    result, err := svc.CreateCategoryType(ctx, createCategoryTypeDTO)

    assert.Empty(t, result.ID)
    assert.Empty(t, result.Name)
    assert.EqualError(t, err, "")
}

func TestDeleteCategoryTypeById_Success(t *testing.T) {
    ctrl, ctx, svc, mockRepo := setup(t)
    defer ctrl.Finish()

    mockRepo.EXPECT().DeleteCategoryTypeById(ctx, id).Return(nil)

    err := svc.DeleteCategoryTypeById(ctx, id)
    assert.NoError(t, err)
}

func TestDeleteCategoryTypeById_RepositoryError(t *testing.T) {
    ctrl, ctx, svc, mockRepo := setup(t)
    defer ctrl.Finish()

    mockRepo.EXPECT().DeleteCategoryTypeById(ctx, id).Return(errors.New(""))

    err := svc.DeleteCategoryTypeById(ctx, id)
    assert.EqualError(t, err, "")
}


func TestGetCategoryTypeById_Success(t *testing.T) {
    ctrl, ctx, svc, mockRepo := setup(t)
    defer ctrl.Finish()

    mockRepo.EXPECT().GetCategoryTypeById(ctx, id).Return(categoryTypeEntity, nil)

    result, err := svc.GetCategoryTypeById(ctx, id)

    assert.NoError(t, err)
    assert.Equal(t, strconv.Itoa(int(categoryTypeEntity.ID)), result.ID)
	assert.Equal(t, categoryTypeEntity.Name, result.Name)
}

func TestGetCategoryTypeById_RepositoryError(t *testing.T) {
    ctrl, ctx, svc, mockRepo := setup(t)
    defer ctrl.Finish()

    mockRepo.EXPECT().GetCategoryTypeById(ctx, id).Return(entity.CategoryType{}, errors.New(""))

    result, err := svc.GetCategoryTypeById(ctx, id)

    assert.Empty(t, result.ID)
    assert.Empty(t, result.Name)
    assert.EqualError(t, err, "")
}

func TestGetCategoryTypes_Success(t *testing.T) {
    ctrl, ctx, svc, mockRepo := setup(t)
    defer ctrl.Finish()

    mockRepo.EXPECT().GetCategoryTypes(ctx, per_page*(page-1), per_page).Return(categoryTypeEntities, nil)

    result, err := svc.GetCategoryTypes(ctx, page, per_page)

    assert.NoError(t, err)
    for i := range len(result) {
		assert.Equal(t, strconv.Itoa(int(categoryTypeEntities[i].ID)), result[i].ID)
		assert.Equal(t, categoryTypeEntities[i].Name, result[i].Name)
	}
}

func TestGetCategoryTypes_RepositoryError(t *testing.T) {
    ctrl, ctx, svc, mockRepo := setup(t)
    defer ctrl.Finish()

    mockRepo.EXPECT().GetCategoryTypes(ctx, per_page*(page-1), per_page).Return(nil, errors.New(""))

    result, err := svc.GetCategoryTypes(ctx, page, per_page)

    assert.Empty(t, result)
    assert.EqualError(t, err, "")
}

func TestUpdateCategoryTypeById_Success(t *testing.T) {
    ctrl, ctx, svc, mockRepo := setup(t)
    defer ctrl.Finish()

    mockRepo.EXPECT().UpdateCategoryTypeById(ctx, repository.UpdateCategoryTypeByIdParams{ID: id, Name: categoryName}).Return(nil)

    err := svc.UpdateCategoryTypeById(ctx, id, patchCategoryTypeDTO)

    assert.NoError(t, err)
}

func TestUpdateCategoryTypeById_RepositoryError(t *testing.T) {
    ctrl, ctx, svc, mockRepo := setup(t)
    defer ctrl.Finish()

    mockRepo.EXPECT().UpdateCategoryTypeById(ctx, repository.UpdateCategoryTypeByIdParams{ID: id, Name: categoryName}).Return(errors.New(""))

    err := svc.UpdateCategoryTypeById(ctx, id, patchCategoryTypeDTO)

    assert.EqualError(t, err, "")
}