// Code generated by MockGen. DO NOT EDIT.
// Source: ./internal/common/auth_jwt/jwt.go
//
// Generated by this command:
//
//	mockgen -source=./internal/common/auth_jwt/jwt.go -destination=./internal/common/auth_jwt/mock/mock_jwt.go
//

// Package mock_auth_jwt is a generated GoMock package.
package mock_auth_jwt

import (
	reflect "reflect"

	auth_jwt "github.com/hexley21/fixup/internal/common/auth_jwt"
	rest "github.com/hexley21/fixup/pkg/http/rest"
	gomock "go.uber.org/mock/gomock"
)

// MockManager is a mock of Manager interface.
type MockManager struct {
	ctrl     *gomock.Controller
	recorder *MockManagerMockRecorder
}

// MockManagerMockRecorder is the mock recorder for MockManager.
type MockManagerMockRecorder struct {
	mock *MockManager
}

// NewMockManager creates a new mock instance.
func NewMockManager(ctrl *gomock.Controller) *MockManager {
	mock := &MockManager{ctrl: ctrl}
	mock.recorder = &MockManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockManager) EXPECT() *MockManagerMockRecorder {
	return m.recorder
}

// Generate mocks base method.
func (m *MockManager) Generate(id, role string, verified bool) (string, *rest.ErrorResponse) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Generate", id, role, verified)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(*rest.ErrorResponse)
	return ret0, ret1
}

// Generate indicates an expected call of Generate.
func (mr *MockManagerMockRecorder) Generate(id, role, verified any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Generate", reflect.TypeOf((*MockManager)(nil).Generate), id, role, verified)
}

// Verify mocks base method.
func (m *MockManager) Verify(tokenString string) (auth_jwt.UserClaims, *rest.ErrorResponse) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Verify", tokenString)
	ret0, _ := ret[0].(auth_jwt.UserClaims)
	ret1, _ := ret[1].(*rest.ErrorResponse)
	return ret0, ret1
}

// Verify indicates an expected call of Verify.
func (mr *MockManagerMockRecorder) Verify(tokenString any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Verify", reflect.TypeOf((*MockManager)(nil).Verify), tokenString)
}

// MockGenerator is a mock of Generator interface.
type MockGenerator struct {
	ctrl     *gomock.Controller
	recorder *MockGeneratorMockRecorder
}

// MockGeneratorMockRecorder is the mock recorder for MockGenerator.
type MockGeneratorMockRecorder struct {
	mock *MockGenerator
}

// NewMockGenerator creates a new mock instance.
func NewMockGenerator(ctrl *gomock.Controller) *MockGenerator {
	mock := &MockGenerator{ctrl: ctrl}
	mock.recorder = &MockGeneratorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockGenerator) EXPECT() *MockGeneratorMockRecorder {
	return m.recorder
}

// Generate mocks base method.
func (m *MockGenerator) Generate(id, role string, verified bool) (string, *rest.ErrorResponse) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Generate", id, role, verified)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(*rest.ErrorResponse)
	return ret0, ret1
}

// Generate indicates an expected call of Generate.
func (mr *MockGeneratorMockRecorder) Generate(id, role, verified any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Generate", reflect.TypeOf((*MockGenerator)(nil).Generate), id, role, verified)
}

// MockVerifier is a mock of Verifier interface.
type MockVerifier struct {
	ctrl     *gomock.Controller
	recorder *MockVerifierMockRecorder
}

// MockVerifierMockRecorder is the mock recorder for MockVerifier.
type MockVerifierMockRecorder struct {
	mock *MockVerifier
}

// NewMockVerifier creates a new mock instance.
func NewMockVerifier(ctrl *gomock.Controller) *MockVerifier {
	mock := &MockVerifier{ctrl: ctrl}
	mock.recorder = &MockVerifierMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockVerifier) EXPECT() *MockVerifierMockRecorder {
	return m.recorder
}

// Verify mocks base method.
func (m *MockVerifier) Verify(tokenString string) (auth_jwt.UserClaims, *rest.ErrorResponse) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Verify", tokenString)
	ret0, _ := ret[0].(auth_jwt.UserClaims)
	ret1, _ := ret[1].(*rest.ErrorResponse)
	return ret0, ret1
}

// Verify indicates an expected call of Verify.
func (mr *MockVerifierMockRecorder) Verify(tokenString any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Verify", reflect.TypeOf((*MockVerifier)(nil).Verify), tokenString)
}