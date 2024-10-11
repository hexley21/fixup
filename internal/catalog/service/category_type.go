package service

import (
	"context"

	"github.com/hexley21/fixup/internal/catalog/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/catalog/delivery/http/v1/dto/mapper"
	"github.com/hexley21/fixup/internal/catalog/repository"
)

type CategoryTypeService interface {
	CreateCategoryType(ctx context.Context, createDTO dto.CreateCategoryTypeDTO) (dto.CategoryTypeDTO, error)
	DeleteCategoryTypeById(ctx context.Context, id int32) error
	GetCategoryTypeById(ctx context.Context, id int32) (dto.CategoryTypeDTO, error)
	GetCategoryTypes(ctx context.Context, page int32, perPage int32) ([]dto.CategoryTypeDTO, error)
	UpdateCategoryTypeById(ctx context.Context, id int32, patchDTO dto.PatchCategoryTypeDTO) error
}

type categoryTypeServiceImpl struct {
	categoryTypeRepository repository.CategoryTypeRepository
	defaultPerPage         int32
	maxPerPage             int32
}

func NewCategoryTypeService(categoryTypeRepository repository.CategoryTypeRepository, defaultPerPage int32, maxPerPage int32) *categoryTypeServiceImpl {
	return &categoryTypeServiceImpl{
		categoryTypeRepository: categoryTypeRepository,
		defaultPerPage:         defaultPerPage,
		maxPerPage:             maxPerPage,
	}
}

func (s *categoryTypeServiceImpl) CreateCategoryType(ctx context.Context, createDTO dto.CreateCategoryTypeDTO) (dto.CategoryTypeDTO, error) {
	entity, err := s.categoryTypeRepository.CreateCategoryType(ctx, createDTO.Name)
	if err != nil {
		return dto.CategoryTypeDTO{}, err
	}

	return mapper.MapCategoryTypeToDTO(entity), nil
}

func (s *categoryTypeServiceImpl) DeleteCategoryTypeById(ctx context.Context, id int32) error {
	return s.categoryTypeRepository.DeleteCategoryTypeById(ctx, id)
}

func (s *categoryTypeServiceImpl) GetCategoryTypeById(ctx context.Context, id int32) (dto.CategoryTypeDTO, error) {
	entity, err := s.categoryTypeRepository.GetCategoryTypeById(ctx, id)
	if err != nil {
		return dto.CategoryTypeDTO{}, err
	}

	return mapper.MapCategoryTypeToDTO(entity), nil
}

func (s *categoryTypeServiceImpl) GetCategoryTypes(ctx context.Context, page int32, perPage int32) ([]dto.CategoryTypeDTO, error) {
	if perPage == 0 || perPage > s.maxPerPage {
		perPage = s.defaultPerPage
	}

	entities, err := s.categoryTypeRepository.GetCategoryTypes(ctx, perPage*(page-1), perPage)
	if err != nil {
		return []dto.CategoryTypeDTO{}, err
	}

	categoriesDTO := make([]dto.CategoryTypeDTO, len(entities))
	for i, e := range entities {
		categoriesDTO[i] = mapper.MapCategoryTypeToDTO(e)
	}

	return categoriesDTO, nil
}

func (s *categoryTypeServiceImpl) UpdateCategoryTypeById(ctx context.Context, id int32, patchDTO dto.PatchCategoryTypeDTO) error {
	return s.categoryTypeRepository.UpdateCategoryTypeById(ctx, repository.UpdateCategoryTypeByIdParams{ID: id, Name: patchDTO.Name})
}
