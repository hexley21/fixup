package repository

import (
	"context"

	"github.com/hexley21/fixup/internal/catalog/entity"
	"github.com/hexley21/fixup/pkg/infra/postgres"
	"github.com/hexley21/fixup/pkg/infra/postgres/pg_error"
)

type CategoryRepository interface {
	postgres.Repository[CategoryRepository]
	CreateCategory(ctx context.Context, arg CreateCategoryParams) (entity.Category, error)
	DeleteCategoryById(ctx context.Context, id int32) error
	GetCategories(ctx context.Context, offset int32, limit int32) ([]entity.Category, error)
	GetCategoryById(ctx context.Context, id int32) (entity.Category, error)
	GetCategoriesByTypeId(ctx context.Context, id int32, offset int32, limit int32) ([]entity.Category, error)
	UpdateCategoryById(ctx context.Context, arg UpdateCategoryByIdParams) (entity.Category, error)
}

type postgresCategoryRepository struct {
	db postgres.PGXQuerier
}

func NewCategoryRepository(dbtx postgres.PGXQuerier) *postgresCategoryRepository {
	return &postgresCategoryRepository{
		dbtx,
	}
}

func (r *postgresCategoryRepository) WithTx(tx postgres.PGXQuerier) CategoryRepository {
	return NewCategoryRepository(tx)
}

const createCategory = `-- name: CreateCategory :one
INSERT INTO categories (type_id, name) VALUES ($1, $2) RETURNING id, type_id, name
`

type CreateCategoryParams struct {
	TypeID int32  `json:"type_id"`
	Name   string `json:"name"`
}

func (r *postgresCategoryRepository) CreateCategory(ctx context.Context, arg CreateCategoryParams) (entity.Category, error) {
	row := r.db.QueryRow(ctx, createCategory, arg.TypeID, arg.Name)
	var i entity.Category
	err := row.Scan(&i.ID, &i.TypeID, &i.Name)
	return i, err
}

const deleteCategoryById = `-- name: DeleteCategoryById :exec
DELETE FROM categories WHERE id = $1
`

func (r *postgresCategoryRepository) DeleteCategoryById(ctx context.Context, id int32) error {
	result, err := r.db.Exec(ctx, deleteCategoryById, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pg_error.ErrNotFound
	}

	return nil
}

const getCategories = `-- name: GetCategories :many
SELECT id, type_id, name FROM categories ORDER BY id DESC OFFSET $1 LIMIT $2
`

func (r *postgresCategoryRepository) GetCategories(ctx context.Context, offset int32, limit int32) ([]entity.Category, error) {
	rows, err := r.db.Query(ctx, getCategories, offset, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []entity.Category
	for rows.Next() {
		var i entity.Category
		if err := rows.Scan(&i.ID, &i.TypeID, &i.Name); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getCategoryById = `-- name: GetCategoryById :one
SELECT id, type_id, name FROM categories WHERE id = $1
`

func (r *postgresCategoryRepository) GetCategoryById(ctx context.Context, id int32) (entity.Category, error) {
	row := r.db.QueryRow(ctx, getCategoryById, id)
	var i entity.Category
	err := row.Scan(&i.ID, &i.TypeID, &i.Name)
	return i, err
}

const getCategoriesByTypeId = `-- name: GetCategoriesByTypeId :many
SELECT id, type_id, name FROM categories WHERE id = $1 ORDER BY id DESC OFFSET $2 LIMIT $3
`

func (r *postgresCategoryRepository) GetCategoriesByTypeId(ctx context.Context, id int32, offset int32, limit int32) ([]entity.Category, error) {
	rows, err := r.db.Query(ctx, getCategoriesByTypeId, id, offset, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []entity.Category
	for rows.Next() {
		var i entity.Category
		if err := rows.Scan(&i.ID, &i.TypeID, &i.Name); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateCategoryById = `-- name: UpdateCategoryById :one
UPDATE categories SET name = $2, type_id = $3 WHERE id = $1 Returning id, type_id, name
`

type UpdateCategoryByIdParams struct {
	ID     int32  `json:"id"`
	Name   string `json:"name"`
	TypeID int32  `json:"type_id"`
}

func (r *postgresCategoryRepository) UpdateCategoryById(ctx context.Context, arg UpdateCategoryByIdParams) (entity.Category, error) {
	row := r.db.QueryRow(ctx, updateCategoryById, arg.ID, arg.Name, arg.TypeID)
	var i entity.Category
	err := row.Scan(&i.ID, &i.TypeID, &i.Name)
	return i, err
}
