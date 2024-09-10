// Code generated by MockGen. DO NOT EDIT.
// Source: ./chains/evm/proof/root.go
//
// Generated by this command:
//
//	mockgen -source=./chains/evm/proof/root.go -destination=./mock/root.go -package mock
//
// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	api "github.com/attestantio/go-eth2-client/api"
	v1 "github.com/attestantio/go-eth2-client/api/v1"
	spec "github.com/attestantio/go-eth2-client/spec"
	gomock "go.uber.org/mock/gomock"
)

// MockBeaconClient is a mock of BeaconClient interface.
type MockBeaconClient struct {
	ctrl     *gomock.Controller
	recorder *MockBeaconClientMockRecorder
}

// MockBeaconClientMockRecorder is the mock recorder for MockBeaconClient.
type MockBeaconClientMockRecorder struct {
	mock *MockBeaconClient
}

// NewMockBeaconClient creates a new mock instance.
func NewMockBeaconClient(ctrl *gomock.Controller) *MockBeaconClient {
	mock := &MockBeaconClient{ctrl: ctrl}
	mock.recorder = &MockBeaconClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBeaconClient) EXPECT() *MockBeaconClientMockRecorder {
	return m.recorder
}

// BeaconBlockHeader mocks base method.
func (m *MockBeaconClient) BeaconBlockHeader(ctx context.Context, opts *api.BeaconBlockHeaderOpts) (*api.Response[*v1.BeaconBlockHeader], error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BeaconBlockHeader", ctx, opts)
	ret0, _ := ret[0].(*api.Response[*v1.BeaconBlockHeader])
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BeaconBlockHeader indicates an expected call of BeaconBlockHeader.
func (mr *MockBeaconClientMockRecorder) BeaconBlockHeader(ctx, opts any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BeaconBlockHeader", reflect.TypeOf((*MockBeaconClient)(nil).BeaconBlockHeader), ctx, opts)
}

// BeaconState mocks base method.
func (m *MockBeaconClient) BeaconState(ctx context.Context, opts *api.BeaconStateOpts) (*api.Response[*spec.VersionedBeaconState], error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BeaconState", ctx, opts)
	ret0, _ := ret[0].(*api.Response[*spec.VersionedBeaconState])
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BeaconState indicates an expected call of BeaconState.
func (mr *MockBeaconClientMockRecorder) BeaconState(ctx, opts any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BeaconState", reflect.TypeOf((*MockBeaconClient)(nil).BeaconState), ctx, opts)
}

// SignedBeaconBlock mocks base method.
func (m *MockBeaconClient) SignedBeaconBlock(ctx context.Context, opts *api.SignedBeaconBlockOpts) (*api.Response[*spec.VersionedSignedBeaconBlock], error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SignedBeaconBlock", ctx, opts)
	ret0, _ := ret[0].(*api.Response[*spec.VersionedSignedBeaconBlock])
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SignedBeaconBlock indicates an expected call of SignedBeaconBlock.
func (mr *MockBeaconClientMockRecorder) SignedBeaconBlock(ctx, opts any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SignedBeaconBlock", reflect.TypeOf((*MockBeaconClient)(nil).SignedBeaconBlock), ctx, opts)
}