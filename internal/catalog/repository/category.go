package repository

import (
	"context"

	"github.com/hexley21/fixup/pkg/infra/postgres"
)

type CategoryRepository interface {
	postgres.Repository[CategoryRepository]
	Create(ctx context.Context, arg CreateCategoryParams) (CategoryModel, error)
	Delete(ctx context.Context, id int32) (bool, error)
	List(ctx context.Context, offset int32, limit int32) ([]CategoryModel, error)
	Get(ctx context.Context, id int32) (CategoryModel, error)
	ListByTypeId(ctx context.Context, id int32, offset int32, limit int32) ([]CategoryModel, error)
	Update(ctx context.Context, id int32, arg UpdateCategoryParams) (CategoryModel, error)
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
	TypeID int32
	Name   string
}

func (r *postgresCategoryRepository) Create(ctx context.Context, arg CreateCategoryParams) (CategoryModel, error) {
	row := r.db.QueryRow(ctx, createCategory, arg.TypeID, arg.Name)
	var i CategoryModel
	err := row.Scan(&i.ID, &i.TypeID, &i.Name)
	return i, err
}

const deleteCategory = `-- name: DeleteCategory :exec
DELETE FROM categories WHERE id = $1
`

func (r *postgresCategoryRepository) Delete(ctx context.Context, id int32) (bool, error) {
	result, err := r.db.Exec(ctx, deleteCategory, id)
	return result.RowsAffected() > 0, err
}

const listCategories = `-- name: ListCategories :many
SELECT id, type_id, name FROM categories ORDER BY id DESC OFFSET $1 LIMIT $2
`

func (r *postgresCategoryRepository) List(ctx context.Context, offset int32, limit int32) ([]CategoryModel, error) {
	rows, err := r.db.Query(ctx, listCategories, offset, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []CategoryModel
	for rows.Next() {
		var i CategoryModel
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

const getCategory = `-- name: GetCategory :one
SELECT id, type_id, name FROM categories WHERE id = $1
`

func (r *postgresCategoryRepository) Get(ctx context.Context, id int32) (CategoryModel, error) {
	row := r.db.QueryRow(ctx, getCategory, id)
	var i CategoryModel
	err := row.Scan(&i.ID, &i.TypeID, &i.Name)
	return i, err
}

const listCategoriesByTypeId = `-- name: ListCategoriesByTypeId :many
SELECT id, type_id, name FROM categories WHERE id = $1 ORDER BY id DESC OFFSET $2 LIMIT $3
`

func (r *postgresCategoryRepository) ListByTypeId(ctx context.Context, id int32, offset int32, limit int32) ([]CategoryModel, error) {
	rows, err := r.db.Query(ctx, listCategoriesByTypeId, id, offset, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []CategoryModel
	for rows.Next() {
		var i CategoryModel
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

type UpdateCategoryParams struct {
	Name   string
	TypeID int32
}

func (r *postgresCategoryRepository) Update(ctx context.Context, id int32, arg UpdateCategoryParams) (CategoryModel, error) {
	row := r.db.QueryRow(ctx, updateCategoryById, id, arg.Name, arg.TypeID)
	var i CategoryModel
	err := row.Scan(&i.ID, &i.TypeID, &i.Name)
	return i, err
}
