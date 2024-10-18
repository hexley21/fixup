package repository_test

import (
	"context"
	"testing"

	"github.com/hexley21/fixup/internal/catalog/repository"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
)

const (
	categoryTypeName = "Home"
	categoryTypeId   = int32(1)
)

func setupCategoryType() (
	ctx context.Context,
	pgPool *pgxpool.Pool,
	repo repository.CategoryTypeRepository,
) {
	ctx = context.Background()

	pgPool = getPgPool(ctx)
	repo = repository.NewCategoryTypeRepository(pgPool)

	return
}

func TestCreateCategoryType_Success(t *testing.T) {
	ctx, pgPool, repo := setupCategoryType()
	defer cleanupPostgres(ctx, pgPool)

	categoryType, err := repo.Create(ctx, categoryTypeName)
	assert.NoError(t, err)
	assert.NotEmpty(t, categoryType.ID)
	assert.Equal(t, categoryTypeName, categoryType.Name)
}

func TestCreateCategoryType_InvalidArgs(t *testing.T) {
	ctx, pgPool, repo := setupCategoryType()
	defer cleanupPostgres(ctx, pgPool)

	categoryType, err := repo.Create(ctx, "")

	var pgErr *pgconn.PgError
	if assert.ErrorAs(t, err, &pgErr) {
		assert.Equal(t, pgerrcode.CheckViolation, pgErr.Code)
	}
	assert.Empty(t, categoryType.ID)
	assert.Empty(t, categoryType.Name)
}

func TestCreateCategoryType_Conflict(t *testing.T) {
	ctx, pgPool, repo := setupCategoryType()
	defer cleanupPostgres(ctx, pgPool)

	categoryType, err := repo.Create(ctx, categoryTypeName)
	assert.NoError(t, err)
	assert.NotEmpty(t, categoryType.ID)
	assert.Equal(t, categoryTypeName, categoryType.Name)

	categoryType, err = repo.Create(ctx, categoryTypeName)

	var pgErr *pgconn.PgError
	if assert.ErrorAs(t, err, &pgErr) {
		assert.Equal(t, pgerrcode.UniqueViolation, pgErr.Code)
	}
	assert.Empty(t, categoryType.ID)
	assert.Empty(t, categoryType.Name)
}

func TestDeleteCategoryType_Success(t *testing.T) {
	ctx, pgPool, repo := setupCategoryType()
	defer cleanupPostgres(ctx, pgPool)

	insert, err := insertCategoryType(pgPool, ctx, categoryTypeName)
	if err != nil {
		t.Fatalf("failed to insert category type: %v", err)
	}

	ok, err := repo.Delete(ctx, insert.ID)
	assert.NoError(t, err)

	categoryType, err := getCategoryTypeById(pgPool, ctx, insert.ID)
	assert.Empty(t, categoryType)
	assert.ErrorIs(t, err, pgx.ErrNoRows)
	assert.True(t, ok)
}

func TestDeleteCategoryType_NotFound(t *testing.T) {
	ctx, pgPool, repo := setupCategoryType()
	defer cleanupPostgres(ctx, pgPool)

	ok, err := repo.Delete(ctx, categoryTypeId)
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestGetCategoryType_Success(t *testing.T) {
	ctx, pgPool, repo := setupCategoryType()
	defer cleanupPostgres(ctx, pgPool)

	insert, err := insertCategoryType(pgPool, ctx, categoryTypeName)
	if err != nil {
		t.Fatalf("failed to insert category type: %v", err)
	}

	categoryType, err := repo.Get(ctx, insert.ID)
	assert.NoError(t, err)
	assert.Equal(t, categoryTypeName, categoryType.Name)
}

func TestGetCategoryType_NotFound(t *testing.T) {
	ctx, pgPool, repo := setupCategoryType()
	defer cleanupPostgres(ctx, pgPool)

	categoryType, err := repo.Get(ctx, -1)
	assert.ErrorIs(t, err, pgx.ErrNoRows)
	assert.Empty(t, categoryType.Name)
}

func TestListCategoryTypes_Success(t *testing.T) {
	ctx, pgPool, repo := setupCategoryType()
	defer cleanupPostgres(ctx, pgPool)

	_, err := insertCategoryType(pgPool, ctx, categoryTypeName)
	if err != nil {
		t.Fatalf("failed to insert category type: %v", err)
	}

	entities, err := repo.List(ctx, 1, 0)

	assert.Equal(t, 1, len(entities))
	assert.Equal(t, categoryTypeName, entities[0].Name)
	assert.NoError(t, err)
}

func TestListCategoryTypes_NotFound(t *testing.T) {
	ctx, pgPool, repo := setupCategoryType()
	defer cleanupPostgres(ctx, pgPool)

	entities, err := repo.List(ctx, 1, 0)

	assert.Equal(t, 0, len(entities))
	assert.NoError(t, err)
}

func TestUpdateCategoryType_Success(t *testing.T) {
	ctx, pgPool, repo := setupCategoryType()
	defer cleanupPostgres(ctx, pgPool)

	insert, err := insertCategoryType(pgPool, ctx, categoryTypeName)
	if err != nil {
		t.Fatalf("failed to insert category type: %v", err)
	}

	ok, err := repo.Update(ctx, insert.ID, insert.Name)
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestUpdateCategoryType_NotFound(t *testing.T) {
	ctx, pgPool, repo := setupCategoryType()
	defer cleanupPostgres(ctx, pgPool)

	ok, err := repo.Update(ctx, categoryTypeId, categoryTypeName)

	assert.NoError(t, err)
	assert.False(t, ok)
}

func insertCategoryType(dbPool *pgxpool.Pool, ctx context.Context, name string) (repository.CategoryTypeModel, error) {
	row := dbPool.QueryRow(ctx, "INSERT INTO category_types (name) VALUES ($1) RETURNING id, name", name)
	var i repository.CategoryTypeModel
	err := row.Scan(&i.ID, &i.Name)
	return i, err
}

func getCategoryTypeById(dbPool *pgxpool.Pool, ctx context.Context, id int32) (repository.CategoryTypeModel, error) {
	row := dbPool.QueryRow(ctx, "SELECT id, name FROM category_types WHERE id = $1", id)
	var i repository.CategoryTypeModel
	err := row.Scan(&i.ID, &i.Name)
	return i, err
}
