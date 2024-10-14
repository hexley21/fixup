package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/hexley21/fixup/internal/catalog/entity"
	"github.com/hexley21/fixup/internal/catalog/repository"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Subcategory interface {
	Get(ctx context.Context, id int32) (entity.Subcategory, error)
	List(ctx context.Context, limit int32, offset int32) ([]entity.Subcategory, error)
	ListByCategoryId(ctx context.Context, categoryID int32, limit int32, offset int32) ([]entity.Subcategory, error)
	ListByTypeId(ctx context.Context, typeID int32, limit int32, offset int32) ([]entity.Subcategory, error)
	Create(ctx context.Context, info entity.SubcategoryInfo) (entity.Subcategory, error)
	Update(ctx context.Context, id int32, info entity.SubcategoryInfo) (entity.Subcategory, error)
	Delete(ctx context.Context, id int32) error
}

type subcategoryImpl struct {
	subcategoryRepo repository.Subcategory
}

func NewSubcategoryRepository(subcategoryRepo repository.Subcategory) *subcategoryImpl {
	return &subcategoryImpl{subcategoryRepo: subcategoryRepo}
}

func (s *subcategoryImpl) Get(ctx context.Context, id int32) (entity.Subcategory, error) {
	subcategory, err := s.subcategoryRepo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Subcategory{}, ErrSubcategoryNotFound
		}
		return entity.Subcategory{}, fmt.Errorf("failed to fetch subcategory: %w", err)
	}

	return subcategory, nil
}

func (s *subcategoryImpl) List(ctx context.Context, limit int32, offset int32) ([]entity.Subcategory, error) {
	list, err := s.subcategoryRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch subcategories: %w", err)
	}

	return list, nil
}

func (s *subcategoryImpl) ListByCategoryId(ctx context.Context, categoryID int32, limit int32, offset int32) ([]entity.Subcategory, error) {
	list, err := s.subcategoryRepo.ListByCategoryId(ctx, categoryID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch subcategories: %w", err)
	}

	return list, nil
}

func (s *subcategoryImpl) ListByTypeId(ctx context.Context, typeID int32, limit int32, offset int32) ([]entity.Subcategory, error) {
	list, err := s.subcategoryRepo.ListByTypeId(ctx, typeID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch subcategories: %w", err)
	}

	return list, nil
}

func (s *subcategoryImpl) Create(ctx context.Context, info entity.SubcategoryInfo) (entity.Subcategory, error) {
	subcategory, err := s.subcategoryRepo.Create(ctx, info)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.RaiseException {
			return entity.Subcategory{}, ErrSubcateogryNameTaken
		}

		return entity.Subcategory{}, fmt.Errorf("failed to create subcategory: %w", err)
	}

	return subcategory, nil
}

func (s *subcategoryImpl) Update(ctx context.Context, id int32, info entity.SubcategoryInfo) (entity.Subcategory, error) {
	subcategory, err := s.subcategoryRepo.Update(ctx, id, info)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Subcategory{}, ErrSubcategoryNotFound
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.RaiseException {
			return entity.Subcategory{}, ErrSubcateogryNameTaken
		}

		return entity.Subcategory{}, fmt.Errorf("failed to update subcategory: %w", err)
	}

	return subcategory, nil
}

func (s *subcategoryImpl) Delete(ctx context.Context, id int32) error {
	if err := s.subcategoryRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrSubcategoryNotFound
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.RaiseException {
			return ErrSubcateogryNameTaken
		}

		return fmt.Errorf("failed to update subcategory: %w", err)
	}

	return nil
}
