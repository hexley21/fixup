package service

import (
	"context"
	"errors"

	"github.com/hexley21/fixup/internal/catalog/domain"
	"github.com/hexley21/fixup/internal/catalog/repository"
	"github.com/jackc/pgx/v5"
)

type CategoryType interface {
	Create(ctx context.Context, name string) (domain.CategoryType, error)
	Delete(ctx context.Context, id int32) error
	Get(ctx context.Context, id int32) (domain.CategoryType, error)
	List(ctx context.Context, limit int64, offset int64) ([]domain.CategoryType, error)
	Update(ctx context.Context, id int32, name string) error
}

type categoryTypeImpl struct {
	categoryTypeRepository repository.CategoryTypeRepository
}

func NewCategoryTypeService(categoryTypeRepository repository.CategoryTypeRepository) *categoryTypeImpl {
	return &categoryTypeImpl{
		categoryTypeRepository: categoryTypeRepository,
	}
}

func (s *categoryTypeImpl) Create(ctx context.Context, name string) (domain.CategoryType, error) {
	model, err := s.categoryTypeRepository.Create(ctx, name)
	if err != nil {
		return domain.CategoryType{}, err
	}

	return domain.NewCategoryType(model.ID, model.Name), nil
}

func (s *categoryTypeImpl) Delete(ctx context.Context, id int32) error {
	ok, err := s.categoryTypeRepository.Delete(ctx, id)
	if err != nil {
		return err
	}
	if !ok {
		return ErrCategoryTypeNotFound
	}

	return nil
}

func (s *categoryTypeImpl) Get(ctx context.Context, id int32) (domain.CategoryType, error) {
	model, err := s.categoryTypeRepository.Get(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.CategoryType{}, ErrCategoryTypeNotFound
		}
		return domain.CategoryType{}, err
	}

	return domain.NewCategoryType(model.ID, model.Name), nil
}

func (s *categoryTypeImpl) List(ctx context.Context, limit int64, offset int64) ([]domain.CategoryType, error) {
	list, err := s.categoryTypeRepository.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	categoryTypes := make([]domain.CategoryType, len(list))
	for i, ct := range list {
		categoryTypes[i] = domain.NewCategoryType(ct.ID, ct.Name)
	}

	return categoryTypes, nil
}

func (s *categoryTypeImpl) Update(ctx context.Context, id int32, name string) error {
	ok, err := s.categoryTypeRepository.Update(ctx, id, name)
	if err != nil {
		return err
	}
	if !ok {
		return ErrCategoryTypeNotFound
	}

	return nil
}
