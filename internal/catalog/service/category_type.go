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

type CategoryTypeService interface {
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

// Create adds a new category type to the repository using the provided name.
// If the category type name is already taken, it returns ErrCategoryTypeNameTaken.
func (s *categoryTypeImpl) Create(ctx context.Context, name string) (domain.CategoryType, error) {
	model, err := s.categoryTypeRepository.Create(ctx, name)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return domain.CategoryType{}, ErrCategoryTypeNameTaken
		}
		return domain.CategoryType{}, err
	}

	return domain.NewCategoryType(model.ID, model.Name), nil
}

// Delete removes a category type from the repository by its ID.
// It returns an error if the deletion fails or if the category type is not found (indicated by no rows affected).
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

// Get retrieves a category type by its ID from the repository.
// If the category type is not found, it returns ErrCategoryTypeNotFound.
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

// List retrieves a list of category types from the repository with the specified limit and offset.
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

// Update modifies an existing category type in the repository using the provided ID and name.
// It returns ErrCategoryTypeNameTaken if name is taken
// If category type not found, returns ErrCategoryNotFound.
func (s *categoryTypeImpl) Update(ctx context.Context, id int32, name string) error {
	ok, err := s.categoryTypeRepository.Update(ctx, id, name)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return ErrCategoryTypeNameTaken
		}
		return err
	}
	if !ok {
		return ErrCategoryTypeNotFound
	}

	return nil
}
