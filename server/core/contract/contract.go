package contract

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/kislerdm/diagramastext/server/core"
)

// UserProfile requesting user's profile.
type UserProfile struct {
	UserID                 string
	IsRegistered           bool
	OptOutFromSavingPrompt bool
}

type DiagramGraphPredictionRequest struct {
	Model  string
	Prompt string
}

// DiagramHandler handler the model inference request's attributes
// and communicates with the diagram rendering backend.
type DiagramHandler interface {
	// GetModelRequest returns the model inference request's attributes to generate the diagram graph.
	GetModelRequest(inquery Inquiry) DiagramGraphPredictionRequest

	// RenderPredictionResultsDiagramSVG renders the diagram using the prediction results.
	RenderPredictionResultsDiagramSVG(ctx context.Context, prediction []byte) ([]byte, error)
}

// Entrypoint functional entrypoint to render the text prompt to a diagram.
type Entrypoint func(
	ctx context.Context, inquery Inquiry, diagramImplementationHandler DiagramHandler,
) (
	[]byte, error,
)

// ClientModelInference the logic to infer the model and predict a diagram graph.
type ClientModelInference interface {
	Do(ctx context.Context, prompt, model string) ([]byte, error)
}

// MockClientModelInference mock of the model inference client.
type MockClientModelInference struct {
	Prediction []byte
	Err        error
}

func (m MockClientModelInference) Do(_ context.Context, _, _ string) ([]byte, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Prediction, nil
}

// ClientStorage the client to persist user's prompt and the model's predictions.
type ClientStorage interface {
	// WritePrompt writes user's input prompt.
	WritePrompt(ctx context.Context, requestID string, prompt string, userID string) error

	// WriteModelPrediction writes model's prediction result used to generate diagram.
	WriteModelPrediction(ctx context.Context, requestID string, result string, userID string) error

	// Close closes the connection.
	Close(ctx context.Context) error
}

// MockClientStorage mock of the postgres client.
type MockClientStorage struct {
	Err error
}

func (m MockClientStorage) WritePrompt(_ context.Context, _, _, _ string) error {
	return m.Err
}

func (m MockClientStorage) WriteModelPrediction(_ context.Context, _, _, _ string) error {
	return m.Err
}

func (m MockClientStorage) Close(_ context.Context) error {
	return m.Err
}

// Inquiry model inference inquiry object.
type Inquiry struct {
	// Request API request.
	*core.Request

	// Model the model ID to infer.
	*UserProfile
}

const (
	promptLengthMin = 3

	promptLengthMaxBase       = 100
	promptLengthMaxRegistered = 300
)

// ValidatePrompt validates the request.
func (r Inquiry) ValidatePrompt() error {
	prompt := strings.ReplaceAll(r.Prompt, "\n", "")
	if r.IsRegistered {
		return validatePromptRegisteredUser(prompt)
	}
	return validatePromptBaseUser(prompt)
}

func validatePromptBaseUser(prompt string) error {
	return validatePromptLength(prompt, promptLengthMaxBase)
}

func validatePromptRegisteredUser(prompt string) error {
	return validatePromptLength(prompt, promptLengthMaxRegistered)
}

func validatePromptLength(prompt string, max int) error {
	if len(prompt) < promptLengthMin || len(prompt) > max {
		return errors.New(
			"prompt length must be between " + strconv.Itoa(promptLengthMin) + " and " +
				strconv.Itoa(max) + " characters",
		)
	}
	return nil
}

// ClientHTTP http client.
type ClientHTTP interface {
	Do(req *http.Request) (resp ClientHTTPResp, err error)
}

type ClientHTTPResp struct {
	Body       []byte
	StatusCode int
}

// ClientSecretsmanager client to communicate to the secretsmanager.
type ClientSecretsmanager interface {
	// ReadLatestSecret reads and deserializes the latest version of JSON-encoded secret.
	ReadLatestSecret(ctx context.Context, uri string, output interface{}) error
}
