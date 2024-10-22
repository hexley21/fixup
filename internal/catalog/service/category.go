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

// Create adds a new category to the repository using the provided CategoryInfo.
// If the category name is already taken, it returns ErrCategoryNameTaken.
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

// Delete removes a category from the repository by its ID.
// It returns an error if the deletion fails or if the category is not found (indicated by no rows affected).
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

// Get retrieves a category by its ID from the repository.
// If the category is not found, it returns ErrCategoryNotFound.
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

// List retrieves a list of categories from the repository with the specified limit and offset.
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

// ListByTypeId retrieves a list of categories by their type ID from the repository with the specified limit and offset.
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

// Update modifies an existing category in the repository using the provided CategoryInfo and ID.
// If the category is not found, it returns ErrCategoryNotFound.
// If the category name is already taken, it returns ErrCategoryNameTaken.
// If the category type ID is not found, it returns ErrCategoryTypeNotFound.
func (s *categoryImpl) Update(ctx context.Context, id int32, info domain.CategoryInfo) (domain.Category, error) {
	model, err := s.categoryRepository.Update(ctx, id, info)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Category{}, ErrCategoryNotFound
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.RaiseException:
				return domain.Category{}, ErrCategoryNameTaken
			case pgerrcode.ForeignKeyViolation:
				return domain.Category{}, ErrCategoryTypeNotFound
			}
		}
		return domain.Category{}, err
	}

	return domain.NewCategory(model.ID, model.TypeID, model.Name), nil
}
