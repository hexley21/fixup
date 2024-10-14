package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hexley21/fixup/internal/catalog/entity"
	mock_repository "github.com/hexley21/fixup/internal/catalog/repository/mock"
	"github.com/hexley21/fixup/internal/catalog/service"
	"github.com/hexley21/fixup/pkg/infra/postgres/pg_error"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func setupSubcategory(t *testing.T) (
	ctrl *gomock.Controller,
	ctx context.Context,
	mockSubcategoryRepo *mock_repository.MockSubcategory,
	svc service.Subcategory,

) {
	ctrl = gomock.NewController(t)
	ctx = context.Background()

	mockSubcategoryRepo = mock_repository.NewMockSubcategory(ctrl)
	svc = service.NewSubcategoryService(mockSubcategoryRepo)

	return
}


func TestGetSubcategory(t *testing.T) {
	ctrl, ctx, mockSubcategoryRepo, svc := setupSubcategory(t)
	defer ctrl.Finish()

    tests := []struct {
        name          string
        id            int32
        mockReturn    entity.Subcategory
        mockError     error
        expectedError error
    }{
        {
            name:          "Success",
            id:            1,
            mockReturn:    entity.Subcategory{ID: 1, SubcategoryInfo: entity.SubcategoryInfo{Name: "Test Subcategory"}},
        },
        {
            name:          "NotFound",
            id:            2,
            mockReturn:    entity.Subcategory{},
            mockError:     pgx.ErrNoRows,
            expectedError: service.ErrSubcategoryNotFound,
        },
        {
            name:          "OtherError",
            id:            3,
            mockReturn:    entity.Subcategory{},
            mockError:     errors.New("some error"),
            expectedError: errors.New("failed to fetch subcategory: some error"),
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
			mockSubcategoryRepo.EXPECT().Get(ctx, tt.id).Return(tt.mockReturn, tt.mockError)

            result, err := svc.Get(ctx, tt.id)

            if tt.expectedError != nil {
                assert.Error(t, err)
                assert.EqualError(t, err, tt.expectedError.Error())
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.mockReturn, result)
            }
        })
    }
}

func TestListSubcategories(t *testing.T) {
    ctrl, ctx, mockSubcategoryRepo, svc := setupSubcategory(t)
    defer ctrl.Finish()

    tests := []struct {
        name          string
        limit         int32
        offset        int32
        mockReturn    []entity.Subcategory
        mockError     error
        expectedError error
    }{
        {
            name:       "Success",
            limit:      10,
            offset:     0,
            mockReturn: []entity.Subcategory{{ID: 1, SubcategoryInfo: entity.SubcategoryInfo{Name: "Test Subcategory"}}},
        },
        {
            name:          "EmptyList",
            limit:         10,
            offset:        0,
            mockReturn:    []entity.Subcategory{},
            expectedError: nil,
        },
        {
            name:          "OtherError",
            limit:         10,
            offset:        0,
            mockReturn:    nil,
            mockError:     errors.New("some error"),
            expectedError: errors.New("failed to fetch subcategories: some error"),
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockSubcategoryRepo.EXPECT().List(ctx, tt.limit, tt.offset).Return(tt.mockReturn, tt.mockError)

            result, err := svc.List(ctx, tt.limit, tt.offset)

            if tt.expectedError != nil {
                assert.Error(t, err)
                assert.EqualError(t, err, tt.expectedError.Error())
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.mockReturn, result)
            }
        })
    }
}

func TestListByCategoryId(t *testing.T) {
    ctrl, ctx, mockSubcategoryRepo, svc := setupSubcategory(t)
    defer ctrl.Finish()

    tests := []struct {
        name          string
        categoryID    int32
        limit         int32
        offset        int32
        mockReturn    []entity.Subcategory
        mockError     error
        expectedError error
    }{
        {
            name:       "Success",
            categoryID: 1,
            limit:      10,
            offset:     0,
            mockReturn: []entity.Subcategory{{ID: 1, SubcategoryInfo: entity.SubcategoryInfo{Name: "Test Subcategory"}}},
        },
        {
            name:       "EmptyList",
            categoryID: 1,
            limit:      10,
            offset:     0,
            mockReturn: []entity.Subcategory{},
        },
        {
            name:          "OtherError",
            categoryID:    1,
            limit:         10,
            offset:        0,
            mockReturn:    nil,
            mockError:     errors.New("some error"),
            expectedError: errors.New("failed to fetch subcategories: some error"),
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockSubcategoryRepo.EXPECT().ListByCategoryId(ctx, tt.categoryID, tt.limit, tt.offset).Return(tt.mockReturn, tt.mockError)

            result, err := svc.ListByCategoryId(ctx, tt.categoryID, tt.limit, tt.offset)

            if tt.expectedError != nil {
                assert.Error(t, err)
                assert.EqualError(t, err, tt.expectedError.Error())
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.mockReturn, result)
            }
        })
    }
}

func TestListByTypeId(t *testing.T) {
    ctrl, ctx, mockSubcategoryRepo, svc := setupSubcategory(t)
    defer ctrl.Finish()

    tests := []struct {
        name          string
        typeID        int32
        limit         int32
        offset        int32
        mockReturn    []entity.Subcategory
        mockError     error
        expectedError error
    }{
        {
            name:       "Success",
            typeID:     1,
            limit:      10,
            offset:     0,
            mockReturn: []entity.Subcategory{{ID: 1, SubcategoryInfo: entity.SubcategoryInfo{Name: "Test Subcategory"}}},
        },
        {
            name:       "EmptyList",
            typeID:     1,
            limit:      10,
            offset:     0,
            mockReturn: []entity.Subcategory{},
        },
        {
            name:          "OtherError",
            typeID:        1,
            limit:         10,
            offset:        0,
            mockReturn:    nil,
            mockError:     errors.New("some error"),
            expectedError: errors.New("failed to fetch subcategories: some error"),
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockSubcategoryRepo.EXPECT().ListByTypeId(ctx, tt.typeID, tt.limit, tt.offset).Return(tt.mockReturn, tt.mockError)

            result, err := svc.ListByTypeId(ctx, tt.typeID, tt.limit, tt.offset)

            if tt.expectedError != nil {
                assert.Error(t, err)
                assert.EqualError(t, err, tt.expectedError.Error())
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.mockReturn, result)
            }
        })
    }
}

func TestCreateSubcategory(t *testing.T) {
    ctrl, ctx, mockSubcategoryRepo, svc := setupSubcategory(t)
    defer ctrl.Finish()

    tests := []struct {
        name          string
        info          entity.SubcategoryInfo
        mockReturnID  int32
        mockError     error
        expectedID    int32
        expectedError error
    }{
        {
            name:         "Success",
            info:         entity.SubcategoryInfo{Name: "Test Subcategory"},
            mockReturnID: 1,
            expectedID:   1,
        },
        {
            name:          "NameTaken",
            info:          entity.SubcategoryInfo{Name: "Duplicate Subcategory"},
            mockError:     &pgconn.PgError{Code: pgerrcode.RaiseException},
            expectedError: service.ErrSubcateogryNameTaken,
        },
        {
            name:          "OtherError",
            info:          entity.SubcategoryInfo{Name: "Another Subcategory"},
            mockError:     errors.New("some error"),
            expectedError: errors.New("failed to create subcategory: some error"),
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockSubcategoryRepo.EXPECT().Create(ctx, tt.info).Return(tt.mockReturnID, tt.mockError)

            resultID, err := svc.Create(ctx, tt.info)

            if tt.expectedError != nil {
                assert.Error(t, err)
                assert.EqualError(t, err, tt.expectedError.Error())
                assert.Equal(t, int32(0), resultID)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expectedID, resultID)
            }
        })
    }
}

func TestUpdateSubcategory(t *testing.T) {
    ctrl, ctx, mockSubcategoryRepo, svc := setupSubcategory(t)
    defer ctrl.Finish()

    tests := []struct {
        name          string
        id            int32
        info          entity.SubcategoryInfo
        mockReturn    entity.Subcategory
        mockError     error
        expectedError error
    }{
        {
            name:       "Success",
            id:         1,
            info:       entity.SubcategoryInfo{Name: "Updated Subcategory"},
            mockReturn: entity.Subcategory{ID: 1, SubcategoryInfo: entity.SubcategoryInfo{Name: "Updated Subcategory"}},
        },
        {
            name:          "NotFound",
            id:            2,
            info:          entity.SubcategoryInfo{Name: "Non-existent Subcategory"},
            mockError:     pgx.ErrNoRows,
            expectedError: service.ErrSubcategoryNotFound,
        },
        {
            name:          "NameTaken",
            id:            3,
            info:          entity.SubcategoryInfo{Name: "Duplicate Subcategory"},
            mockError:     &pgconn.PgError{Code: pgerrcode.RaiseException},
            expectedError: service.ErrSubcateogryNameTaken,
        },
        {
            name:          "OtherError",
            id:            4,
            info:          entity.SubcategoryInfo{Name: "Another Subcategory"},
            mockError:     errors.New("some error"),
            expectedError: errors.New("failed to update subcategory: some error"),
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockSubcategoryRepo.EXPECT().Update(ctx, tt.id, tt.info).Return(tt.mockReturn, tt.mockError)

            result, err := svc.Update(ctx, tt.id, tt.info)

            if tt.expectedError != nil {
                assert.Error(t, err)
                assert.EqualError(t, err, tt.expectedError.Error())
                assert.Equal(t, entity.Subcategory{}, result)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.mockReturn, result)
            }
        })
    }
}

func TestDeleteSubcategory(t *testing.T) {
    ctrl, ctx, mockSubcategoryRepo, svc := setupSubcategory(t)
    defer ctrl.Finish()

    tests := []struct {
        name          string
        id            int32
        mockError     error
        expectedError error
    }{
        {
            name:          "Success",
            id:            1,
            mockError:     nil,
            expectedError: nil,
        },
        {
            name:          "NotFound",
            id:            2,
            mockError:     pg_error.ErrNotFound,
            expectedError: service.ErrSubcategoryNotFound,
        },
        {
            name:          "OtherError",
            id:            3,
            mockError:     errors.New("some error"),
            expectedError: errors.New("failed to delete subcategory: some error"),
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockSubcategoryRepo.EXPECT().Delete(ctx, tt.id).Return(tt.mockError)

            err := svc.Delete(ctx, tt.id)

            if tt.expectedError != nil {
                assert.Error(t, err)
                assert.EqualError(t, err, tt.expectedError.Error())
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
