package repository_test

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/hexley21/fixup/internal/user/entity"
	"github.com/hexley21/fixup/internal/user/repository"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestCreateProvider(t *testing.T) {
	ctx := context.Background()
	dbPool := getDbPool(ctx)
	setupDatabaseCleanup(t, ctx, dbPool)

	repo := repository.NewProviderRepository(dbPool)

	user, err := insertUser(dbPool, ctx, userCreateArgs, 1)
	if err != nil {
		assert.FailNow(t, fmt.Sprintf("failed to insert user: %v", err))
	}

	args := repository.CreateProviderParams{
		PersonalIDNumber:  []byte("123456789"),
		PersonalIDPreview: "12345",
		UserID:            user.ID,
	}

	assert.NoError(t, repo.Create(ctx, args))

	row := dbPool.QueryRow(ctx, "SELECT * FROM  providers WHERE user_id = $1", args.UserID)
	var p entity.Provider
	err = row.Scan(&p.PersonalIDNumber, &p.PersonalIDPreview, &p.UserID)
	assert.NoError(t, err)

	assert.Equal(t, p.PersonalIDNumber, args.PersonalIDNumber)
	assert.Equal(t, p.PersonalIDPreview, args.PersonalIDPreview)
	assert.Equal(t, p.UserID, args.UserID)
}

func TestCreateProviderWithInvalidArgs(t *testing.T) {
	ctx := context.Background()
	dbPool := getDbPool(ctx)
	setupDatabaseCleanup(t, ctx, dbPool)

	repo := repository.NewProviderRepository(dbPool)

	user, err := insertUser(dbPool, ctx, userCreateArgs, 1)
	if err != nil {
		assert.FailNow(t, fmt.Sprintf("failed to insert user: %v", err))
	}

	invalidArgs := []repository.CreateProviderParams{
		{PersonalIDPreview: "12345", UserID: user.ID},
		{PersonalIDNumber: []byte("123456789"), PersonalIDPreview: "1234567", UserID: user.ID},
	}

	i := 0
	for _, args := range invalidArgs {
		err := repo.Create(ctx, args)
		if !assert.Error(t, err) {
			log.Println("create provider:", i)
		}
		i++
	}
}

func TestCreateProviderForNonexistentUser(t *testing.T) {
	ctx := context.Background()
	dbPool := getDbPool(ctx)
	setupDatabaseCleanup(t, ctx, dbPool)

	repo := repository.NewProviderRepository(dbPool)

	args := repository.CreateProviderParams{
		PersonalIDNumber:  []byte("123456789"),
		PersonalIDPreview: "12345",
		UserID:            1,
	}

	err := repo.Create(ctx, args)
	var pgErr *pgconn.PgError
	if assert.ErrorAs(t, err, &pgErr) {
		assert.Equal(t, pgErr.Code, pgerrcode.ForeignKeyViolation)
	}

	row := dbPool.QueryRow(ctx, "SELECT * FROM  providers WHERE user_id = $1", args.UserID)
	var p entity.Provider
	err = row.Scan(&p.PersonalIDNumber, &p.PersonalIDPreview, &p.UserID)
	assert.ErrorIs(t, err, pgx.ErrNoRows)
}

func TestGetByUserId(t *testing.T) {
	ctx := context.Background()
	dbPool := getDbPool(ctx)
	setupDatabaseCleanup(t, ctx, dbPool)

	repo := repository.NewProviderRepository(dbPool)

	user, err := insertUser(dbPool, ctx, userCreateArgs, 1)
	if err != nil {
		assert.FailNow(t, fmt.Sprintf("failed to insert user: %v", err))
	}

	row := dbPool.QueryRow(
		ctx,
		"INSERT INTO providers (personal_id_number, personal_id_preview, user_id) VALUES ($1, $2, $3) RETURNING *",
		[]byte("123456789"),
		"12345",
		user.ID,
	)
	var provider entity.Provider
	err = row.Scan(
		&provider.PersonalIDNumber,
		&provider.PersonalIDPreview,
		&provider.UserID,
	)
	if err != nil {
		assert.FailNow(t, fmt.Sprintf("failed to insert provider: %v", err))
	}

	res, err := repo.GetByUserId(ctx, provider.UserID)
	assert.NoError(t, err)

	assert.Equal(t, res.PersonalIDNumber, provider.PersonalIDNumber)
	assert.Equal(t, res.PersonalIDPreview, provider.PersonalIDPreview)
	assert.Equal(t, res.UserID, provider.UserID)
}

func TestGetByUserIdWithNonexistentProvider(t *testing.T) {
	ctx := context.Background()
	dbPool := getDbPool(ctx)
	setupDatabaseCleanup(t, ctx, dbPool)

	repo := repository.NewProviderRepository(dbPool)

	user, err := insertUser(dbPool, ctx, userCreateArgs, 1)
	if err != nil {
		assert.FailNow(t, fmt.Sprintf("failed to insert user: %v", err))
	}

	res, err := repo.GetByUserId(ctx, user.ID)
	assert.ErrorIs(t, err, pgx.ErrNoRows)
	assert.Empty(t, res)
}
