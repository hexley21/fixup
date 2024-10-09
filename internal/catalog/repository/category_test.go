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
	categoryName = "Maintenance"
)

func setupCategory() (
	ctx context.Context,
	pgPool *pgxpool.Pool,
	repo repository.CategoryRepository,
) {
	ctx = context.Background()

	pgPool = getPgPool(ctx)
	repo = repository.NewCategoryRepository(pgPool)

	return
}

func TestCreateCategory_Success(t *testing.T) {
	ctx, pgPool, repo := setupCategory()
	defer cleanupPostgres(ctx, pgPool)

	categoryType, err := insertCategoryType(pgPool, ctx, categoryTypeName)
	if err != nil {
		t.Fatalf("failed to insert category type: %v", err)
	}

	entity, err := repo.CreateCategory(ctx, repository.CreateCategoryParams{TypeID: categoryType.ID, Name: categoryName})
	assert.NoError(t, err)
	assert.NotEmpty(t, entity.ID)
	assert.Equal(t, categoryName, entity.Name)
}

func TestCreateCategory_InvalidArgs(t *testing.T) {
	ctx, pgPool, repo := setupCategory()
	defer cleanupPostgres(ctx, pgPool)

	categoryType, err := insertCategoryType(pgPool, ctx, categoryTypeName)
	if err != nil {
		t.Fatalf("failed to insert category type: %v", err)
	}

	entity, err := repo.CreateCategory(ctx, repository.CreateCategoryParams{TypeID: categoryType.ID, Name: ""})

	var pgErr *pgconn.PgError
	if assert.ErrorAs(t, err, &pgErr) {
		assert.Equal(t, pgerrcode.CheckViolation, pgErr.Code)
	}
	assert.Empty(t, entity.ID)
	assert.Empty(t, entity.Name)
}

func TestCreateCategory_NonexistentType(t *testing.T) {
	ctx, pgPool, repo := setupCategory()
	defer cleanupPostgres(ctx, pgPool)

	entity, err := repo.CreateCategory(ctx, repository.CreateCategoryParams{TypeID: 0, Name: categoryName})

	var pgErr *pgconn.PgError
	if assert.ErrorAs(t, err, &pgErr) {
		assert.Equal(t, pgerrcode.ForeignKeyViolation, pgErr.Code)
	}
	assert.Empty(t, entity.ID)
	assert.Empty(t, entity.Name)
}

func TestCreateCategory_Conflict(t *testing.T) {
	ctx, pgPool, repo := setupCategory()
	defer cleanupPostgres(ctx, pgPool)

	categoryType, err := insertCategoryType(pgPool, ctx, categoryTypeName)
	if err != nil {
		t.Fatalf("failed to insert category type: %v", err)
	}

	entity, err := repo.CreateCategory(ctx, repository.CreateCategoryParams{TypeID: categoryType.ID, Name: categoryName})
	assert.NoError(t, err)
	assert.NotEmpty(t, entity.ID)
	assert.Equal(t, categoryName, entity.Name)

	entity, err = repo.CreateCategory(ctx, repository.CreateCategoryParams{TypeID: categoryType.ID, Name: categoryName})
	var pgErr *pgconn.PgError
	if assert.ErrorAs(t, err, &pgErr) {
		assert.Equal(t, pgerrcode.RaiseException, pgErr.Code)
	}
	assert.Empty(t, entity.ID)
	assert.Empty(t, entity.Name)
}

func TestDeleteCategoryById_Success(t *testing.T) {
	ctx, pgPool, repo := setupCategory()
	defer cleanupPostgres(ctx, pgPool)

	insertCategoryType, err := insertCategoryType(pgPool, ctx, categoryTypeName)
	if err != nil {
		t.Fatalf("failed to insert category type: %v", err)
	}

	insert, err := insertCategory(pgPool, ctx, insertCategoryType.ID, categoryName)
	if err != nil {
		t.Fatalf("failed to insert category: %v", err)
	}

	err = repo.DeleteCategoryById(ctx, insert.ID)
	assert.NoError(t, err)

	entity, err := getCategoryById(pgPool, ctx, insert.ID)
	assert.Empty(t, entity)
	assert.ErrorIs(t, err, pgx.ErrNoRows)
}

func TestDeleteCategoryById_NotFound(t *testing.T) {
	ctx, pgPool, repo := setupCategory()
	defer cleanupPostgres(ctx, pgPool)

	err := repo.DeleteCategoryById(ctx, categoryTypeId)
	assert.ErrorIs(t, err, pg_error.ErrNotFound)
}

func TestGetCategoryById_Success(t *testing.T) {
	ctx, pgPool, repo := setupCategory()
	defer cleanupPostgres(ctx, pgPool)

	insertCategoryType, err := insertCategoryType(pgPool, ctx, categoryTypeName)
	if err != nil {
		t.Fatalf("failed to insert category type: %v", err)
	}

	insert, err := insertCategory(pgPool, ctx, insertCategoryType.ID, categoryName)
	if err != nil {
		t.Fatalf("failed to insert category: %v", err)
	}

	entity, err := repo.GetCategoryById(ctx, insert.ID)
	assert.NoError(t, err)
	assert.Equal(t, categoryName, entity.Name)
}

func TestGetCategoryById_NotFound(t *testing.T) {
	ctx, pgPool, repo := setupCategory()
	defer cleanupPostgres(ctx, pgPool)

	entity, err := repo.GetCategoryById(ctx, -1)
	assert.ErrorIs(t, err, pgx.ErrNoRows)
	assert.Empty(t, entity.Name)
}

func TestGetCategories_Success(t *testing.T) {
	ctx, pgPool, repo := setupCategory()
	defer cleanupPostgres(ctx, pgPool)

	insertCategoryType, err := insertCategoryType(pgPool, ctx, categoryTypeName)
	if err != nil {
		t.Fatalf("failed to insert category type: %v", err)
	}

	_, err = insertCategory(pgPool, ctx, insertCategoryType.ID, categoryName)
	if err != nil {
		t.Fatalf("failed to insert category: %v", err)
	}

	entities, err := repo.GetCategories(ctx, 0, 1)

	assert.Equal(t, 1, len(entities))
	assert.Equal(t, categoryName, entities[0].Name)
	assert.NoError(t, err)
}

func TestGetCategories_NotFound(t *testing.T) {
	ctx, pgPool, repo := setupCategory()
	defer cleanupPostgres(ctx, pgPool)

	entities, err := repo.GetCategories(ctx, 0, 1)

	assert.Equal(t, 0, len(entities))
	assert.NoError(t, err)
}

func TestUpdateCategoryById_Success(t *testing.T) {
	ctx, pgPool, repo := setupCategory()
	defer cleanupPostgres(ctx, pgPool)

	insertCategoryType, err := insertCategoryType(pgPool, ctx, categoryTypeName)
	if err != nil {
		t.Fatalf("failed to insert category type: %v", err)
	}

	insert, err := insertCategory(pgPool, ctx, insertCategoryType.ID, categoryName)
	if err != nil {
		t.Fatalf("failed to insert category: %v", err)
	}

	update, err := repo.UpdateCategoryById(ctx, repository.UpdateCategoryByIdParams{ID: insert.ID, TypeID: insert.TypeID, Name: insert.Name})

	assert.NoError(t, err)
	assert.Equal(t, insert.ID, update.ID)
	assert.Equal(t, insert.TypeID, update.TypeID)
	assert.Equal(t, insert.Name, update.Name)
}

func TestUpdateCategoryById_NotFound(t *testing.T) {
	ctx, pgPool, repo := setupCategory()
	defer cleanupPostgres(ctx, pgPool)

	update, err := repo.UpdateCategoryById(ctx, repository.UpdateCategoryByIdParams{ID: categoryTypeId, Name: categoryTypeName})

	assert.ErrorIs(t, err, pgx.ErrNoRows)
	assert.Empty(t, update)
}

func TestUpdateCategoryById_Conflict(t *testing.T) {
	ctx, pgPool, repo := setupCategory()
	defer cleanupPostgres(ctx, pgPool)

	insertCategoryType, err := insertCategoryType(pgPool, ctx, categoryTypeName)
	if err != nil {
		t.Fatalf("failed to insert category type: %v", err)
	}

	insert, err := insertCategory(pgPool, ctx, insertCategoryType.ID, categoryName)
	if err != nil {
		t.Fatalf("failed to insert category: %v", err)
	}

	_, err = insertCategory(pgPool, ctx, insertCategoryType.ID, "Fix")
	if err != nil {
		t.Fatalf("failed to insert category: %v", err)
	}

	entity, err := repo.UpdateCategoryById(ctx, repository.UpdateCategoryByIdParams{ID: insert.ID, TypeID: insert.TypeID, Name: "Fix"})

	var pgErr *pgconn.PgError
	if assert.ErrorAs(t, err, &pgErr) {
		assert.Equal(t, pgerrcode.RaiseException, pgErr.Code)
	}
	assert.Empty(t, entity.ID)
	assert.Empty(t, entity.Name)
	assert.Empty(t, entity.TypeID)
}

func insertCategory(dbPool *pgxpool.Pool, ctx context.Context, typeId int32, name string) (entity.Category, error) {
	row := dbPool.QueryRow(ctx, "INSERT INTO categories (type_id, name) VALUES ($1, $2) RETURNING *", typeId, name)
	var i entity.Category
	err := row.Scan(&i.ID, &i.TypeID, &i.Name)
	return i, err
}

func getCategoryById(dbPool *pgxpool.Pool, ctx context.Context, id int32) (entity.Category, error) {
	row := dbPool.QueryRow(ctx, "SELECT * FROM categories WHERE id = $1", id)
	var i entity.Category
	err := row.Scan(&i.ID, &i.Name)
	return i, err
}
