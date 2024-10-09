package service

import (
	"context"
	"strconv"

	"github.com/hexley21/fixup/internal/catalog/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/catalog/delivery/http/v1/dto/mapper"
	"github.com/hexley21/fixup/internal/catalog/repository"
)

type CategoryService interface {
	CreateCategory(ctx context.Context, dto dto.CreateCategoryDTO) (category dto.CategoryDTO, err error)
	DeleteCategoryById(ctx context.Context, id int32) error
	GetCategoryById(ctx context.Context, id int32) (dto dto.CategoryDTO, err error)
	GetCategories(ctx context.Context, page int32, per_page int32) (dtos []dto.CategoryDTO, err error)
	GetCategoriesByTypeId(ctx context.Context, id int32, page int32, per_page int32) ([]dto.CategoryDTO, error)
	UpdateCategoryById(ctx context.Context, id int32, dto dto.PatchCategoryDTO) error
}

type categoryServiceImpl struct {
	categoryRepository repository.CategoryRepository
	defaultPerPage     int32
	maxPerPage         int32
}

func NewCategoryService(categoryRepository repository.CategoryRepository, defaultPerPage int32, maxPerPage int32) *categoryServiceImpl {
	return &categoryServiceImpl{
		categoryRepository: categoryRepository,
		defaultPerPage:     defaultPerPage,
		maxPerPage:         maxPerPage,
	}
}

func (s *categoryServiceImpl) CreateCategory(ctx context.Context, dto dto.CreateCategoryDTO) (category dto.CategoryDTO, err error) {
	intTypeId, err := strconv.Atoi(dto.TypeID)
	if err != nil {
		return category, err
	}

	entity, err := s.categoryRepository.CreateCategory(ctx, repository.CreateCategoryParams{TypeID: int32(intTypeId), Name: dto.Name})
	if err != nil {
		return category, err
	}

	return mapper.MapCategoryToDTO(entity), err
}

func (s *categoryServiceImpl) DeleteCategoryById(ctx context.Context, id int32) error {
	return s.categoryRepository.DeleteCategoryById(ctx, id)
}

func (s *categoryServiceImpl) GetCategoryById(ctx context.Context, id int32) (dto dto.CategoryDTO, err error) {
	entity, err := s.categoryRepository.GetCategoryById(ctx, id)
	if err != nil {
		return dto, err
	}

	return mapper.MapCategoryToDTO(entity), err
}

func (s *categoryServiceImpl) GetCategories(ctx context.Context, page int32, per_page int32) ([]dto.CategoryDTO, error) {
	if per_page == 0 || per_page > s.maxPerPage {
		per_page = s.defaultPerPage
	}

	

	entities, err := s.categoryRepository.GetCategories(ctx, per_page*(page-1), per_page)
	if err != nil {
		return []dto.CategoryDTO{}, err
	}

	categories := make([]dto.CategoryDTO, len(entities))
	for i, e := range entities {
		categories[i] = mapper.MapCategoryToDTO(e)
	}

	return categories, err
}

func (s *categoryServiceImpl) GetCategoriesByTypeId(ctx context.Context, id int32, page int32, per_page int32) ([]dto.CategoryDTO, error) {
	if per_page == 0 || per_page > s.maxPerPage {
		per_page = s.defaultPerPage
	}

	entities, err := s.categoryRepository.GetCategoriesByTypeId(ctx, id, per_page*(page-1), per_page)
	if err != nil {
		return []dto.CategoryDTO{}, err
	}

	categories := make([]dto.CategoryDTO, len(entities))
	for i, e := range entities {
		categories[i] = mapper.MapCategoryToDTO(e)
	}
	
	return categories, err
}

func (s *categoryServiceImpl) UpdateCategoryById(ctx context.Context, id int32, dto dto.PatchCategoryDTO) error {
	return s.categoryRepository.UpdateCategoryById(ctx, repository.UpdateCategoryByIdParams{ID: id, Name: dto.Name})
}
