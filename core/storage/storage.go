package storage

import (
	"context"
	"time"
)

// UserInput type defining the user's input.
type UserInput struct {
	RequestID string
	UserID    string
	Prompt    string
	Timestamp time.Time
}

// ModelOutput type defining the model's output.
type ModelOutput struct {
	RequestID string
	UserID    string
	Response  string
	Timestamp time.Time
}

// Client defines the client to communicate to the storage to persist core logic transactions.
type Client interface {
	// WritePrompt writes user's input prompt.
	WritePrompt(ctx context.Context, v UserInput) error

	// WriteModelPrediction writes model's prediction result used to generate diagram.
	WriteModelPrediction(ctx context.Context, v ModelOutput) error

	// Close closes the connection.
	Close(ctx context.Context) error
}
