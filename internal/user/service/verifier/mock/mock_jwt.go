// Code generated by MockGen. DO NOT EDIT.
// Source: internal/user/service/verifier/jwt.go
//
// Generated by this command:
//
//	mockgen -source=internal/user/service/verifier/jwt.go -destination=internal/user/service/verifier/mock/mock_jwt.go
//

// Package mock_verifier is a generated GoMock package.
package mock_verifier

import (
	reflect "reflect"

	verifier "github.com/hexley21/fixup/internal/user/service/verifier"
	gomock "go.uber.org/mock/gomock"
)

// MockJwt is a mock of Jwt interface.
type MockJwt struct {
	ctrl     *gomock.Controller
	recorder *MockJwtMockRecorder
}

// MockJwtMockRecorder is the mock recorder for MockJwt.
type MockJwtMockRecorder struct {
	mock *MockJwt
}

// NewMockJwt creates a new mock instance.
func NewMockJwt(ctrl *gomock.Controller) *MockJwt {
	mock := &MockJwt{ctrl: ctrl}
	mock.recorder = &MockJwtMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockJwt) EXPECT() *MockJwtMockRecorder {
	return m.recorder
}

// GenerateToken mocks base method.
func (m *MockJwt) GenerateToken(id, email string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenerateToken", id, email)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GenerateToken indicates an expected call of GenerateToken.
func (mr *MockJwtMockRecorder) GenerateToken(id, email any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenerateToken", reflect.TypeOf((*MockJwt)(nil).GenerateToken), id, email)
}

// VerifyJWT mocks base method.
func (m *MockJwt) VerifyJWT(tokenString string) (verifier.VerifyClaims, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "VerifyJWT", tokenString)
	ret0, _ := ret[0].(verifier.VerifyClaims)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// VerifyJWT indicates an expected call of VerifyJWT.
func (mr *MockJwtMockRecorder) VerifyJWT(tokenString any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "VerifyJWT", reflect.TypeOf((*MockJwt)(nil).VerifyJWT), tokenString)
}

// MockJwtGenerator is a mock of JwtGenerator interface.
type MockJwtGenerator struct {
	ctrl     *gomock.Controller
	recorder *MockJwtGeneratorMockRecorder
}

// MockJwtGeneratorMockRecorder is the mock recorder for MockJwtGenerator.
type MockJwtGeneratorMockRecorder struct {
	mock *MockJwtGenerator
}

// NewMockJwtGenerator creates a new mock instance.
func NewMockJwtGenerator(ctrl *gomock.Controller) *MockJwtGenerator {
	mock := &MockJwtGenerator{ctrl: ctrl}
	mock.recorder = &MockJwtGeneratorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockJwtGenerator) EXPECT() *MockJwtGeneratorMockRecorder {
	return m.recorder
}

// GenerateToken mocks base method.
func (m *MockJwtGenerator) GenerateToken(id, email string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenerateToken", id, email)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GenerateToken indicates an expected call of GenerateToken.
func (mr *MockJwtGeneratorMockRecorder) GenerateToken(id, email any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenerateToken", reflect.TypeOf((*MockJwtGenerator)(nil).GenerateToken), id, email)
}

// MockJwtVerifier is a mock of JwtVerifier interface.
type MockJwtVerifier struct {
	ctrl     *gomock.Controller
	recorder *MockJwtVerifierMockRecorder
}

// MockJwtVerifierMockRecorder is the mock recorder for MockJwtVerifier.
type MockJwtVerifierMockRecorder struct {
	mock *MockJwtVerifier
}

// NewMockJwtVerifier creates a new mock instance.
func NewMockJwtVerifier(ctrl *gomock.Controller) *MockJwtVerifier {
	mock := &MockJwtVerifier{ctrl: ctrl}
	mock.recorder = &MockJwtVerifierMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockJwtVerifier) EXPECT() *MockJwtVerifierMockRecorder {
	return m.recorder
}

// VerifyJWT mocks base method.
func (m *MockJwtVerifier) VerifyJWT(tokenString string) (verifier.VerifyClaims, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "VerifyJWT", tokenString)
	ret0, _ := ret[0].(verifier.VerifyClaims)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// VerifyJWT indicates an expected call of VerifyJWT.
func (mr *MockJwtVerifierMockRecorder) VerifyJWT(tokenString any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "VerifyJWT", reflect.TypeOf((*MockJwtVerifier)(nil).VerifyJWT), tokenString)
}