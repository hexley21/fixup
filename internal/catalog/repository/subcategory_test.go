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
	subcategoryName1 = "Door fix"
	subcategoryName2 = "AC Montage"
	subcategoryId    = int32(1)
)

func setupSubategory() (
	ctx context.Context,
	pgPool *pgxpool.Pool,
	repo repository.SubcategoryRepository,
) {
	ctx = context.Background()

	pgPool = getPgPool(ctx)
	repo = repository.NewSubcategoryRepository(pgPool)

	return
}

func TestCreateSubcategory(t *testing.T) {
	ctx, pgPool, repo := setupSubategory()
	defer cleanupPostgres(ctx, pgPool)

	category := &entity.Category{}

	tests := []struct {
		name             string
		subcategoryName1 string
		setup            func()
		expectedCode     string
	}{
		{
			name:             "Nonexistend Category Type",
			setup:            func() {},
			subcategoryName1: subcategoryName1,
			expectedCode:     pgerrcode.ForeignKeyViolation,
		},
		{
			name:             "Success",
			subcategoryName1: subcategoryName1,
			setup: func() {
				ct, err := insertCategoryType(pgPool, ctx, categoryTypeName)
				if err != nil {
					t.Fatalf("failed to insert category type: %v", err)
				}

				c, err := insertCategory(pgPool, ctx, ct.ID, categoryName)
				if err != nil {
					t.Fatalf("failed to insert category: %v", err)
				}

				category = &c
			},
		},
		{
			name:         "InvalidArgs",
			setup:        func() {},
			expectedCode: pgerrcode.CheckViolation,
		},
		{
			name:             "Conflict",
			subcategoryName1: subcategoryName1,
			setup: func() {
				_, _ = repo.CreateSubcategory(ctx, repository.CreateSubcategoryParams{Name: subcategoryName1, CategoryID: category.ID})
			},
			expectedCode: pgerrcode.RaiseException,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			subcategory, err := repo.CreateSubcategory(ctx, repository.CreateSubcategoryParams{Name: tt.subcategoryName1, CategoryID: category.ID})

			if tt.expectedCode == "" {
				assert.NoError(t, err)
				assert.NotEmpty(t, subcategory.ID)
				assert.Equal(t, tt.subcategoryName1, subcategory.Name)
				assert.Equal(t, category.ID, subcategory.CategoryID)
			} else {
				var pgErr *pgconn.PgError
				if assert.ErrorAs(t, err, &pgErr) {
					assert.Equal(t, tt.expectedCode, pgErr.Code)
				}
				assert.Empty(t, subcategory.ID)
				assert.Empty(t, subcategory.Name)
				assert.Empty(t, subcategory.CategoryID)
			}
		})
	}
}

func TestGetCategoryById(t *testing.T) {
	ctx, pgPool, repo := setupSubategory()
	defer cleanupPostgres(ctx, pgPool)

	_, category := insertSubcategoryDependencies(t, pgPool, ctx)

	tests := []struct {
		name          string
		setup         func() entity.Subcategory
		expectedError error
	}{
		{
			name: "Not Found",
			setup: func() entity.Subcategory {
				return entity.Subcategory{ID: -1}
			},
			expectedError: pgx.ErrNoRows,
		},
		{
			name: "Success",
			setup: func() entity.Subcategory {
				sc, err := insertSubcategory(pgPool, ctx, category.ID, subcategoryName1)
				if err != nil {
					t.Fatalf("failed to insert subcategory: %v", err)
				}
				return sc
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subcategoryInsert := tt.setup()

			subcategory, err := repo.GetSubcategoryById(ctx, subcategoryInsert.ID)

			if tt.expectedError == nil {
				assert.NoError(t, err)
				assert.Equal(t, subcategoryInsert.Name, subcategory.Name)
			} else {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Empty(t, subcategory.Name)
			}
		})
	}
}

func TestGetCategorise(t *testing.T) {
	ctx, pgPool, repo := setupSubategory()
	defer cleanupPostgres(ctx, pgPool)

	_, category := insertSubcategoryDependencies(t, pgPool, ctx)

	tests := []struct {
		name  string
		setup func()
		len   int
	}{
		{
			name:  "Not Found",
			setup: func() {},
			len:   0,
		},
		{
			name: "Success",
			setup: func() {
				_, _ = insertSubcategory(pgPool, ctx, category.ID, subcategoryName1)
				_, err := insertSubcategory(pgPool, ctx, category.ID, "AC Montage")
				if err != nil {
					t.Fatalf("failed to insert subcategory: %v", err)
				}
			},
			len: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			subcategories, err := repo.GetSubategories(ctx, 5, 0)

			assert.NoError(t, err)
			assert.Len(t, subcategories, tt.len)
		})
	}
}

func TestGetCategoriseByCategoryId(t *testing.T) {
	ctx, pgPool, repo := setupSubategory()
	defer cleanupPostgres(ctx, pgPool)

	_, category := insertSubcategoryDependencies(t, pgPool, ctx)

	tests := []struct {
		name       string
		setup      func()
		len        int
		categoryId int32
	}{
		{
			name:       "Not Found",
			setup:      func() {},
			len:        0,
			categoryId: category.ID,
		},
		{
			name: "Success",
			setup: func() {
				_, _ = insertSubcategory(pgPool, ctx, category.ID, subcategoryName1)
				_, err := insertSubcategory(pgPool, ctx, category.ID, subcategoryName2)
				if err != nil {
					t.Fatalf("failed to insert subcategory: %v", err)
				}
			},
			len:        2,
			categoryId: category.ID,
		},
		{
			name:       "nonexistend category",
			setup:      func() {},
			categoryId: -1,
			len:        0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			subcategories, err := repo.GetSubategoriesByCategoryId(ctx, tt.categoryId, 5, 0)

			assert.NoError(t, err)
			assert.Len(t, subcategories, tt.len)
		})
	}
}

func TestGetCategoriseByTypeId(t *testing.T) {
	ctx, pgPool, repo := setupSubategory()
	defer cleanupPostgres(ctx, pgPool)

	categoryType, category := insertSubcategoryDependencies(t, pgPool, ctx)

	tests := []struct {
		name   string
		setup  func()
		len    int
		typeId int32
	}{
		{
			name:   "Not Found",
			setup:  func() {},
			len:    0,
			typeId: categoryType.ID,
		},
		{
			name: "Success",
			setup: func() {
				_, _ = insertSubcategory(pgPool, ctx, category.ID, subcategoryName1)
				_, err := insertSubcategory(pgPool, ctx, category.ID, subcategoryName2)
				if err != nil {
					t.Fatalf("failed to insert subcategory: %v", err)
				}
			},
			len:    2,
			typeId: categoryType.ID,
		},
		{
			name:   "nonexistend category",
			setup:  func() {},
			typeId: -1,
			len:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			subcategories, err := repo.GetSubategoriesByTypeId(ctx, tt.typeId, 5, 0)

			assert.NoError(t, err)
			assert.Len(t, subcategories, tt.len)
		})
	}
}

func TestUpdateCategoryById(t *testing.T) {
	ctx, pgPool, repo := setupSubategory()
	defer cleanupPostgres(ctx, pgPool)

	_, category := insertSubcategoryDependencies(t, pgPool, ctx)

	tests := []struct {
		name              string
		setup             func() entity.Subcategory
		subcategoryName   string
		expectedError     error
		expectedErrorCode string
	}{
		{
			name: "Not Found",
			setup: func() entity.Subcategory {
				return entity.Subcategory{ID: -1}
			},
			expectedError: pgx.ErrNoRows,
		},
		{
			name: "Success",
			setup: func() entity.Subcategory {
				sc, err := insertSubcategory(pgPool, ctx, category.ID, subcategoryName1)
				if err != nil {
					t.Fatalf("failed to insert subcategory1: %v", err)
				}

				return sc
			},
			subcategoryName: subcategoryName2,
		},
		{
			name: "Conflict",
			setup: func() entity.Subcategory {
				sc, err := insertSubcategory(pgPool, ctx, category.ID, subcategoryName1)
				if err != nil {
					t.Fatalf("failed to insert subcategory2: %v", err)
				}

				return sc
			},
			subcategoryName:   subcategoryName2,
			expectedErrorCode: pgerrcode.RaiseException,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := tt.setup()

			subcategory, err := repo.UpdateSubcategoryById(ctx, repository.UpdateSubcategoryByIdParams{Name: tt.subcategoryName, CategoryID: sc.CategoryID, ID: sc.ID})

			if tt.expectedErrorCode != "" {
				var pgErr *pgconn.PgError
				if assert.ErrorAs(t, err, &pgErr) {
					assert.Equal(t, tt.expectedErrorCode, pgErr.Code)
				}
			} else if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Empty(t, subcategory.Name)

			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.subcategoryName, subcategory.Name)
			}
		})
	}
}

func TestDeleteCategoryById(t *testing.T) {
	ctx, pgPool, repo := setupSubategory()
	defer cleanupPostgres(ctx, pgPool)

	_, category := insertSubcategoryDependencies(t, pgPool, ctx)

	tests := []struct {
		name          string
		setup         func() int32
		expectedError error
	}{
		{
			name: "Not Found",
			setup: func() int32 {
				return -1
			},
			expectedError: pg_error.ErrNotFound,
		},
		{
			name: "Success",
			setup: func() int32 {
				sc, err := insertSubcategory(pgPool, ctx, category.ID, subcategoryName1)
				if err != nil {
					t.Fatalf("failed to insert subcategory: %v", err)
				}

				return sc.ID
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subcategoryId := tt.setup()

			err := repo.DeleteSubcategoryById(ctx, subcategoryId)

			if tt.expectedError == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tt.expectedError)
			}
		})
	}
}

func insertSubcategoryDependencies(t *testing.T, dbPool *pgxpool.Pool, ctx context.Context) (entity.CategoryType, entity.Category) {
	categoryType, err := insertCategoryType(dbPool, ctx, categoryTypeName)
	if err != nil {
		t.Fatalf("failed to insert category type: %v", err)
	}

	category, err := insertCategory(dbPool, ctx, categoryType.ID, categoryName)
	if err != nil {
		t.Fatalf("failed to insert category: %v", err)
	}

	return categoryType, category
}

func insertSubcategory(dbPool *pgxpool.Pool, ctx context.Context, categoryID int32, name string) (entity.Subcategory, error) {
	row := dbPool.QueryRow(ctx, "INSERT INTO subcategories (category_id, name) VALUES ($1, $2) RETURNING id, category_id, name", categoryID, name)
	var i entity.Subcategory
	err := row.Scan(&i.ID, &i.CategoryID, &i.Name)
	return i, err
}
