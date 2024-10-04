// Code generated by MockGen. DO NOT EDIT.
// Source: internal/user/repository/provider.go
//
// Generated by this command:
//
//	mockgen -source=internal/user/repository/provider.go -destination=internal/user/repository/mock/mock_provider.go
//

// Package mock_repository is a generated GoMock package.
package mock_repository

import (
	context "context"
	reflect "reflect"

	entity "github.com/hexley21/fixup/internal/user/entity"
	repository "github.com/hexley21/fixup/internal/user/repository"
	postgres "github.com/hexley21/fixup/pkg/infra/postgres"
	gomock "go.uber.org/mock/gomock"
)

// MockProviderRepository is a mock of ProviderRepository interface.
type MockProviderRepository struct {
	ctrl     *gomock.Controller
	recorder *MockProviderRepositoryMockRecorder
}

// MockProviderRepositoryMockRecorder is the mock recorder for MockProviderRepository.
type MockProviderRepositoryMockRecorder struct {
	mock *MockProviderRepository
}

// NewMockProviderRepository creates a new mock instance.
func NewMockProviderRepository(ctrl *gomock.Controller) *MockProviderRepository {
	mock := &MockProviderRepository{ctrl: ctrl}
	mock.recorder = &MockProviderRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProviderRepository) EXPECT() *MockProviderRepositoryMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockProviderRepository) Create(ctx context.Context, arg repository.CreateProviderParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", ctx, arg)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create.
func (mr *MockProviderRepositoryMockRecorder) Create(ctx, arg any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockProviderRepository)(nil).Create), ctx, arg)
}

// GetByUserId mocks base method.
func (m *MockProviderRepository) GetByUserId(ctx context.Context, userID int64) (entity.Provider, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByUserId", ctx, userID)
	ret0, _ := ret[0].(entity.Provider)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByUserId indicates an expected call of GetByUserId.
func (mr *MockProviderRepositoryMockRecorder) GetByUserId(ctx, userID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByUserId", reflect.TypeOf((*MockProviderRepository)(nil).GetByUserId), ctx, userID)
}

// WithTx mocks base method.
func (m *MockProviderRepository) WithTx(q postgres.PGXQuerier) repository.ProviderRepository {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithTx", q)
	ret0, _ := ret[0].(repository.ProviderRepository)
	return ret0
}

// WithTx indicates an expected call of WithTx.
func (mr *MockProviderRepositoryMockRecorder) WithTx(q any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithTx", reflect.TypeOf((*MockProviderRepository)(nil).WithTx), q)
}
