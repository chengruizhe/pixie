// Code generated by MockGen. DO NOT EDIT.
// Source: server.go

// Package mock_controller is a generated GoMock package.
package mock_controller

import (
	reflect "reflect"

	uuid "github.com/gofrs/uuid"
	gomock "github.com/golang/mock/gomock"
	datastore "px.dev/pixie/src/cloud/profile/datastore"
)

// MockDatastore is a mock of Datastore interface.
type MockDatastore struct {
	ctrl     *gomock.Controller
	recorder *MockDatastoreMockRecorder
}

// MockDatastoreMockRecorder is the mock recorder for MockDatastore.
type MockDatastoreMockRecorder struct {
	mock *MockDatastore
}

// NewMockDatastore creates a new mock instance.
func NewMockDatastore(ctrl *gomock.Controller) *MockDatastore {
	mock := &MockDatastore{ctrl: ctrl}
	mock.recorder = &MockDatastoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDatastore) EXPECT() *MockDatastoreMockRecorder {
	return m.recorder
}

// ApproveAllOrgUsers mocks base method.
func (m *MockDatastore) ApproveAllOrgUsers(arg0 uuid.UUID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ApproveAllOrgUsers", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// ApproveAllOrgUsers indicates an expected call of ApproveAllOrgUsers.
func (mr *MockDatastoreMockRecorder) ApproveAllOrgUsers(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ApproveAllOrgUsers", reflect.TypeOf((*MockDatastore)(nil).ApproveAllOrgUsers), arg0)
}

// CreateUser mocks base method.
func (m *MockDatastore) CreateUser(arg0 *datastore.UserInfo) (uuid.UUID, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateUser", arg0)
	ret0, _ := ret[0].(uuid.UUID)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateUser indicates an expected call of CreateUser.
func (mr *MockDatastoreMockRecorder) CreateUser(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateUser", reflect.TypeOf((*MockDatastore)(nil).CreateUser), arg0)
}

// CreateUserAndOrg mocks base method.
func (m *MockDatastore) CreateUserAndOrg(arg0 *datastore.OrgInfo, arg1 *datastore.UserInfo) (uuid.UUID, uuid.UUID, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateUserAndOrg", arg0, arg1)
	ret0, _ := ret[0].(uuid.UUID)
	ret1, _ := ret[1].(uuid.UUID)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// CreateUserAndOrg indicates an expected call of CreateUserAndOrg.
func (mr *MockDatastoreMockRecorder) CreateUserAndOrg(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateUserAndOrg", reflect.TypeOf((*MockDatastore)(nil).CreateUserAndOrg), arg0, arg1)
}

// DeleteOrgAndUsers mocks base method.
func (m *MockDatastore) DeleteOrgAndUsers(arg0 uuid.UUID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteOrgAndUsers", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteOrgAndUsers indicates an expected call of DeleteOrgAndUsers.
func (mr *MockDatastoreMockRecorder) DeleteOrgAndUsers(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteOrgAndUsers", reflect.TypeOf((*MockDatastore)(nil).DeleteOrgAndUsers), arg0)
}

// GetOrg mocks base method.
func (m *MockDatastore) GetOrg(arg0 uuid.UUID) (*datastore.OrgInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOrg", arg0)
	ret0, _ := ret[0].(*datastore.OrgInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOrg indicates an expected call of GetOrg.
func (mr *MockDatastoreMockRecorder) GetOrg(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOrg", reflect.TypeOf((*MockDatastore)(nil).GetOrg), arg0)
}

// GetOrgByDomain mocks base method.
func (m *MockDatastore) GetOrgByDomain(arg0 string) (*datastore.OrgInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOrgByDomain", arg0)
	ret0, _ := ret[0].(*datastore.OrgInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOrgByDomain indicates an expected call of GetOrgByDomain.
func (mr *MockDatastoreMockRecorder) GetOrgByDomain(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOrgByDomain", reflect.TypeOf((*MockDatastore)(nil).GetOrgByDomain), arg0)
}

// GetOrgs mocks base method.
func (m *MockDatastore) GetOrgs() ([]*datastore.OrgInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOrgs")
	ret0, _ := ret[0].([]*datastore.OrgInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOrgs indicates an expected call of GetOrgs.
func (mr *MockDatastoreMockRecorder) GetOrgs() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOrgs", reflect.TypeOf((*MockDatastore)(nil).GetOrgs))
}

// GetUser mocks base method.
func (m *MockDatastore) GetUser(arg0 uuid.UUID) (*datastore.UserInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUser", arg0)
	ret0, _ := ret[0].(*datastore.UserInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUser indicates an expected call of GetUser.
func (mr *MockDatastoreMockRecorder) GetUser(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUser", reflect.TypeOf((*MockDatastore)(nil).GetUser), arg0)
}

// GetUserByEmail mocks base method.
func (m *MockDatastore) GetUserByEmail(arg0 string) (*datastore.UserInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserByEmail", arg0)
	ret0, _ := ret[0].(*datastore.UserInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserByEmail indicates an expected call of GetUserByEmail.
func (mr *MockDatastoreMockRecorder) GetUserByEmail(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserByEmail", reflect.TypeOf((*MockDatastore)(nil).GetUserByEmail), arg0)
}

// GetUsersInOrg mocks base method.
func (m *MockDatastore) GetUsersInOrg(arg0 uuid.UUID) ([]*datastore.UserInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUsersInOrg", arg0)
	ret0, _ := ret[0].([]*datastore.UserInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUsersInOrg indicates an expected call of GetUsersInOrg.
func (mr *MockDatastoreMockRecorder) GetUsersInOrg(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUsersInOrg", reflect.TypeOf((*MockDatastore)(nil).GetUsersInOrg), arg0)
}

// UpdateOrg mocks base method.
func (m *MockDatastore) UpdateOrg(arg0 *datastore.OrgInfo) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateOrg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateOrg indicates an expected call of UpdateOrg.
func (mr *MockDatastoreMockRecorder) UpdateOrg(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateOrg", reflect.TypeOf((*MockDatastore)(nil).UpdateOrg), arg0)
}

// UpdateUser mocks base method.
func (m *MockDatastore) UpdateUser(arg0 *datastore.UserInfo) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateUser", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateUser indicates an expected call of UpdateUser.
func (mr *MockDatastoreMockRecorder) UpdateUser(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateUser", reflect.TypeOf((*MockDatastore)(nil).UpdateUser), arg0)
}

// MockUserSettingsDatastore is a mock of UserSettingsDatastore interface.
type MockUserSettingsDatastore struct {
	ctrl     *gomock.Controller
	recorder *MockUserSettingsDatastoreMockRecorder
}

// MockUserSettingsDatastoreMockRecorder is the mock recorder for MockUserSettingsDatastore.
type MockUserSettingsDatastoreMockRecorder struct {
	mock *MockUserSettingsDatastore
}

// NewMockUserSettingsDatastore creates a new mock instance.
func NewMockUserSettingsDatastore(ctrl *gomock.Controller) *MockUserSettingsDatastore {
	mock := &MockUserSettingsDatastore{ctrl: ctrl}
	mock.recorder = &MockUserSettingsDatastoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUserSettingsDatastore) EXPECT() *MockUserSettingsDatastoreMockRecorder {
	return m.recorder
}

// GetUserSettings mocks base method.
func (m *MockUserSettingsDatastore) GetUserSettings(arg0 uuid.UUID, arg1 []string) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserSettings", arg0, arg1)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserSettings indicates an expected call of GetUserSettings.
func (mr *MockUserSettingsDatastoreMockRecorder) GetUserSettings(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserSettings", reflect.TypeOf((*MockUserSettingsDatastore)(nil).GetUserSettings), arg0, arg1)
}

// UpdateUserSettings mocks base method.
func (m *MockUserSettingsDatastore) UpdateUserSettings(arg0 uuid.UUID, arg1, arg2 []string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateUserSettings", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateUserSettings indicates an expected call of UpdateUserSettings.
func (mr *MockUserSettingsDatastoreMockRecorder) UpdateUserSettings(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateUserSettings", reflect.TypeOf((*MockUserSettingsDatastore)(nil).UpdateUserSettings), arg0, arg1, arg2)
}
