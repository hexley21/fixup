package repository_test

import (
	"context"
	"testing"

	"github.com/hexley21/fixup/internal/catalog/entity"
	"github.com/hexley21/fixup/internal/catalog/repository"
	"github.com/hexley21/fixup/pkg/infra/postgres/pg_error"
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

	categoryType, err := repo.CreateCategoryType(ctx, categoryTypeName)
	assert.NoError(t, err)
	assert.NotEmpty(t, categoryType.ID)
	assert.Equal(t, categoryTypeName, categoryType.Name)
}

func TestCreateCategoryType_InvalidArgs(t *testing.T) {
	ctx, pgPool, repo := setupCategoryType()
	defer cleanupPostgres(ctx, pgPool)

	categoryType, err := repo.CreateCategoryType(ctx, "")

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

	categoryType, err := repo.CreateCategoryType(ctx, categoryTypeName)
	assert.NoError(t, err)
	assert.NotEmpty(t, categoryType.ID)
	assert.Equal(t, categoryTypeName, categoryType.Name)

	categoryType, err = repo.CreateCategoryType(ctx, categoryTypeName)

	var pgErr *pgconn.PgError
	if assert.ErrorAs(t, err, &pgErr) {
		assert.Equal(t, pgerrcode.UniqueViolation, pgErr.Code)
	}
	assert.Empty(t, categoryType.ID)
	assert.Empty(t, categoryType.Name)
}

func TestDeleteCategoryTypeById_Success(t *testing.T) {
	ctx, pgPool, repo := setupCategoryType()
	defer cleanupPostgres(ctx, pgPool)

	insert, err := insertCategoryType(pgPool, ctx, categoryTypeName)
	if err != nil {
		t.Fatalf("failed to insert category type: %v", err)
	}

	err = repo.DeleteCategoryTypeById(ctx, insert.ID)
	assert.NoError(t, err)

	categoryType, err := getCategoryTypeById(pgPool, ctx, insert.ID)
	assert.Empty(t, categoryType)
	assert.ErrorIs(t, err, pgx.ErrNoRows)
}

func TestDeleteCategoryTypeById_NotFound(t *testing.T) {
	ctx, pgPool, repo := setupCategoryType()
	defer cleanupPostgres(ctx, pgPool)

	err := repo.DeleteCategoryTypeById(ctx, categoryTypeId)
	assert.ErrorIs(t, err, pg_error.ErrNotFound)
}

func TestGetCategoryTypeById_Success(t *testing.T) {
	ctx, pgPool, repo := setupCategoryType()
	defer cleanupPostgres(ctx, pgPool)

	insert, err := insertCategoryType(pgPool, ctx, categoryTypeName)
	if err != nil {
		t.Fatalf("failed to insert category type: %v", err)
	}

	categoryType, err := repo.GetCategoryTypeById(ctx, insert.ID)
	assert.NoError(t, err)
	assert.Equal(t, categoryTypeName, categoryType.Name)
}

func TestGetCategoryTypeById_NotFound(t *testing.T) {
	ctx, pgPool, repo := setupCategoryType()
	defer cleanupPostgres(ctx, pgPool)

	categoryType, err := repo.GetCategoryTypeById(ctx, -1)
	assert.ErrorIs(t, err, pgx.ErrNoRows)
	assert.Empty(t, categoryType.Name)
}

func TestGetCategoryTypes_Success(t *testing.T) {
	ctx, pgPool, repo := setupCategoryType()
	defer cleanupPostgres(ctx, pgPool)

	_, err := insertCategoryType(pgPool, ctx, categoryTypeName)
	if err != nil {
		t.Fatalf("failed to insert category type: %v", err)
	}

	entities, err := repo.GetCategoryTypes(ctx, 0, 1)

	assert.Equal(t, 1, len(entities))
	assert.Equal(t, categoryTypeName, entities[0].Name)
	assert.NoError(t, err)
}

func TestGetCategoryTypes_NotFound(t *testing.T) {
	ctx, pgPool, repo := setupCategoryType()
	defer cleanupPostgres(ctx, pgPool)

	entities, err := repo.GetCategoryTypes(ctx, 0, 1)

	assert.Equal(t, 0, len(entities))
	assert.NoError(t, err)
}

func TestUpdateCategoryTypeById_Success(t *testing.T) {
	ctx, pgPool, repo := setupCategoryType()
	defer cleanupPostgres(ctx, pgPool)

	insert, err := insertCategoryType(pgPool, ctx, categoryTypeName)
	if err != nil {
		t.Fatalf("failed to insert category type: %v", err)
	}

	err = repo.UpdateCategoryTypeById(ctx, repository.UpdateCategoryTypeByIdParams{ID: insert.ID, Name: insert.Name})

	assert.NoError(t, err)
}

func TestUpdateCategoryTypeById_NotFound(t *testing.T) {
	ctx, pgPool, repo := setupCategoryType()
	defer cleanupPostgres(ctx, pgPool)

	err := repo.UpdateCategoryTypeById(ctx, repository.UpdateCategoryTypeByIdParams{ID: categoryTypeId, Name: categoryTypeName})

	assert.ErrorIs(t, err, pg_error.ErrNotFound)
}

func insertCategoryType(dbPool *pgxpool.Pool, ctx context.Context, name string) (entity.CategoryType, error) {
	row := dbPool.QueryRow(ctx, "INSERT INTO category_types (name) VALUES ($1) RETURNING id, name", name)
	var i entity.CategoryType
	err := row.Scan(&i.ID, &i.Name)
	return i, err
}

func getCategoryTypeById(dbPool *pgxpool.Pool, ctx context.Context, id int32) (entity.CategoryType, error) {
	row := dbPool.QueryRow(ctx, "SELECT id, name FROM category_types WHERE id = $1", id)
	var i entity.CategoryType
	err := row.Scan(&i.ID, &i.Name)
	return i, err
}
