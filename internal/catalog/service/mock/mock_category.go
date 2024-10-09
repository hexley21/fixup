// Code generated by MockGen. DO NOT EDIT.
// Source: internal/catalog/service/category.go
//
// Generated by this command:
//
//	mockgen -source=internal/catalog/service/category.go -destination=internal/catalog/service/mock/mock_category.go
//

// Package mock_service is a generated GoMock package.
package mock_service

import (
	context "context"
	reflect "reflect"

	dto "github.com/hexley21/fixup/internal/catalog/delivery/http/v1/dto"
	gomock "go.uber.org/mock/gomock"
)

// MockCategoryService is a mock of CategoryService interface.
type MockCategoryService struct {
	ctrl     *gomock.Controller
	recorder *MockCategoryServiceMockRecorder
}

// MockCategoryServiceMockRecorder is the mock recorder for MockCategoryService.
type MockCategoryServiceMockRecorder struct {
	mock *MockCategoryService
}

// NewMockCategoryService creates a new mock instance.
func NewMockCategoryService(ctrl *gomock.Controller) *MockCategoryService {
	mock := &MockCategoryService{ctrl: ctrl}
	mock.recorder = &MockCategoryServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCategoryService) EXPECT() *MockCategoryServiceMockRecorder {
	return m.recorder
}

// CreateCategory mocks base method.
func (m *MockCategoryService) CreateCategory(ctx context.Context, createDTO dto.CreateCategoryDTO) (dto.CategoryDTO, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateCategory", ctx, createDTO)
	ret0, _ := ret[0].(dto.CategoryDTO)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateCategory indicates an expected call of CreateCategory.
func (mr *MockCategoryServiceMockRecorder) CreateCategory(ctx, dto any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateCategory", reflect.TypeOf((*MockCategoryService)(nil).CreateCategory), ctx, dto)
}

// DeleteCategoryById mocks base method.
func (m *MockCategoryService) DeleteCategoryById(ctx context.Context, id int32) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteCategoryById", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteCategoryById indicates an expected call of DeleteCategoryById.
func (mr *MockCategoryServiceMockRecorder) DeleteCategoryById(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteCategoryById", reflect.TypeOf((*MockCategoryService)(nil).DeleteCategoryById), ctx, id)
}

// GetCategories mocks base method.
func (m *MockCategoryService) GetCategories(ctx context.Context, page, per_page int32) ([]dto.CategoryDTO, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCategories", ctx, page, per_page)
	ret0, _ := ret[0].([]dto.CategoryDTO)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCategories indicates an expected call of GetCategories.
func (mr *MockCategoryServiceMockRecorder) GetCategories(ctx, page, per_page any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCategories", reflect.TypeOf((*MockCategoryService)(nil).GetCategories), ctx, page, per_page)
}

// GetCategoriesByTypeId mocks base method.
func (m *MockCategoryService) GetCategoriesByTypeId(ctx context.Context, id, page, per_page int32) ([]dto.CategoryDTO, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCategoriesByTypeId", ctx, id, page, per_page)
	ret0, _ := ret[0].([]dto.CategoryDTO)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCategoriesByTypeId indicates an expected call of GetCategoriesByTypeId.
func (mr *MockCategoryServiceMockRecorder) GetCategoriesByTypeId(ctx, id, page, per_page any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCategoriesByTypeId", reflect.TypeOf((*MockCategoryService)(nil).GetCategoriesByTypeId), ctx, id, page, per_page)
}

// GetCategoryById mocks base method.
func (m *MockCategoryService) GetCategoryById(ctx context.Context, id int32) (dto.CategoryDTO, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCategoryById", ctx, id)
	ret0, _ := ret[0].(dto.CategoryDTO)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCategoryById indicates an expected call of GetCategoryById.
func (mr *MockCategoryServiceMockRecorder) GetCategoryById(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCategoryById", reflect.TypeOf((*MockCategoryService)(nil).GetCategoryById), ctx, id)
}

// UpdateCategoryById mocks base method.
func (m *MockCategoryService) UpdateCategoryById(ctx context.Context, id int32, patchDTO dto.PatchCategoryDTO) (dto.CategoryDTO, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateCategoryById", ctx, id, patchDTO)
	ret0, _ := ret[0].(dto.CategoryDTO)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateCategoryById indicates an expected call of UpdateCategoryById.
func (mr *MockCategoryServiceMockRecorder) UpdateCategoryById(ctx, id, dto any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateCategoryById", reflect.TypeOf((*MockCategoryService)(nil).UpdateCategoryById), ctx, id, dto)
}
