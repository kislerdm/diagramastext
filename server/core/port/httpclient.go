package port

import (
	"io"
	"net/http"
)

// HTTPResponse HTTPClient response object.
type HTTPResponse struct {
	Body       io.ReadCloser
	StatusCode int
}

// HTTPClient client to communicate over http.
type HTTPClient interface {
	Do(req *http.Request) (*HTTPResponse, error)
}

type MockHTTPClient struct {
	V   *HTTPResponse
	Err error
}

func (m MockHTTPClient) Do(_ *http.Request) (*HTTPResponse, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.V, nil
}
