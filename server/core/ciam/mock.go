package ciam

import "net/http"

type MockWriter struct {
	Headers    http.Header
	StatusCode int
	V          []byte
}

func (m *MockWriter) Header() http.Header {
	return m.Headers
}

func (m *MockWriter) Write(bytes []byte) (int, error) {
	m.V = bytes
	return len(bytes), nil
}

func (m *MockWriter) WriteHeader(statusCode int) {
	m.StatusCode = statusCode
}
