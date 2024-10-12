package repository

import (
	"context"

	"github.com/hexley21/fixup/internal/catalog/entity"
	"github.com/hexley21/fixup/pkg/infra/postgres"
	"github.com/hexley21/fixup/pkg/infra/postgres/pg_error"
)

type SubcategoryRepository interface {
	postgres.Repository[SubcategoryRepository]
	CreateSubcategory(ctx context.Context, arg CreateSubcategoryParams) (entity.Subcategory, error)
	GetSubcategoryById(ctx context.Context, id int32) (entity.Subcategory, error)
	GetSubategories(ctx context.Context, limit int32, offset int32) ([]entity.Subcategory, error)
	GetSubategoriesByCategoryId(ctx context.Context, categoryID int32, limit int32, offset int32) ([]entity.Subcategory, error)
	GetSubategoriesByTypeId(ctx context.Context, typeID int32, limit int32, offset int32) ([]entity.Subcategory, error)
	UpdateSubcategoryById(ctx context.Context, arg UpdateSubcategoryByIdParams) (entity.Subcategory, error)
	DeleteSubcategoryById(ctx context.Context, id int32) error
}

type postgresSubcategoryRepository struct {
	db postgres.PGXQuerier
}

func NewSubcategoryRepository(dbtx postgres.PGXQuerier) *postgresSubcategoryRepository {
	return &postgresSubcategoryRepository{
		dbtx,
	}
}

func (r *postgresSubcategoryRepository) WithTx(tx postgres.PGXQuerier) SubcategoryRepository {
	return NewSubcategoryRepository(tx)
}

const createSubcategory = `-- name: CreateSubcategory :one
INSERT INTO subcategories (category_id, name) VALUES ($1, $2) RETURNING id, category_id, name
`

type CreateSubcategoryParams struct {
	CategoryID int32  `json:"category_id"`
	Name       string `json:"name"`
}

func (r *postgresSubcategoryRepository) CreateSubcategory(ctx context.Context, arg CreateSubcategoryParams) (entity.Subcategory, error) {
	row := r.db.QueryRow(ctx, createSubcategory, arg.CategoryID, arg.Name)
	var i entity.Subcategory
	err := row.Scan(&i.ID, &i.CategoryID, &i.Name)
	return i, err
}

const getSubategories = `-- name: GetSubategories :many
SELECT id, category_id, name FROM subcategories ORDER BY id LIMIT $1 OFFSET $2
`

func (r *postgresSubcategoryRepository) GetSubategories(ctx context.Context, limit int32, offset int32) ([]entity.Subcategory, error) {
	rows, err := r.db.Query(ctx, getSubategories, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []entity.Subcategory
	for rows.Next() {
		var i entity.Subcategory
		if err := rows.Scan(&i.ID, &i.CategoryID, &i.Name); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getSubategoriesByCategoryId = `-- name: GetSubategoriesByCategoryId :many
SELECT id, category_id, name FROM subcategories WHERE category_id = $1 ORDER BY id LIMIT $2 OFFSET $3
`

func (r *postgresSubcategoryRepository) GetSubategoriesByCategoryId(ctx context.Context, categoryID int32, limit int32, offset int32) ([]entity.Subcategory, error) {
	rows, err := r.db.Query(ctx, getSubategoriesByCategoryId, categoryID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []entity.Subcategory
	for rows.Next() {
		var i entity.Subcategory
		if err := rows.Scan(&i.ID, &i.CategoryID, &i.Name); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getSubategoriesByTypeId = `-- name: GetSubategoriesByTypeId :many
SELECT s.id, s.category_id, s.name 
FROM subcategories s
JOIN categories c ON s.category_id = c.id
WHERE c.type_id = $1
ORDER BY s.id LIMIT $2 OFFSET $3
`

func (r *postgresSubcategoryRepository) GetSubategoriesByTypeId(ctx context.Context, typeID int32, limit int32, offset int32) ([]entity.Subcategory, error) {
	rows, err := r.db.Query(ctx, getSubategoriesByTypeId, typeID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []entity.Subcategory
	for rows.Next() {
		var i entity.Subcategory
		if err := rows.Scan(&i.ID, &i.CategoryID, &i.Name); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getSubcategoryById = `-- name: GetSubcategoryById :one
SELECT id, category_id, name FROM subcategories WHERE id = $1
`

func (r *postgresSubcategoryRepository) GetSubcategoryById(ctx context.Context, id int32) (entity.Subcategory, error) {
	row := r.db.QueryRow(ctx, getSubcategoryById, id)
	var i entity.Subcategory
	err := row.Scan(&i.ID, &i.CategoryID, &i.Name)
	return i, err
}

const updateSubcategoryById = `-- name: UpdateSubcategoryById :one
UPDATE subcategories SET name = $1, category_id = $2 WHERE id = $3 RETURNING id, category_id, name
`

type UpdateSubcategoryByIdParams struct {
	Name       string `json:"name"`
	CategoryID int32  `json:"category_id"`
	ID         int32  `json:"id"`
}

func (r *postgresSubcategoryRepository) UpdateSubcategoryById(ctx context.Context, arg UpdateSubcategoryByIdParams) (entity.Subcategory, error) {
	row := r.db.QueryRow(ctx, updateSubcategoryById, arg.Name, arg.CategoryID, arg.ID)
	var i entity.Subcategory
	err := row.Scan(&i.ID, &i.CategoryID, &i.Name)
	return i, err
}

const deleteSubcategoryById = `-- name: DeleteSubcategoryById :exec
DELETE FROM subcategories WHERE id = $1
`

func (r *postgresSubcategoryRepository) DeleteSubcategoryById(ctx context.Context, id int32) error {
	result, err := r.db.Exec(ctx, deleteSubcategoryById, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pg_error.ErrNotFound
	}

	return nil
}
