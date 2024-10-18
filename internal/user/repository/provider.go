package repository

import (
	"context"

	"github.com/hexley21/fixup/pkg/infra/postgres"
)

type ProviderRepository interface {
	postgres.Repository[ProviderRepository]
	Create(ctx context.Context, arg CreateProviderParams) (bool, error)
	Get(ctx context.Context, userID int64) (Provider, error)
}

type pgsqlProviderRepository struct {
	db postgres.PGXQuerier
}

func NewProviderRepository(dbtx postgres.PGXQuerier) *pgsqlProviderRepository {
	return &pgsqlProviderRepository{
		dbtx,
	}
}

func (r *pgsqlProviderRepository) WithTx(tx postgres.PGXQuerier) ProviderRepository {
	return NewProviderRepository(tx)
}

const createProvider = `-- name: CreateProvider :exec
INSERT INTO providers (
  personal_id_number, personal_id_preview, user_id
) VALUES (
  $1, $2, $3
)
`

type CreateProviderParams struct {
	PersonalIDNumber  []byte `json:"personal_id_number"`
	PersonalIDPreview string `json:"personal_id_preview"`
	UserID            int64  `json:"user_id"`
}

func (r *pgsqlProviderRepository) Create(ctx context.Context, arg CreateProviderParams) (bool, error) {
	result, err := r.db.Exec(ctx, createProvider, arg.PersonalIDNumber, arg.PersonalIDPreview, arg.UserID)
	return result.RowsAffected() > 0, err
}

const getByUserId = `-- name: GetByUserId :one
SELECT 
  personal_id_number, 
  personal_id_preview, 
  user_id 
FROM 
  providers
WHERE 
  user_id = $1
`

func (r *pgsqlProviderRepository) Get(ctx context.Context, userID int64) (Provider, error) {
	row := r.db.QueryRow(ctx, getByUserId, userID)
	var i Provider
	err := row.Scan(&i.PersonalIDNumber, &i.PersonalIDPreview, &i.UserID)
	return i, err
}
