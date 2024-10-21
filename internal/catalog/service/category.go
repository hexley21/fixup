package service

import (
	"context"
	"errors"

	"github.com/hexley21/fixup/internal/catalog/domain"
	"github.com/hexley21/fixup/internal/catalog/repository"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type CategoryService interface {
	Create(ctx context.Context, info domain.CategoryInfo) (int32, error)
	Delete(ctx context.Context, id int32) error
	Get(ctx context.Context, id int32) (domain.Category, error)
	List(ctx context.Context, limit int64, offset int64) ([]domain.Category, error)
	ListByTypeId(ctx context.Context, id int32, limit int64, offset int64) ([]domain.Category, error)
	Update(ctx context.Context, id int32, info domain.CategoryInfo) (domain.Category, error)
}

type categoryImpl struct {
	categoryRepository repository.CategoryRepository
}

func NewCategoryService(categoryRepository repository.CategoryRepository) *categoryImpl {
	return &categoryImpl{
		categoryRepository: categoryRepository,
	}
}

func (s *categoryImpl) Create(ctx context.Context, info domain.CategoryInfo) (int32, error) {
	categoryId, err := s.categoryRepository.Create(ctx, info)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.RaiseException {
			return 0, ErrCategoryNameTaken
		}

		return 0, err
	}

	return categoryId, nil
}

func (s *categoryImpl) Delete(ctx context.Context, id int32) error {
	ok, err := s.categoryRepository.Delete(ctx, id)
	if err != nil {
		return err
	}
	if !ok {
		return ErrCategoryNotFound
	}

	return nil
}

func (s *categoryImpl) Get(ctx context.Context, id int32) (domain.Category, error) {
	model, err := s.categoryRepository.Get(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Category{}, ErrCategoryNotFound
		}
		return domain.Category{}, err
	}

	return domain.NewCategory(model.ID, model.TypeID, model.Name), nil
}

func (s *categoryImpl) List(ctx context.Context, limit int64, offset int64) ([]domain.Category, error) {
	list, err := s.categoryRepository.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	categories := make([]domain.Category, len(list))
	for i, c := range list {
		categories[i] = domain.NewCategory(c.ID, c.TypeID, c.Name)
	}

	return categories, nil
}

func (s *categoryImpl) ListByTypeId(ctx context.Context, id int32, limit int64, offset int64) ([]domain.Category, error) {
	list, err := s.categoryRepository.ListByTypeId(ctx, id, limit, offset)
	if err != nil {
		return nil, err
	}

	categories := make([]domain.Category, len(list))
	for i, c := range list {
		categories[i] = domain.NewCategory(c.ID, c.TypeID, c.Name)
	}

	return categories, nil
}

func (s *categoryImpl) Update(ctx context.Context, id int32, info domain.CategoryInfo) (domain.Category, error) {
	model, err := s.categoryRepository.Update(ctx, id, info)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Category{}, ErrCategoryNotFound
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.RaiseException {
			return domain.Category{}, ErrCategoryNameTaken
		}

		return domain.Category{}, err
	}

	return domain.NewCategory(model.ID, model.TypeID, model.Name), nil
}
