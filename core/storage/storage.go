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

// Client defines the client to store the prompt and prediction.
type Client interface {
	WritePrompt(ctx context.Context, v UserInput) error
	WriteModelPrediction(ctx context.Context, v ModelOutput) error
}
