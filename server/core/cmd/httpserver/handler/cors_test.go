package handler_test

import (
	"net/http"
	"testing"

	"httpserver/handler"
)

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

func Test_cordHandler_ServeHTTP(t *testing.T) {
	t.Parallel()
	t.Run(
		`shall set Access-Control-Allow-Origin on "" input`, func(t *testing.T) {
			w := &MockWriter{Headers: http.Header{}}

			handler.CORSHandler(
				map[string]string{
					"Access-Control-Allow-Origin": "",
				},
				nil,
			).ServeHTTP(w, &http.Request{})

			if w.Header().Get("Access-Control-Allow-Origin") != "*" {
				t.Fatalf("Access-Control-Allow-Origin expected to be set to *")
			}
		},
	)

	t.Run(
		`shall set Access-Control-Allow-Origin on '*' input`, func(t *testing.T) {
			w := &MockWriter{Headers: http.Header{}}

			handler.CORSHandler(
				map[string]string{
					"Access-Control-Allow-Origin": "'*'",
				},
				nil,
			).ServeHTTP(w, &http.Request{})

			if w.Header().Get("Access-Control-Allow-Origin") != "*" {
				t.Error("Access-Control-Allow-Origin expected to be set to *")
			}
		},
	)

	t.Run(
		`shall only set the headers foo and bar to values qux and quxx respectively`, func(t *testing.T) {
			w := &MockWriter{Headers: http.Header{}}
			m := map[string]string{
				"foo": "qux",
				"bar": "quxx",
			}

			handler.CORSHandler(m, nil).ServeHTTP(w, &http.Request{})

			for k, want := range m {
				got := w.Header().Get(k)
				if got != want {
					t.Errorf("header %s want: %s, got: %s", k, want, got)
				}
			}

			if len(w.Header()) != len(m) {
				t.Errorf("number of headers does not match expectation")
			}
		},
	)

	t.Run(
		"shall shorten the handlers chain on OPTIONS request", func(t *testing.T) {
			w := &MockWriter{Headers: http.Header{}}
			m := map[string]string{
				"foo": "bar",
			}
			// Note: it must differ from 200
			const probeStatus = 201

			handler.CORSHandler(
				m,
				chainHandler{probeStatus},
			).ServeHTTP(w, &http.Request{Method: http.MethodOptions})

			if w.StatusCode == probeStatus {
				t.Error("200 is expected as the status code")
			}
		},
	)
}

type chainHandler struct {
	status int
}

func (c chainHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(c.status)
	return
}
