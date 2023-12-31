// Code generated by MockGen. DO NOT EDIT.
// Source: ./checkpoint.go

// Package mock_entity is a generated GoMock package.
package mock_entity

import (
	reflect "reflect"

	entity "github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/entity"
	gomock "github.com/golang/mock/gomock"
)

// MockCheckpointService is a mock of CheckpointService interface.
type MockCheckpointService struct {
	ctrl     *gomock.Controller
	recorder *MockCheckpointServiceMockRecorder
}

// MockCheckpointServiceMockRecorder is the mock recorder for MockCheckpointService.
type MockCheckpointServiceMockRecorder struct {
	mock *MockCheckpointService
}

// NewMockCheckpointService creates a new mock instance.
func NewMockCheckpointService(ctrl *gomock.Controller) *MockCheckpointService {
	mock := &MockCheckpointService{ctrl: ctrl}
	mock.recorder = &MockCheckpointServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCheckpointService) EXPECT() *MockCheckpointServiceMockRecorder {
	return m.recorder
}

// Checkpoint mocks base method.
func (m *MockCheckpointService) Checkpoint(config *entity.CheckpointConfig) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Checkpoint", config)
	ret0, _ := ret[0].(error)
	return ret0
}

// Checkpoint indicates an expected call of Checkpoint.
func (mr *MockCheckpointServiceMockRecorder) Checkpoint(config interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Checkpoint", reflect.TypeOf((*MockCheckpointService)(nil).Checkpoint), config)
}
