package core

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/aws/smithy-go/rand"
)

// UserProfile requesting user's profile.
type UserProfile struct {
	UserID                 string
	IsRegistered           bool
	OptOutFromSavingPrompt bool
}

// Handler functional entrypoint to render the text prompt to a diagram.
type Handler func(
	ctx context.Context, inquery Inquiry, clientInference ClientModelInference, clientStorage ClientStorage,
) (
	[]byte, error,
)

// MockClientModelInference mock of the model inference client.
type MockClientModelInference struct {
	Prediction []byte
	Err        error
}

func (m MockClientModelInference) Do(_ context.Context, prompt, model string) ([]byte, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Prediction, nil
}

// ClientModelInference the logic to infer the model and predict a diagram graph.
type ClientModelInference interface {
	Do(ctx context.Context, prompt, model string) ([]byte, error)
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

// MockClientStorage mock of the storage client.
type MockClientStorage struct {
	Err error
}

func (m MockClientStorage) WritePrompt(_ context.Context, _ string, _ string, _ string) error {
	return m.Err
}

func (m MockClientStorage) WriteModelPrediction(_ context.Context, _ string, _ string, _ string) error {
	return m.Err
}

func (m MockClientStorage) Close(_ context.Context) error {
	return m.Err
}

// Inquiry model inference inquiry object.
type Inquiry struct {
	// Request API request.
	*Request

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

func generateRequestID() string {
	o, _ := rand.NewUUID(rand.Reader).GetUUID()
	return o
}

// StorePrediction persists the user's prompt and prediction results.
func StorePrediction(
	ctx context.Context, clientStorage ClientStorage, userID string, prompt string, graphPrediction string,
) {
	requestID := generateRequestID()
	// FIXME: combine both transactions into a single atomic transaction.
	if err := clientStorage.WritePrompt(ctx, requestID, prompt, userID); err == nil {
		_ = clientStorage.WriteModelPrediction(ctx, requestID, graphPrediction, userID)
	}
}
