// Code generated by MockGen. DO NOT EDIT.
// Source: ./chains/evm/listener/handlers/deposit.go
//
// Generated by this command:
//
//	mockgen -source=./chains/evm/listener/handlers/deposit.go -destination=./mock/deposit.go -package mock
//
// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	big "math/big"
	reflect "reflect"

	common "github.com/ethereum/go-ethereum/common"
	types "github.com/ethereum/go-ethereum/core/types"
	gomock "go.uber.org/mock/gomock"
)

// MockClient is a mock of Client interface.
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *MockClientMockRecorder
}

// MockClientMockRecorder is the mock recorder for MockClient.
type MockClientMockRecorder struct {
	mock *MockClient
}

// NewMockClient creates a new mock instance.
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &MockClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClient) EXPECT() *MockClientMockRecorder {
	return m.recorder
}

// BlockByHash mocks base method.
func (m *MockClient) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BlockByHash", ctx, hash)
	ret0, _ := ret[0].(*types.Block)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BlockByHash indicates an expected call of BlockByHash.
func (mr *MockClientMockRecorder) BlockByHash(ctx, hash any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BlockByHash", reflect.TypeOf((*MockClient)(nil).BlockByHash), ctx, hash)
}

// CallContext mocks base method.
func (m *MockClient) CallContext(ctx context.Context, target any, rpcMethod string, args ...any) error {
	m.ctrl.T.Helper()
	varargs := []any{ctx, target, rpcMethod}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "CallContext", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// CallContext indicates an expected call of CallContext.
func (mr *MockClientMockRecorder) CallContext(ctx, target, rpcMethod any, args ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, target, rpcMethod}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CallContext", reflect.TypeOf((*MockClient)(nil).CallContext), varargs...)
}

// FetchEventLogs mocks base method.
func (m *MockClient) FetchEventLogs(ctx context.Context, contractAddress common.Address, event string, startBlock, endBlock *big.Int) ([]types.Log, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchEventLogs", ctx, contractAddress, event, startBlock, endBlock)
	ret0, _ := ret[0].([]types.Log)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FetchEventLogs indicates an expected call of FetchEventLogs.
func (mr *MockClientMockRecorder) FetchEventLogs(ctx, contractAddress, event, startBlock, endBlock any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchEventLogs", reflect.TypeOf((*MockClient)(nil).FetchEventLogs), ctx, contractAddress, event, startBlock, endBlock)
}

// TransactionReceipt mocks base method.
func (m *MockClient) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TransactionReceipt", ctx, txHash)
	ret0, _ := ret[0].(*types.Receipt)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// TransactionReceipt indicates an expected call of TransactionReceipt.
func (mr *MockClientMockRecorder) TransactionReceipt(ctx, txHash any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TransactionReceipt", reflect.TypeOf((*MockClient)(nil).TransactionReceipt), ctx, txHash)
}