// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/buildpacks/lifecycle (interfaces: Platform)

// Package testmock is a generated GoMock package.
package testmock

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"

	platform "github.com/buildpacks/lifecycle/platform"
)

// MockPlatform is a mock of Platform interface.
type MockPlatform struct {
	ctrl     *gomock.Controller
	recorder *MockPlatformMockRecorder
}

// MockPlatformMockRecorder is the mock recorder for MockPlatform.
type MockPlatformMockRecorder struct {
	mock *MockPlatform
}

// NewMockPlatform creates a new mock instance.
func NewMockPlatform(ctrl *gomock.Controller) *MockPlatform {
	mock := &MockPlatform{ctrl: ctrl}
	mock.recorder = &MockPlatformMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPlatform) EXPECT() *MockPlatformMockRecorder {
	return m.recorder
}

// API mocks base method.
func (m *MockPlatform) API() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "API")
	ret0, _ := ret[0].(string)
	return ret0
}

// API indicates an expected call of API.
func (mr *MockPlatformMockRecorder) API() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "API", reflect.TypeOf((*MockPlatform)(nil).API))
}

// NewAnalyzedMetadata mocks base method.
func (m *MockPlatform) NewAnalyzedMetadata(arg0 platform.AnalyzedMetadataConfig) platform.AnalyzedMetadata {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewAnalyzedMetadata", arg0)
	ret0, _ := ret[0].(platform.AnalyzedMetadata)
	return ret0
}

// NewAnalyzedMetadata indicates an expected call of NewAnalyzedMetadata.
func (mr *MockPlatformMockRecorder) NewAnalyzedMetadata(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewAnalyzedMetadata", reflect.TypeOf((*MockPlatform)(nil).NewAnalyzedMetadata), arg0)
}

// SupportsAssetPackages mocks base method.
func (m *MockPlatform) SupportsAssetPackages() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SupportsAssetPackages")
	ret0, _ := ret[0].(bool)
	return ret0
}

// SupportsAssetPackages indicates an expected call of SupportsAssetPackages.
func (mr *MockPlatformMockRecorder) SupportsAssetPackages() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SupportsAssetPackages", reflect.TypeOf((*MockPlatform)(nil).SupportsAssetPackages))
}

// SupportsMixinValidation mocks base method.
func (m *MockPlatform) SupportsMixinValidation() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SupportsMixinValidation")
	ret0, _ := ret[0].(bool)
	return ret0
}

// SupportsMixinValidation indicates an expected call of SupportsMixinValidation.
func (mr *MockPlatformMockRecorder) SupportsMixinValidation() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SupportsMixinValidation", reflect.TypeOf((*MockPlatform)(nil).SupportsMixinValidation))
}
