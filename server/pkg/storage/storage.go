package storage

import (
	"context"
	"time"
)

// CallID client request's composed ID.
type CallID struct {
	RequestID string
	UserID    string
}

// UserInput type defining the user's input.
type UserInput struct {
	CallID
	Prompt    string
	Timestamp time.Time
}

// ModelOutput type defining the model's output.
type ModelOutput struct {
	CallID
	Response  string
	Timestamp time.Time
}

// Client defines the pgClient to communicate to the storage to persist core logic transactions.
type Client interface {
	// WritePrompt writes user's input prompt.
	WritePrompt(ctx context.Context, v UserInput) error

	// WriteModelPrediction writes model's prediction result used to generate diagram.
	WriteModelPrediction(ctx context.Context, v ModelOutput) error

	// Close closes the connection.
	Close(ctx context.Context) error
}

type MockClientStorage struct {
	Err error
}

func (m MockClientStorage) WritePrompt(ctx context.Context, v UserInput) error {
	return m.Err
}

func (m MockClientStorage) WriteModelPrediction(ctx context.Context, v ModelOutput) error {
	return m.Err
}

func (m MockClientStorage) Close(ctx context.Context) error {
	return m.Err
}
