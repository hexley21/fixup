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
	subcategoryName1 = "Door fix"
	subcategoryName2 = "AC Montage"
	subcategoryId    = int32(1)
)

func setupSubategory() (
	ctx context.Context,
	pgPool *pgxpool.Pool,
	repo repository.Subcategory,
) {
	ctx = context.Background()

	pgPool = getPgPool(ctx)
	repo = repository.NewSubcategoryRepository(pgPool)

	return
}

func TestCreateSubcategory(t *testing.T) {
	ctx, pgPool, repo := setupSubategory()
	defer cleanupPostgres(ctx, pgPool)

	category := &repository.CategoryModel{}

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
				_, _ = repo.Create(ctx, domain.SubcategoryInfo{Name: subcategoryName1, CategoryID: category.ID})
			},
			expectedCode: pgerrcode.RaiseException,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			id, err := repo.Create(ctx, domain.SubcategoryInfo{Name: tt.subcategoryName1, CategoryID: category.ID})

			if tt.expectedCode == "" {
				assert.NoError(t, err)
				assert.NotEmpty(t, id)
			} else {
				var pgErr *pgconn.PgError
				if assert.ErrorAs(t, err, &pgErr) {
					assert.Equal(t, tt.expectedCode, pgErr.Code)
				}
				assert.Empty(t, id)
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
		setup         func() repository.SubcategoryModel
		expectedError error
	}{
		{
			name: "Not Found",
			setup: func() repository.SubcategoryModel {
				return repository.SubcategoryModel{ID: -1}
			},
			expectedError: pgx.ErrNoRows,
		},
		{
			name: "Success",
			setup: func() repository.SubcategoryModel {
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

			subcategory, err := repo.Get(ctx, subcategoryInsert.ID)

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
				_, err := insertSubcategory(pgPool, ctx, category.ID, subcategoryName2)
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

			subcategories, err := repo.List(ctx, 0, 5)

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

			subcategories, err := repo.ListByCategoryId(ctx, tt.categoryId, 0, 5)

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

			subcategories, err := repo.ListByTypeId(ctx, tt.typeId, 0, 5)

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
		setup             func() repository.SubcategoryModel
		subcategoryName   string
		expectedError     error
		expectedErrorCode string
	}{
		{
			name: "Not Found",
			setup: func() repository.SubcategoryModel {
				return repository.SubcategoryModel{ID: -1}
			},
			expectedError: pgx.ErrNoRows,
		},
		{
			name: "Success",
			setup: func() repository.SubcategoryModel {
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
			setup: func() repository.SubcategoryModel {
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

			subcategory, err := repo.Update(ctx, sc.ID, domain.SubcategoryInfo{Name: tt.subcategoryName, CategoryID: sc.CategoryID})

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
		name       string
		setup      func() int32
		expectedOk bool
	}{
		{
			name: "Not Found",
			setup: func() int32 {
				return -1
			},
			expectedOk: false,
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
			expectedOk: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subcategoryId := tt.setup()

			ok, err := repo.Delete(ctx, subcategoryId)
			assert.Equal(t, tt.expectedOk, ok)
			assert.NoError(t, err)
		})
	}
}

func insertSubcategoryDependencies(t *testing.T, dbPool *pgxpool.Pool, ctx context.Context) (repository.CategoryTypeModel, repository.CategoryModel) {
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

func insertSubcategory(dbPool *pgxpool.Pool, ctx context.Context, categoryID int32, name string) (repository.SubcategoryModel, error) {
	row := dbPool.QueryRow(ctx, "INSERT INTO subcategories (category_id, name) VALUES ($1, $2) RETURNING id, category_id, name", categoryID, name)
	var i repository.SubcategoryModel
	err := row.Scan(&i.ID, &i.CategoryID, &i.Name)
	return i, err
}
