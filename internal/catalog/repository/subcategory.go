package repository

import (
	"context"

	"github.com/hexley21/fixup/internal/catalog/domain"
	"github.com/hexley21/fixup/pkg/infra/postgres"
)

type Subcategory interface {
	postgres.Repository[Subcategory]
	Get(ctx context.Context, id int32) (SubcategoryModel, error)
	List(ctx context.Context, offset int32, limit int32) ([]SubcategoryModel, error)
	ListByCategoryId(ctx context.Context, categoryID int32, offset int32, limit int32) ([]SubcategoryModel, error)
	ListByTypeId(ctx context.Context, typeID int32, offset int32, limit int32) ([]SubcategoryModel, error)
	Create(ctx context.Context, info domain.SubcategoryInfo) (int32, error)
	Update(ctx context.Context, id int32, info domain.SubcategoryInfo) (SubcategoryModel, error)
	Delete(ctx context.Context, id int32) (bool, error)
}

type postgresSubcategoryRepository struct {
	db postgres.PGXQuerier
}

func NewSubcategoryRepository(dbtx postgres.PGXQuerier) *postgresSubcategoryRepository {
	return &postgresSubcategoryRepository{
		dbtx,
	}
}

func (r *postgresSubcategoryRepository) WithTx(tx postgres.PGXQuerier) Subcategory {
	return NewSubcategoryRepository(tx)
}

const getSubcategoryById = `-- name: GetSubcategoryById :one
SELECT id, category_id, name FROM subcategories WHERE id = $1
`

func (r *postgresSubcategoryRepository) Get(ctx context.Context, id int32) (SubcategoryModel, error) {
	row := r.db.QueryRow(ctx, getSubcategoryById, id)
	var i SubcategoryModel
	err := row.Scan(&i.ID, &i.CategoryID, &i.Name)
	return i, err
}

const listSubategories = `-- name: ListSubategories :many
SELECT id, category_id, name FROM subcategories ORDER BY id OFFSET $1 LIMIT $2
`

func (r *postgresSubcategoryRepository) List(ctx context.Context, offset int32, limit int32) ([]SubcategoryModel, error) {
	rows, err := r.db.Query(ctx, listSubategories, offset, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SubcategoryModel
	for rows.Next() {
		var i SubcategoryModel
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

const listSubategoriesByCategoryId = `-- name: ListSubategoriesByCategoryId :many
SELECT id, category_id, name FROM subcategories WHERE category_id = $1 ORDER BY id OFFSET $2 LIMIT $3
`

func (r *postgresSubcategoryRepository) ListByCategoryId(ctx context.Context, categoryID int32, offset int32, limit int32) ([]SubcategoryModel, error) {
	rows, err := r.db.Query(ctx, listSubategoriesByCategoryId, categoryID, offset, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SubcategoryModel
	for rows.Next() {
		var i SubcategoryModel
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

const listSubategoriesByTypeId = `-- name: ListSubategoriesByTypeId :many
SELECT s.id, s.category_id, s.name 
FROM subcategories s
JOIN categories c ON s.category_id = c.id
WHERE c.type_id = $1
ORDER BY s.id OFFSET $2 LIMIT $3
`

func (r *postgresSubcategoryRepository) ListByTypeId(ctx context.Context, typeID int32, offset int32, limit int32) ([]SubcategoryModel, error) {
	rows, err := r.db.Query(ctx, listSubategoriesByTypeId, typeID, offset, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SubcategoryModel
	for rows.Next() {
		var i SubcategoryModel
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

const createSubcategory = `-- name: CreateSubcategory :one
INSERT INTO subcategories (category_id, name) VALUES ($1, $2) RETURNING id
`

func (r *postgresSubcategoryRepository) Create(ctx context.Context, info domain.SubcategoryInfo) (int32, error) {
	row := r.db.QueryRow(ctx, createSubcategory, info.CategoryID, info.Name)
	var id int32
	err := row.Scan(&id)
	return id, err
}

const updateSubcategory = `-- name: UpdateSubcategory :one
UPDATE subcategories SET name = $1, category_id = $2 WHERE id = $3 RETURNING *
`

// TODO: include partial update ability
func (r *postgresSubcategoryRepository) Update(ctx context.Context, id int32, info domain.SubcategoryInfo) (SubcategoryModel, error) {
	row := r.db.QueryRow(ctx, updateSubcategory, info.Name, info.CategoryID, id)
	var i SubcategoryModel
	err := row.Scan(&i.ID, &i.CategoryID, &i.Name)
	return i, err
}

const deleteSubcategory = `-- name: DeleteSubcategory :exec
DELETE FROM subcategories WHERE id = $1
`

func (r *postgresSubcategoryRepository) Delete(ctx context.Context, id int32) (bool, error) {
	result, err := r.db.Exec(ctx, deleteSubcategory, id)
	return result.RowsAffected() > 0, err
}
