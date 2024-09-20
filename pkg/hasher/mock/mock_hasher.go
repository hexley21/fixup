// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/hasher/hasher.go
//
// Generated by this command:
//
//	mockgen -source=pkg/hasher/hasher.go -destination=pkg/hasher/mock/mock_hasher.go
//

// Package mock_hasher is a generated GoMock package.
package mock_hasher

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockHasher is a mock of Hasher interface.
type MockHasher struct {
	ctrl     *gomock.Controller
	recorder *MockHasherMockRecorder
}

// MockHasherMockRecorder is the mock recorder for MockHasher.
type MockHasherMockRecorder struct {
	mock *MockHasher
}

// NewMockHasher creates a new mock instance.
func NewMockHasher(ctrl *gomock.Controller) *MockHasher {
	mock := &MockHasher{ctrl: ctrl}
	mock.recorder = &MockHasherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockHasher) EXPECT() *MockHasherMockRecorder {
	return m.recorder
}

// GetSalt mocks base method.
func (m *MockHasher) GetSalt() []byte {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSalt")
	ret0, _ := ret[0].([]byte)
	return ret0
}

// GetSalt indicates an expected call of GetSalt.
func (mr *MockHasherMockRecorder) GetSalt() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSalt", reflect.TypeOf((*MockHasher)(nil).GetSalt))
}

// HashPassword mocks base method.
func (m *MockHasher) HashPassword(password string) string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HashPassword", password)
	ret0, _ := ret[0].(string)
	return ret0
}

// HashPassword indicates an expected call of HashPassword.
func (mr *MockHasherMockRecorder) HashPassword(password any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HashPassword", reflect.TypeOf((*MockHasher)(nil).HashPassword), password)
}

// HashPasswordWithSalt mocks base method.
func (m *MockHasher) HashPasswordWithSalt(password, salt string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HashPasswordWithSalt", password, salt)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// HashPasswordWithSalt indicates an expected call of HashPasswordWithSalt.
func (mr *MockHasherMockRecorder) HashPasswordWithSalt(password, salt any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HashPasswordWithSalt", reflect.TypeOf((*MockHasher)(nil).HashPasswordWithSalt), password, salt)
}

// VerifyPassword mocks base method.
func (m *MockHasher) VerifyPassword(password, hash string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "VerifyPassword", password, hash)
	ret0, _ := ret[0].(error)
	return ret0
}

// VerifyPassword indicates an expected call of VerifyPassword.
func (mr *MockHasherMockRecorder) VerifyPassword(password, hash any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "VerifyPassword", reflect.TypeOf((*MockHasher)(nil).VerifyPassword), password, hash)
}
