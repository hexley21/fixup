package service

import (
	"context"
	"strconv"

	"github.com/hexley21/fixup/internal/catalog/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/catalog/delivery/http/v1/dto/mapper"
	"github.com/hexley21/fixup/internal/catalog/repository"
)

type CategoryService interface {
	CreateCategory(ctx context.Context, createDTO dto.CreateCategoryDTO) (dto.CategoryDTO, error)
	DeleteCategoryById(ctx context.Context, id int32) error
	GetCategoryById(ctx context.Context, id int32) (dto.CategoryDTO, error)
	GetCategories(ctx context.Context, page int32, perPage int32) ([]dto.CategoryDTO, error)
	GetCategoriesByTypeId(ctx context.Context, id int32, page int32, perPage int32) ([]dto.CategoryDTO, error)
	UpdateCategoryById(ctx context.Context, id int32, patchDTO dto.PatchCategoryDTO) (dto.CategoryDTO, error)
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

func (s *categoryServiceImpl) CreateCategory(ctx context.Context, createDTO dto.CreateCategoryDTO) (dto.CategoryDTO, error) {
	intTypeId, err := strconv.Atoi(createDTO.TypeID)
	if err != nil {
		return dto.CategoryDTO{}, err
	}

	entity, err := s.categoryRepository.CreateCategory(ctx, repository.CreateCategoryParams{TypeID: int32(intTypeId), Name: createDTO.Name})
	if err != nil {
		return dto.CategoryDTO{}, err
	}

	return mapper.MapCategoryToDTO(entity), nil
}

func (s *categoryServiceImpl) DeleteCategoryById(ctx context.Context, id int32) error {
	return s.categoryRepository.DeleteCategoryById(ctx, id)
}

func (s *categoryServiceImpl) GetCategoryById(ctx context.Context, id int32) (dto.CategoryDTO, error) {
	entity, err := s.categoryRepository.GetCategoryById(ctx, id)
	if err != nil {
		return dto.CategoryDTO{}, err
	}

	return mapper.MapCategoryToDTO(entity), nil
}

func (s *categoryServiceImpl) GetCategories(ctx context.Context, page int32, perPage int32) ([]dto.CategoryDTO, error) {
	if perPage == 0 || perPage > s.maxPerPage {
		perPage = s.defaultPerPage
	}

	entities, err := s.categoryRepository.GetCategories(ctx, perPage*(page-1), perPage)
	if err != nil {
		return []dto.CategoryDTO{}, err
	}

	categories := make([]dto.CategoryDTO, len(entities))
	for i, e := range entities {
		categories[i] = mapper.MapCategoryToDTO(e)
	}

	return categories, nil
}

func (s *categoryServiceImpl) GetCategoriesByTypeId(ctx context.Context, id int32, page int32, perPage int32) ([]dto.CategoryDTO, error) {
	if perPage == 0 || perPage > s.maxPerPage {
		perPage = s.defaultPerPage
	}

	entities, err := s.categoryRepository.GetCategoriesByTypeId(ctx, id, perPage*(page-1), perPage)
	if err != nil {
		return []dto.CategoryDTO{}, err
	}

	categories := make([]dto.CategoryDTO, len(entities))
	for i, e := range entities {
		categories[i] = mapper.MapCategoryToDTO(e)
	}

	return categories, nil
}

func (s *categoryServiceImpl) UpdateCategoryById(ctx context.Context, id int32, patchDTO dto.PatchCategoryDTO) (dto.CategoryDTO, error) {
	entity, err := s.categoryRepository.UpdateCategoryById(ctx, repository.UpdateCategoryByIdParams{ID: id, Name: patchDTO.Name})

	if err != nil {
		return dto.CategoryDTO{}, err
	}

	return mapper.MapCategoryToDTO(entity), nil
}
