// Automatically generated by MockGen. DO NOT EDIT!
// Source: ca.go

package services

import (
	gomock "github.com/golang/mock/gomock"
	ca "github.com/spiffe/spire/pkg/server/ca"
)

// Mock of CA interface
type MockCA struct {
	ctrl     *gomock.Controller
	recorder *_MockCARecorder
}

// Recorder for MockCA (not exported)
type _MockCARecorder struct {
	mock *MockCA
}

func NewMockCA(ctrl *gomock.Controller) *MockCA {
	mock := &MockCA{ctrl: ctrl}
	mock.recorder = &_MockCARecorder{mock}
	return mock
}

func (_m *MockCA) EXPECT() *_MockCARecorder {
	return _m.recorder
}

func (_m *MockCA) SignCsr(request *ca.SignCsrRequest) (*ca.SignCsrResponse, error) {
	ret := _m.ctrl.Call(_m, "SignCsr", request)
	ret0, _ := ret[0].(*ca.SignCsrResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockCARecorder) SignCsr(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SignCsr", arg0)
}

func (_m *MockCA) GetSpiffeIDFromCSR(csr []byte) (string, error) {
	ret := _m.ctrl.Call(_m, "GetSpiffeIDFromCSR", csr)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockCARecorder) GetSpiffeIDFromCSR(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetSpiffeIDFromCSR", arg0)
}
