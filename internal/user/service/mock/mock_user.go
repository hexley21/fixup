// Code generated by MockGen. DO NOT EDIT.
// Source: internal/user/service/user.go
//
// Generated by this command:
//
//	mockgen -source=internal/user/service/user.go -destination=internal/user/service/mock/mock_user.go
//

// Package mock_service is a generated GoMock package.
package mock_service

import (
	context "context"
	io "io"
	reflect "reflect"

	domain "github.com/hexley21/fixup/internal/user/domain"
	gomock "go.uber.org/mock/gomock"
)

// MockUserService is a mock of UserService interface.
type MockUserService struct {
	ctrl     *gomock.Controller
	recorder *MockUserServiceMockRecorder
}

// MockUserServiceMockRecorder is the mock recorder for MockUserService.
type MockUserServiceMockRecorder struct {
	mock *MockUserService
}

// NewMockUserService creates a new mock instance.
func NewMockUserService(ctrl *gomock.Controller) *MockUserService {
	mock := &MockUserService{ctrl: ctrl}
	mock.recorder = &MockUserServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUserService) EXPECT() *MockUserServiceMockRecorder {
	return m.recorder
}

// Delete mocks base method.
func (m *MockUserService) Delete(ctx context.Context, userId int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", ctx, userId)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockUserServiceMockRecorder) Delete(ctx, userId any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockUserService)(nil).Delete), ctx, userId)
}

// Get mocks base method.
func (m *MockUserService) Get(ctx context.Context, userId int64) (*domain.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, userId)
	ret0, _ := ret[0].(*domain.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockUserServiceMockRecorder) Get(ctx, userId any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockUserService)(nil).Get), ctx, userId)
}

// UpdatePassword mocks base method.
func (m *MockUserService) UpdatePassword(ctx context.Context, id int64, oldPassowrd, newPassword string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdatePassword", ctx, id, oldPassowrd, newPassword)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdatePassword indicates an expected call of UpdatePassword.
func (mr *MockUserServiceMockRecorder) UpdatePassword(ctx, id, oldPassowrd, newPassword any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdatePassword", reflect.TypeOf((*MockUserService)(nil).UpdatePassword), ctx, id, oldPassowrd, newPassword)
}

// UpdatePersonalInfo mocks base method.
func (m *MockUserService) UpdatePersonalInfo(ctx context.Context, id int64, personalInfo *domain.UserPersonalInfo) (*domain.UserPersonalInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdatePersonalInfo", ctx, id, personalInfo)
	ret0, _ := ret[0].(*domain.UserPersonalInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdatePersonalInfo indicates an expected call of UpdatePersonalInfo.
func (mr *MockUserServiceMockRecorder) UpdatePersonalInfo(ctx, id, personalInfo any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdatePersonalInfo", reflect.TypeOf((*MockUserService)(nil).UpdatePersonalInfo), ctx, id, personalInfo)
}

// UpdateProfilePicture mocks base method.
func (m *MockUserService) UpdateProfilePicture(ctx context.Context, userId int64, file io.Reader, fileName string, fileSize int64, fileType string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateProfilePicture", ctx, userId, file, fileName, fileSize, fileType)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateProfilePicture indicates an expected call of UpdateProfilePicture.
func (mr *MockUserServiceMockRecorder) UpdateProfilePicture(ctx, userId, file, fileName, fileSize, fileType any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateProfilePicture", reflect.TypeOf((*MockUserService)(nil).UpdateProfilePicture), ctx, userId, file, fileName, fileSize, fileType)
}
