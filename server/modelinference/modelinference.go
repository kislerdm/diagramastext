package modelinference

import (
	"context"
	"net/http"
)

// HttpClient http base client.
type HttpClient interface {
	Do(req *http.Request) (resp *http.Response, err error)
}

// Request model inference request.
type Request struct {
	BestOf uint8
	Prompt string
	Model  string
}

// Client interface defining the client to infer a model to convert user's prompt to a serialised data structure.
type Client interface {
	Do(ctx context.Context, req Request) ([]byte, error)
}
