package repository_test

import (
	"context"
	"testing"

	"github.com/hexley21/fixup/internal/catalog/domain"
	"github.com/hexley21/fixup/internal/catalog/repository"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
)

const (
	categoryName = "Maintenance"
	categoryId   = int32(1)
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

	categoryId, err := repo.Create(ctx, domain.CategoryInfo{TypeID: categoryType.ID, Name: categoryName})
	assert.NoError(t, err)
	assert.NotEmpty(t, categoryId)
}

func TestCreateCategory_InvalidArgs(t *testing.T) {
	ctx, pgPool, repo := setupCategory()
	defer cleanupPostgres(ctx, pgPool)

	categoryType, err := insertCategoryType(pgPool, ctx, categoryTypeName)
	if err != nil {
		t.Fatalf("failed to insert category type: %v", err)
	}

	categoryId, err := repo.Create(ctx, domain.CategoryInfo{TypeID: categoryType.ID, Name: ""})

	var pgErr *pgconn.PgError
	if assert.ErrorAs(t, err, &pgErr) {
		assert.Equal(t, pgerrcode.CheckViolation, pgErr.Code)
	}
	assert.Empty(t, categoryId)
}

func TestCreateCategory_NonexistentType(t *testing.T) {
	ctx, pgPool, repo := setupCategory()
	defer cleanupPostgres(ctx, pgPool)

	categoryId, err := repo.Create(ctx, domain.CategoryInfo{TypeID: 0, Name: categoryName})

	var pgErr *pgconn.PgError
	if assert.ErrorAs(t, err, &pgErr) {
		assert.Equal(t, pgerrcode.ForeignKeyViolation, pgErr.Code)
	}
	assert.Empty(t, categoryId)
}

func TestCreateCategory_Conflict(t *testing.T) {
	ctx, pgPool, repo := setupCategory()
	defer cleanupPostgres(ctx, pgPool)

	categoryType, err := insertCategoryType(pgPool, ctx, categoryTypeName)
	if err != nil {
		t.Fatalf("failed to insert category type: %v", err)
	}

	categoryId, err := repo.Create(ctx, domain.CategoryInfo{TypeID: categoryType.ID, Name: categoryName})
	assert.NoError(t, err)
	assert.NotEmpty(t, categoryId)

	categoryId, err = repo.Create(ctx, domain.CategoryInfo{TypeID: categoryType.ID, Name: categoryName})
	var pgErr *pgconn.PgError
	if assert.ErrorAs(t, err, &pgErr) {
		assert.Equal(t, pgerrcode.RaiseException, pgErr.Code)
	}
	assert.Empty(t, categoryId)
}

func TestDeleteCategory_Success(t *testing.T) {
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

	ok, err := repo.Delete(ctx, insert.ID)
	assert.NoError(t, err)
	assert.True(t, ok)

	category, err := getCategoryById(pgPool, ctx, insert.ID)
	assert.Empty(t, category)
	assert.ErrorIs(t, err, pgx.ErrNoRows)
}

func TestDeleteCategory_NotFound(t *testing.T) {
	ctx, pgPool, repo := setupCategory()
	defer cleanupPostgres(ctx, pgPool)

	ok, err := repo.Delete(ctx, categoryTypeId)
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestGetCategory_Success(t *testing.T) {
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

	category, err := repo.Get(ctx, insert.ID)
	assert.NoError(t, err)
	assert.Equal(t, categoryName, category.Name)
}

func TestGetCategory_NotFound(t *testing.T) {
	ctx, pgPool, repo := setupCategory()
	defer cleanupPostgres(ctx, pgPool)

	category, err := repo.Get(ctx, -1)
	assert.ErrorIs(t, err, pgx.ErrNoRows)
	assert.Empty(t, category.Name)
}

func TestListCategories_Success(t *testing.T) {
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

	entities, err := repo.List(ctx, 1, 0)

	assert.Equal(t, 1, len(entities))
	assert.Equal(t, categoryName, entities[0].Name)
	assert.NoError(t, err)
}

func TestListCategories_NotFound(t *testing.T) {
	ctx, pgPool, repo := setupCategory()
	defer cleanupPostgres(ctx, pgPool)

	entities, err := repo.List(ctx, 1, 0)

	assert.Equal(t, 0, len(entities))
	assert.NoError(t, err)
}

func TestUpdateCategory_Success(t *testing.T) {
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

	update, err := repo.Update(ctx, insert.ID, domain.CategoryInfo{TypeID: insert.TypeID, Name: insert.Name})

	assert.NoError(t, err)
	assert.Equal(t, insert.ID, update.ID)
	assert.Equal(t, insert.TypeID, update.TypeID)
	assert.Equal(t, insert.Name, update.Name)
}

func TestUpdateCategory_NotFound(t *testing.T) {
	ctx, pgPool, repo := setupCategory()
	defer cleanupPostgres(ctx, pgPool)

	update, err := repo.Update(ctx, categoryId, domain.CategoryInfo{Name: categoryTypeName})
	assert.ErrorIs(t, err, pgx.ErrNoRows)
	assert.Empty(t, update)
}

func TestUpdateCategory_Conflict(t *testing.T) {
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



	category, err := repo.Update(ctx, insert.ID, domain.CategoryInfo{TypeID: insert.TypeID, Name: "Fix"})

	var pgErr *pgconn.PgError
	if assert.ErrorAs(t, err, &pgErr) {
		assert.Equal(t, pgerrcode.RaiseException, pgErr.Code)
	}
	assert.Empty(t, category.ID)
	assert.Empty(t, category.Name)
	assert.Empty(t, category.TypeID)
}

func insertCategory(dbPool *pgxpool.Pool, ctx context.Context, typeId int32, name string) (repository.CategoryModel, error) {
	row := dbPool.QueryRow(ctx, "INSERT INTO categories (type_id, name) VALUES ($1, $2) RETURNING *", typeId, name)
	var i repository.CategoryModel
	err := row.Scan(&i.ID, &i.TypeID, &i.Name)
	return i, err
}

func getCategoryById(dbPool *pgxpool.Pool, ctx context.Context, id int32) (repository.CategoryModel, error) {
	row := dbPool.QueryRow(ctx, "SELECT * FROM categories WHERE id = $1", id)
	var i repository.CategoryModel
	err := row.Scan(&i.ID, &i.Name)
	return i, err
}
