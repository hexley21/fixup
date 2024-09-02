package repository

import (
	"context"

	"github.com/hexley21/handy/pkg/infra/postgres"
)

type ProviderRepository interface {
	postgres.Repository[ProviderRepository]
	CreateProvider(ctx context.Context, arg CreateProviderParams) error
}

type providerRepositoryImpl struct {
	db postgres.DBTX
}

func NewProviderRepository(dbtx postgres.DBTX) ProviderRepository {
	return &providerRepositoryImpl{
		dbtx,
	}
}

func (r providerRepositoryImpl) WithTx(tx postgres.DBTX) ProviderRepository {
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

func (r *providerRepositoryImpl) CreateProvider(ctx context.Context, arg CreateProviderParams) error {
	_, err := r.db.Exec(ctx, createProvider, arg.PersonalIDNumber, arg.PersonalIDPreview, arg.UserID)
	return err
}
