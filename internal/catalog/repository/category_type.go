package repository

import (
	"context"

	"github.com/hexley21/fixup/internal/catalog/entity"
	"github.com/hexley21/fixup/pkg/infra/postgres"
	"github.com/hexley21/fixup/pkg/infra/postgres/pg_error"
)

type CategoryTypeRepository interface {
	postgres.Repository[CategoryTypeRepository]
	WithTx(tx postgres.PGXQuerier) CategoryTypeRepository
	CreateCategoryType(ctx context.Context, name string) (entity.CategoryType, error)
	DeleteCategoryTypeById(ctx context.Context, id int32) error
	GetCategoryTypeById(ctx context.Context, id int32) (entity.CategoryType, error)
	GetCategoryTypes(ctx context.Context, offset int32, limit int32) ([]entity.CategoryType, error)
	UpdateCategoryTypeById(ctx context.Context, arg UpdateCategoryTypeByIdParams) error
}

type categoryTypeRepositoryImpl struct {
	db postgres.PGXQuerier
}

func NewCategoryTypeRepository(dbtx postgres.PGXQuerier) *categoryTypeRepositoryImpl {
	return &categoryTypeRepositoryImpl{
		dbtx,
	}
}

func (r *categoryTypeRepositoryImpl) WithTx(tx postgres.PGXQuerier) CategoryTypeRepository {
	return NewCategoryTypeRepository(tx)
}

const createCategoryType = `-- name: CreateCategoryType :one
INSERT INTO category_types (name) VALUES ($1) RETURNING id, name
`

func (r *categoryTypeRepositoryImpl) CreateCategoryType(ctx context.Context, name string) (entity.CategoryType, error) {
	row := r.db.QueryRow(ctx, createCategoryType, name)
	var i entity.CategoryType
	err := row.Scan(&i.ID, &i.Name)
	return i, err
}

const deleteCategoryTypeById = `-- name: DeleteCategoryTypeById :exec
DELETE FROM category_types WHERE id = $1
`

func (r *categoryTypeRepositoryImpl) DeleteCategoryTypeById(ctx context.Context, id int32) error {
	result, err := r.db.Exec(ctx, deleteCategoryTypeById, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pg_error.ErrNotFound
	}

	return nil
}

const getCategoryTypeById = `-- name: GetCategoryTypeById :one
SELECT id, name FROM category_types WHERE id = $1
`

func (r *categoryTypeRepositoryImpl) GetCategoryTypeById(ctx context.Context, id int32) (entity.CategoryType, error) {
	row := r.db.QueryRow(ctx, getCategoryTypeById, id)
	var i entity.CategoryType
	err := row.Scan(&i.ID, &i.Name)
	return i, err
}

const updateCategoryTypeById = `-- name: UpdateCategoryTypeById :exec
UPDATE category_types SET name = $2 WHERE id = $1 Returning id, name
`

type UpdateCategoryTypeByIdParams struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

func (r *categoryTypeRepositoryImpl) UpdateCategoryTypeById(ctx context.Context, arg UpdateCategoryTypeByIdParams) error {
	result, err := r.db.Exec(ctx, updateCategoryTypeById, arg.ID, arg.Name)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pg_error.ErrNotFound
	}

	return nil
}

const getCategoryTypes = `-- name: GetCategoryTypes :many
SELECT id, name FROM category_types ORDER BY id DESC OFFSET $1 LIMIT $2
`

func (r *categoryTypeRepositoryImpl) GetCategoryTypes(ctx context.Context, offset int32, limit int32) ([]entity.CategoryType, error) {
	rows, err := r.db.Query(ctx, getCategoryTypes, offset, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []entity.CategoryType
	for rows.Next() {
		var i entity.CategoryType
		if err := rows.Scan(&i.ID, &i.Name); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
