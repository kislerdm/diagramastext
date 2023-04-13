package diagram

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// HTTPHandler handler to generate a diagram given the input.
type HTTPHandler func(ctx context.Context, input Input) (Output, error)

// RepositoryPrediction defines the interface to store prediction input (prompt) and model result.
type RepositoryPrediction interface {
	// WriteInputPrompt records user's input prompt.
	WriteInputPrompt(ctx context.Context, requestID, userID, prompt string) error

	// WriteModelResult records the model's prediction result and the associated costs in tokens.
	WriteModelResult(
		ctx context.Context, requestID, userID, predictionRaw, prediction, model string,
		usageTokensPrompt, usageTokensCompletions uint16,
	) error

	// WriteSuccessFlag records the instance of a successful diagram generation
	// based on the model's prediction result.
	WriteSuccessFlag(ctx context.Context, requestID, userID, token string) error

	// GetDailySuccessfulResultsTimestampsByUserID reads the timestamps of all user's successful requests
	// which led to successful diagrams generation over the last 24 hours / day.
	GetDailySuccessfulResultsTimestampsByUserID(ctx context.Context, userID string) ([]time.Time, error)

	// Close closes connection to persistence service.
	Close(ctx context.Context) error
}

type MockRepositoryPrediction struct {
	Timestamps []time.Time
	Err        error
}

func (m MockRepositoryPrediction) GetDailySuccessfulResultsTimestampsByUserID(_ context.Context, _ string) (
	[]time.Time, error,
) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Timestamps, nil
}

func (m MockRepositoryPrediction) WriteInputPrompt(_ context.Context, _, _, _ string) error {
	return m.Err
}
func (m MockRepositoryPrediction) WriteModelResult(_ context.Context, _, _, _, _, _ string, _, _ uint16) error {
	return m.Err
}

func (m MockRepositoryPrediction) WriteSuccessFlag(_ context.Context, _, _, _ string) error {
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
	Do(ctx context.Context, userPrompt string, systemContent string, model string) (
		predictionRaw string, prediction []byte, usageTokensPrompt uint16, usageTokensCompletions uint16, err error,
	)
}

type MockModelInference struct {
	V               []byte
	UsagePrompt     uint16
	UsageCompletion uint16
	Err             error
}

func (m MockModelInference) Do(_ context.Context, _, _, _ string) (string, []byte, uint16, uint16, error) {
	if m.Err != nil {
		return "", nil, 0, 0, m.Err
	}
	return string(m.V), m.V, m.UsagePrompt, m.UsageCompletion, nil
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

// RepositoryToken defines the communication port to persistence layer hosting API access tokens.
type RepositoryToken interface {
	// GetActiveUserIDByActiveTokenID reads userID from the repository given the tokenID.
	// It returns a non-empty value if and only if the token and user are active.
	GetActiveUserIDByActiveTokenID(ctx context.Context, id string) (string, error)
}

type MockRepositoryToken struct {
	V   string
	Err error
}

func (m MockRepositoryToken) GetActiveUserIDByActiveTokenID(_ context.Context, _ string) (string, error) {
	return m.V, m.Err
}
