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

type SubcategoryService interface {
	Get(ctx context.Context, id int32) (domain.Subcategory, error)
	List(ctx context.Context, limit int64, offset int64) ([]domain.Subcategory, error)
	ListByCategoryId(ctx context.Context, categoryID int32, limit int64, offset int64) ([]domain.Subcategory, error)
	ListByTypeId(ctx context.Context, typeID int32, limit int64, offset int64) ([]domain.Subcategory, error)
	Create(ctx context.Context, info domain.SubcategoryInfo) (int32, error)
	Update(ctx context.Context, id int32, info domain.SubcategoryInfo) (domain.Subcategory, error)
	Delete(ctx context.Context, id int32) error
}

type subcategoryImpl struct {
	subcategoryRepo repository.Subcategory
}

func NewSubcategoryService(subcategoryRepo repository.Subcategory) *subcategoryImpl {
	return &subcategoryImpl{subcategoryRepo: subcategoryRepo}
}

// Get retrieves a subcategory by its ID from the repository.
// If the subcategory is not found, it returns ErrSubcategoryNotFound.
func (s *subcategoryImpl) Get(ctx context.Context, id int32) (domain.Subcategory, error) {
	subcategory, err := s.subcategoryRepo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Subcategory{}, ErrSubcategoryNotFound
		}
		return domain.Subcategory{}, err
	}

	return domain.NewSubcategory(subcategory.ID, subcategory.CategoryID, subcategory.Name), nil
}

// List retrieves a list of subcategories from the repository with the specified limit and offset.
func (s *subcategoryImpl) List(ctx context.Context, limit int64, offset int64) ([]domain.Subcategory, error) {
	list, err := s.subcategoryRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	entities := make([]domain.Subcategory, len(list))
	for i, sc := range list {
		entities[i] = domain.NewSubcategory(sc.ID, sc.CategoryID, sc.Name)
	}

	return entities, nil
}

// ListByCategoryId retrieves a list of subcategories by their category ID from the repository with the specified limit and offset.
func (s *subcategoryImpl) ListByCategoryId(ctx context.Context, categoryID int32, limit int64, offset int64) ([]domain.Subcategory, error) {
	list, err := s.subcategoryRepo.ListByCategoryId(ctx, categoryID, limit, offset)
	if err != nil {
		return nil, err
	}
	entities := make([]domain.Subcategory, len(list))
	for i, sc := range list {
		entities[i] = domain.NewSubcategory(sc.ID, sc.CategoryID, sc.Name)
	}

	return entities, nil
}

// ListByTypeId retrieves a list of subcategories by their type ID from the repository with the specified limit and offset.
func (s *subcategoryImpl) ListByTypeId(ctx context.Context, typeID int32, limit int64, offset int64) ([]domain.Subcategory, error) {
	list, err := s.subcategoryRepo.ListByTypeId(ctx, typeID, limit, offset)
	if err != nil {
		return nil, err
	}
	entities := make([]domain.Subcategory, len(list))
	for i, sc := range list {
		entities[i] = domain.NewSubcategory(sc.ID, sc.CategoryID, sc.Name)
	}

	return entities, nil
}

// Create adds a new subcategory to the repository using the provided SubcategoryInfo.
// If the subcategory name is already taken, it returns ErrSubcategoryNameTaken.
func (s *subcategoryImpl) Create(ctx context.Context, info domain.SubcategoryInfo) (int32, error) {
	subcategoryId, err := s.subcategoryRepo.Create(ctx, info)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.RaiseException {
			return 0, ErrSubcateogryNameTaken
		}

		return 0, err
	}

	return subcategoryId, nil
}

// Update modifies an existing subcategory in the repository using the provided SubcategoryInfo and ID.
// It returns the updated subcategory or an error if the update fails.
// If the subcategory is not found, it returns ErrSubcategoryNotFound.
// If the subcategory name is already taken, it returns ErrSubcategoryNameTaken.
// If the category ID is not found, it returns ErrCategoryNotFound.
func (s *subcategoryImpl) Update(ctx context.Context, id int32, info domain.SubcategoryInfo) (domain.Subcategory, error) {
	subcategory, err := s.subcategoryRepo.Update(ctx, id, info)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Subcategory{}, ErrSubcategoryNotFound
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.RaiseException:
				return domain.Subcategory{}, ErrSubcateogryNameTaken
			case pgerrcode.ForeignKeyViolation:
				return domain.Subcategory{}, ErrCategoryNotFound
			}
		}

		return domain.Subcategory{}, err
	}

	return domain.NewSubcategory(subcategory.ID, subcategory.CategoryID, subcategory.Name), nil
}

// Delete removes a subcategory from the repository by its ID.
// It returns an error if the deletion fails or if the subcategory is not found (indicated by no rows affected).
func (s *subcategoryImpl) Delete(ctx context.Context, id int32) error {
	ok, err := s.subcategoryRepo.Delete(ctx, id)
	if err != nil {
		return err
	}
	if !ok {
		return ErrSubcategoryNotFound
	}

	return nil
}
