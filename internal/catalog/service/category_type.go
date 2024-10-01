package service

import (
	"context"

	"github.com/hexley21/fixup/internal/catalog/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/catalog/delivery/http/v1/dto/mapper"
	"github.com/hexley21/fixup/internal/catalog/repository"
)

type CategoryTypeService interface {
	CreateCategoryType(ctx context.Context, dto dto.CreateCategoryTypeDTO) (categoryType dto.CategoryTypeDTO, err error)
	DeleteCategoryTypeById(ctx context.Context, id int32) error
	GetCategoryTypeById(ctx context.Context, id int32) (dto dto.CategoryTypeDTO, err error)
	ListCategoryTypes(ctx context.Context) (dtos []dto.CategoryTypeDTO, err error)
	UpdateCategoryTypeById(ctx context.Context, id int32, dto dto.PatchCategoryTypeDTO) error
}

type categoryTypeServiceImpl struct {
	categoryTypeRepository repository.CategoryTypeRepository
}

func NewCategoryTypeService(categoryTypeRepository repository.CategoryTypeRepository) *categoryTypeServiceImpl {
	return &categoryTypeServiceImpl{
		categoryTypeRepository: categoryTypeRepository,
	}
}

func (s *categoryTypeServiceImpl) CreateCategoryType(ctx context.Context, dto dto.CreateCategoryTypeDTO) (categoryType dto.CategoryTypeDTO, err error) {
	entity, err := s.categoryTypeRepository.CreateCategoryType(ctx, dto.Name)
	if err != nil {
		return categoryType, err
	}

	return mapper.MapCategoryTypeToDTO(entity), err
}

func (s *categoryTypeServiceImpl) DeleteCategoryTypeById(ctx context.Context, id int32) error {
	return s.categoryTypeRepository.DeleteCategoryTypeById(ctx, id)
}

func (s *categoryTypeServiceImpl) GetCategoryTypeById(ctx context.Context, id int32) (dto dto.CategoryTypeDTO, err error) {
	entity, err := s.categoryTypeRepository.GetCategoryTypeById(ctx, id)
	if err != nil {
		return dto, err
	}

	return mapper.MapCategoryTypeToDTO(entity), err
}

func (s *categoryTypeServiceImpl) ListCategoryTypes(ctx context.Context) (dtos []dto.CategoryTypeDTO, err error) {
	entities, err := s.categoryTypeRepository.GetCategoryTypes(ctx)
	if err != nil {
		return dtos, nil
	}

	for _, e := range entities {
		dtos = append(dtos, mapper.MapCategoryTypeToDTO(e))
	}

	return
}

func (s *categoryTypeServiceImpl) UpdateCategoryTypeById(ctx context.Context, id int32, dto dto.PatchCategoryTypeDTO) error {
	return s.categoryTypeRepository.UpdateCategoryTypeById(ctx, repository.UpdateCategoryTypeByIdParams{ID: id, Name: dto.Name})
}
