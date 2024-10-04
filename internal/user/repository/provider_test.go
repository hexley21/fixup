package repository_test

import (
	"context"
	"log"
	"testing"

	"github.com/hexley21/fixup/internal/user/entity"
	"github.com/hexley21/fixup/internal/user/repository"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestCreateProvider_Success(t *testing.T) {
	ctx := context.Background()
	dbPool := getPgPool(ctx)
	defer cleanupPostgres(ctx, dbPool)

	repo := repository.NewProviderRepository(dbPool)

	user, err := insertUser(dbPool, ctx, userCreateArgs, 1)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
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

	assert.Equal(t, args.PersonalIDNumber, p.PersonalIDNumber)
	assert.Equal(t, args.PersonalIDPreview, p.PersonalIDPreview)
	assert.Equal(t, args.UserID, p.UserID)
}

func TestCreateProvider_InvalidArguments(t *testing.T) {
	ctx := context.Background()
	dbPool := getPgPool(ctx)
	defer cleanupPostgres(ctx, dbPool)

	repo := repository.NewProviderRepository(dbPool)

	user, err := insertUser(dbPool, ctx, userCreateArgs, 1)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
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

func TestCreateProvider_NotFound(t *testing.T) {
	ctx := context.Background()
	dbPool := getPgPool(ctx)
	defer cleanupPostgres(ctx, dbPool)

	repo := repository.NewProviderRepository(dbPool)

	args := repository.CreateProviderParams{
		PersonalIDNumber:  []byte("123456789"),
		PersonalIDPreview: "12345",
		UserID:            1,
	}

	err := repo.Create(ctx, args)
	var pgErr *pgconn.PgError
	if assert.ErrorAs(t, err, &pgErr) {
		assert.Equal(t, pgerrcode.ForeignKeyViolation, pgErr.Code)
	}

	row := dbPool.QueryRow(ctx, "SELECT * FROM  providers WHERE user_id = $1", args.UserID)
	var p entity.Provider
	err = row.Scan(&p.PersonalIDNumber, &p.PersonalIDPreview, &p.UserID)
	assert.ErrorIs(t, err, pgx.ErrNoRows)
}

func TestGetByUserId_Success(t *testing.T) {
	ctx := context.Background()
	dbPool := getPgPool(ctx)
	defer cleanupPostgres(ctx, dbPool)

	repo := repository.NewProviderRepository(dbPool)

	user, err := insertUser(dbPool, ctx, userCreateArgs, 1)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	row := dbPool.QueryRow(
		ctx,
		"INSERT INTO providers (personal_id_number, personal_id_preview, user_id) VALUES ($1, $2, $3) RETURNING *",
		[]byte("123456789"),
		"12345",
		user.ID,
	)
	var selection entity.Provider
	err = row.Scan(
		&selection.PersonalIDNumber,
		&selection.PersonalIDPreview,
		&selection.UserID,
	)
	if err != nil {
		t.Fatalf("failed to insert provider: %v", err)
	}

	provider, err := repo.GetByUserId(ctx, selection.UserID)
	assert.NoError(t, err)

	assert.Equal(t, selection.PersonalIDNumber, provider.PersonalIDNumber)
	assert.Equal(t, selection.PersonalIDPreview, provider.PersonalIDPreview)
	assert.Equal(t, selection.UserID, provider.UserID)
}

func TestGetByUserId_NotFound(t *testing.T) {
	ctx := context.Background()
	dbPool := getPgPool(ctx)
	defer cleanupPostgres(ctx, dbPool)

	repo := repository.NewProviderRepository(dbPool)

	user, err := insertUser(dbPool, ctx, userCreateArgs, 1)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	provider, err := repo.GetByUserId(ctx, user.ID)
	assert.ErrorIs(t, err, pgx.ErrNoRows)
	assert.Empty(t, provider)
}
