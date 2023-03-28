package diagram

import (
	"context"
	"encoding/json"
	"net/http"
)

// HTTPHandler handler to generate a diagram given the input.
type HTTPHandler func(ctx context.Context, input Input) (Output, error)

// RepositoryPrediction defines the interface to store prediction input (prompt) and model result.
type RepositoryPrediction interface {
	WriteInputPrompt(ctx context.Context, requestID, userID, prompt string) error
	WriteModelResult(ctx context.Context, requestID, userID, prediction string) error
	Close(ctx context.Context) error
}

type MockRepositoryPrediction struct {
	Err error
}

func (m MockRepositoryPrediction) WriteInputPrompt(_ context.Context, _, _, _ string) error {
	return m.Err
}

func (m MockRepositoryPrediction) WriteModelResult(_ context.Context, _, _, _ string) error {
	return m.Err
}

func (m MockRepositoryPrediction) Close(_ context.Context) error {
	return m.Err
}

// RepositorySecretsVault defines the interface to read secrets from the vault.
type RepositorySecretsVault interface {
	ReadLastVersion(ctx context.Context, uri string, output interface{}) error
}

type MockRepositorySecretsVault struct {
	V   []byte
	Err error
}

func (m MockRepositorySecretsVault) ReadLastVersion(_ context.Context, _ string, o interface{}) error {
	if m.Err != nil {
		return m.Err
	}
	return json.Unmarshal(m.V, o)
}

// ModelInference interface to communicate with the model.
type ModelInference interface {
	Do(ctx context.Context, userPrompt string, systemContent string, model string) ([]byte, error)
}

type MockModelInference struct {
	V   []byte
	Err error
}

func (m MockModelInference) Do(_ context.Context, _, _, _ string) ([]byte, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.V, nil
}

// HTTPClient client to communicate over http.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type MockHTTPClient struct {
	V   *http.Response
	Err error
}

func (m MockHTTPClient) Do(_ *http.Request) (*http.Response, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.V, nil
}
