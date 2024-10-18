package repository

import (
	"context"

	"github.com/hexley21/fixup/pkg/infra/postgres"
)

type CategoryTypeRepository interface {
	postgres.Repository[CategoryTypeRepository]
	Create(ctx context.Context, name string) (CategoryTypeModel, error)
	Delete(ctx context.Context, id int32) (bool, error)
	Get(ctx context.Context, id int32) (CategoryTypeModel, error)
	Update(ctx context.Context, id int32, name string) (bool, error)
	List(ctx context.Context, limit int64, offset int64) ([]CategoryTypeModel, error)
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

func (r *categoryTypeRepositoryImpl) Create(ctx context.Context, name string) (CategoryTypeModel, error) {
	row := r.db.QueryRow(ctx, createCategoryType, name)
	var i CategoryTypeModel
	err := row.Scan(&i.ID, &i.Name)
	return i, err
}

const deleteCategoryType = `-- name: DeleteCategoryType :exec
DELETE FROM category_types WHERE id = $1
`

func (r *categoryTypeRepositoryImpl) Delete(ctx context.Context, id int32) (bool, error) {
	result, err := r.db.Exec(ctx, deleteCategoryType, id)
	return result.RowsAffected() > 0, err
}

const getCategoryType = `-- name: GetCategoryType :one
SELECT id, name FROM category_types WHERE id = $1
`

func (r *categoryTypeRepositoryImpl) Get(ctx context.Context, id int32) (CategoryTypeModel, error) {
	row := r.db.QueryRow(ctx, getCategoryType, id)
	var i CategoryTypeModel
	err := row.Scan(&i.ID, &i.Name)
	return i, err
}

const updateCategoryType = `-- name: UpdateCategoryType :exec
UPDATE category_types SET name = $2 WHERE id = $1 Returning id, name
`

func (r *categoryTypeRepositoryImpl) Update(ctx context.Context, id int32, name string) (bool, error) {
	result, err := r.db.Exec(ctx, updateCategoryType, id, name)
	return result.RowsAffected() > 0, err
}

const getCategoryTypes = `-- name: GetCategoryTypes :many
SELECT id, name FROM category_types ORDER BY id DESC LIMIT $1 OFFSET $2
`

func (r *categoryTypeRepositoryImpl) List(ctx context.Context, limit int64, offset int64) ([]CategoryTypeModel, error) {
	rows, err := r.db.Query(ctx, getCategoryTypes, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []CategoryTypeModel
	for rows.Next() {
		var i CategoryTypeModel
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
