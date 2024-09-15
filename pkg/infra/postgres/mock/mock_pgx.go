// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/infra/postgres/pgx.go
//
// Generated by this command:
//
//	mockgen -source=pkg/infra/postgres/pgx.go -destination=pkg/infra/postgres/mock/mock_pgx.go
//

// Package mock_postgres is a generated GoMock package.
package mock_postgres

import (
	context "context"
	reflect "reflect"

	postgres "github.com/hexley21/fixup/pkg/infra/postgres"
	pgx "github.com/jackc/pgx/v5"
	pgconn "github.com/jackc/pgx/v5/pgconn"
	gomock "go.uber.org/mock/gomock"
)

// MockPGX is a mock of PGX interface.
type MockPGX struct {
	ctrl     *gomock.Controller
	recorder *MockPGXMockRecorder
}

// MockPGXMockRecorder is the mock recorder for MockPGX.
type MockPGXMockRecorder struct {
	mock *MockPGX
}

// NewMockPGX creates a new mock instance.
func NewMockPGX(ctrl *gomock.Controller) *MockPGX {
	mock := &MockPGX{ctrl: ctrl}
	mock.recorder = &MockPGXMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPGX) EXPECT() *MockPGXMockRecorder {
	return m.recorder
}

// Begin mocks base method.
func (m *MockPGX) Begin(ctx context.Context) (pgx.Tx, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Begin", ctx)
	ret0, _ := ret[0].(pgx.Tx)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Begin indicates an expected call of Begin.
func (mr *MockPGXMockRecorder) Begin(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Begin", reflect.TypeOf((*MockPGX)(nil).Begin), ctx)
}

// BeginTx mocks base method.
func (m *MockPGX) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BeginTx", ctx, txOptions)
	ret0, _ := ret[0].(pgx.Tx)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BeginTx indicates an expected call of BeginTx.
func (mr *MockPGXMockRecorder) BeginTx(ctx, txOptions any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BeginTx", reflect.TypeOf((*MockPGX)(nil).BeginTx), ctx, txOptions)
}

// Exec mocks base method.
func (m *MockPGX) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	m.ctrl.T.Helper()
	varargs := []any{ctx, sql}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Exec", varargs...)
	ret0, _ := ret[0].(pgconn.CommandTag)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Exec indicates an expected call of Exec.
func (mr *MockPGXMockRecorder) Exec(ctx, sql any, args ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, sql}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exec", reflect.TypeOf((*MockPGX)(nil).Exec), varargs...)
}

// Query mocks base method.
func (m *MockPGX) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	m.ctrl.T.Helper()
	varargs := []any{ctx, sql}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Query", varargs...)
	ret0, _ := ret[0].(pgx.Rows)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Query indicates an expected call of Query.
func (mr *MockPGXMockRecorder) Query(ctx, sql any, args ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, sql}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Query", reflect.TypeOf((*MockPGX)(nil).Query), varargs...)
}

// QueryRow mocks base method.
func (m *MockPGX) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	m.ctrl.T.Helper()
	varargs := []any{ctx, sql}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "QueryRow", varargs...)
	ret0, _ := ret[0].(pgx.Row)
	return ret0
}

// QueryRow indicates an expected call of QueryRow.
func (mr *MockPGXMockRecorder) QueryRow(ctx, sql any, args ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, sql}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryRow", reflect.TypeOf((*MockPGX)(nil).QueryRow), varargs...)
}

// MockPGXQuerier is a mock of PGXQuerier interface.
type MockPGXQuerier struct {
	ctrl     *gomock.Controller
	recorder *MockPGXQuerierMockRecorder
}

// MockPGXQuerierMockRecorder is the mock recorder for MockPGXQuerier.
type MockPGXQuerierMockRecorder struct {
	mock *MockPGXQuerier
}

// NewMockPGXQuerier creates a new mock instance.
func NewMockPGXQuerier(ctrl *gomock.Controller) *MockPGXQuerier {
	mock := &MockPGXQuerier{ctrl: ctrl}
	mock.recorder = &MockPGXQuerierMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPGXQuerier) EXPECT() *MockPGXQuerierMockRecorder {
	return m.recorder
}

// Begin mocks base method.
func (m *MockPGXQuerier) Begin(ctx context.Context) (pgx.Tx, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Begin", ctx)
	ret0, _ := ret[0].(pgx.Tx)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Begin indicates an expected call of Begin.
func (mr *MockPGXQuerierMockRecorder) Begin(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Begin", reflect.TypeOf((*MockPGXQuerier)(nil).Begin), ctx)
}

// Exec mocks base method.
func (m *MockPGXQuerier) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	m.ctrl.T.Helper()
	varargs := []any{ctx, sql}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Exec", varargs...)
	ret0, _ := ret[0].(pgconn.CommandTag)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Exec indicates an expected call of Exec.
func (mr *MockPGXQuerierMockRecorder) Exec(ctx, sql any, args ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, sql}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exec", reflect.TypeOf((*MockPGXQuerier)(nil).Exec), varargs...)
}

// Query mocks base method.
func (m *MockPGXQuerier) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	m.ctrl.T.Helper()
	varargs := []any{ctx, sql}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Query", varargs...)
	ret0, _ := ret[0].(pgx.Rows)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Query indicates an expected call of Query.
func (mr *MockPGXQuerierMockRecorder) Query(ctx, sql any, args ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, sql}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Query", reflect.TypeOf((*MockPGXQuerier)(nil).Query), varargs...)
}

// QueryRow mocks base method.
func (m *MockPGXQuerier) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	m.ctrl.T.Helper()
	varargs := []any{ctx, sql}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "QueryRow", varargs...)
	ret0, _ := ret[0].(pgx.Row)
	return ret0
}

// QueryRow indicates an expected call of QueryRow.
func (mr *MockPGXQuerierMockRecorder) QueryRow(ctx, sql any, args ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, sql}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryRow", reflect.TypeOf((*MockPGXQuerier)(nil).QueryRow), varargs...)
}

// MockRepository is a mock of Repository interface.
type MockRepository[R any] struct {
	ctrl     *gomock.Controller
	recorder *MockRepositoryMockRecorder[R]
}

// MockRepositoryMockRecorder is the mock recorder for MockRepository.
type MockRepositoryMockRecorder[R any] struct {
	mock *MockRepository[R]
}

// NewMockRepository creates a new mock instance.
func NewMockRepository[R any](ctrl *gomock.Controller) *MockRepository[R] {
	mock := &MockRepository[R]{ctrl: ctrl}
	mock.recorder = &MockRepositoryMockRecorder[R]{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRepository[R]) EXPECT() *MockRepositoryMockRecorder[R] {
	return m.recorder
}

// WithTx mocks base method.
func (m *MockRepository[R]) WithTx(q postgres.PGXQuerier) R {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithTx", q)
	ret0, _ := ret[0].(R)
	return ret0
}

// WithTx indicates an expected call of WithTx.
func (mr *MockRepositoryMockRecorder[R]) WithTx(q any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithTx", reflect.TypeOf((*MockRepository[R])(nil).WithTx), q)
}
