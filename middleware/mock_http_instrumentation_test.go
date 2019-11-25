// Code generated by MockGen. DO NOT EDIT.
// Source: middleware/http_instrumentation.go

// Package middleware is a generated GoMock package.
package middleware_test

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockMetricReporter is a mock of MetricReporter interface
type MockMetricReporter struct {
	ctrl     *gomock.Controller
	recorder *MockMetricReporterMockRecorder
}

// MockMetricReporterMockRecorder is the mock recorder for MockMetricReporter
type MockMetricReporterMockRecorder struct {
	mock *MockMetricReporter
}

// NewMockMetricReporter creates a new mock instance
func NewMockMetricReporter(ctrl *gomock.Controller) *MockMetricReporter {
	mock := &MockMetricReporter{ctrl: ctrl}
	mock.recorder = &MockMetricReporterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockMetricReporter) EXPECT() *MockMetricReporterMockRecorder {
	return m.recorder
}

// ReportLatency mocks base method
func (m *MockMetricReporter) ReportLatency(routeName string, statusCode int, duration float64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "ReportLatency", routeName, statusCode, duration)
}

// ReportLatency indicates an expected call of ReportLatency
func (mr *MockMetricReporterMockRecorder) ReportLatency(routeName, statusCode, duration interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReportLatency", reflect.TypeOf((*MockMetricReporter)(nil).ReportLatency), routeName, statusCode, duration)
}
