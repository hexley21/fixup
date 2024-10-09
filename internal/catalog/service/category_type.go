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
	GetCategoryTypes(ctx context.Context, page int32, per_page int32) (dtos []dto.CategoryTypeDTO, err error)
	UpdateCategoryTypeById(ctx context.Context, id int32, dto dto.PatchCategoryTypeDTO) error
}

type categoryTypeServiceImpl struct {
	categoryTypeRepository repository.CategoryTypeRepository
	defaultPerPage     int32
	maxPerPage         int32
}

func NewCategoryTypeService(categoryTypeRepository repository.CategoryTypeRepository, defaultPerPage int32, maxPerPage int32) *categoryTypeServiceImpl {
	return &categoryTypeServiceImpl{
		categoryTypeRepository: categoryTypeRepository,
		defaultPerPage:     defaultPerPage,
		maxPerPage:         maxPerPage,
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

func (s *categoryTypeServiceImpl) GetCategoryTypes(ctx context.Context, page int32, per_page int32) ([]dto.CategoryTypeDTO, error) {
	if per_page == 0 || per_page > s.maxPerPage {
		per_page = s.defaultPerPage
	}

	entities, err := s.categoryTypeRepository.GetCategoryTypes(ctx, per_page*(page-1), per_page)
	if err != nil {
		return []dto.CategoryTypeDTO{}, err
	}

	dtos := make([]dto.CategoryTypeDTO, len(entities))
	for i, e := range entities {
		dtos[i] = mapper.MapCategoryTypeToDTO(e)
	}

	return dtos, err
}

func (s *categoryTypeServiceImpl) UpdateCategoryTypeById(ctx context.Context, id int32, dto dto.PatchCategoryTypeDTO) error {
	return s.categoryTypeRepository.UpdateCategoryTypeById(ctx, repository.UpdateCategoryTypeByIdParams{ID: id, Name: dto.Name})
}
