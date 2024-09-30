package repository

import (
	"context"

	"github.com/hexley21/fixup/internal/user/entity"
	"github.com/hexley21/fixup/pkg/infra/postgres"
)

type ProviderRepository interface {
	postgres.Repository[ProviderRepository]
	Create(ctx context.Context, arg CreateProviderParams) error
	GetByUserId(ctx context.Context, userID int64) (entity.Provider, error)
}

type providerRepositoryImpl struct {
	db postgres.PGXQuerier
}

func NewProviderRepository(dbtx postgres.PGXQuerier) *providerRepositoryImpl {
	return &providerRepositoryImpl{
		dbtx,
	}
}

func (r *providerRepositoryImpl) WithTx(tx postgres.PGXQuerier) ProviderRepository {
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
	PersonalIDNumber  []byte
	PersonalIDPreview string
	UserID            int64
}

func (r *providerRepositoryImpl) Create(ctx context.Context, arg CreateProviderParams) error {
	_, err := r.db.Exec(ctx, createProvider, arg.PersonalIDNumber, arg.PersonalIDPreview, arg.UserID)
	return err
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

func (r *providerRepositoryImpl) GetByUserId(ctx context.Context, userID int64) (entity.Provider, error) {
	row := r.db.QueryRow(ctx, getByUserId, userID)
	var i entity.Provider
	err := row.Scan(&i.PersonalIDNumber, &i.PersonalIDPreview, &i.UserID)
	return i, err
}
