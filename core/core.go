package core

import (
	"context"
	"net/http"
	"time"
)

// DiagramGraph defines the diagram graph.
type DiagramGraph struct {
	Title  string  `json:"title,omitempty"`
	Footer string  `json:"footer,omitempty"`
	Nodes  []*Node `json:"nodes"`
	Links  []*Link `json:"links,omitempty"`
}

// Node diagram's definition node.
type Node struct {
	ID         string `json:"id"`
	Label      string `json:"label,omitempty"`
	Group      string `json:"group,omitempty"`
	Technology string `json:"technology,omitempty"`
	External   bool   `json:"external,omitempty"`
	IsQueue    bool   `json:"is_queue,omitempty"`
	IsDatabase bool   `json:"is_database,omitempty"`
}

// Link diagram's definition link.
type Link struct {
	From       string `json:"from"`
	To         string `json:"to"`
	Direction  string `json:"direction,omitempty"`
	Label      string `json:"label,omitempty"`
	Technology string `json:"technology,omitempty"`
}

// ResponseDiagram response object.
type ResponseDiagram interface {
	// MustMarshal serialises the result as JSON.
	MustMarshal() []byte
}

// ClientInputToGraph client to convert user input inquiry to the DiagramGraph.
type ClientInputToGraph interface {
	Do(context.Context, string) (DiagramGraph, error)
}

// ClientGraphToDiagram client to convert DiagramGraph to diagram artifact, e.g. svg image.
type ClientGraphToDiagram interface {
	Do(context.Context, DiagramGraph) (ResponseDiagram, error)
}

// HttpClient http base client.
type HttpClient interface {
	Do(req *http.Request) (resp *http.Response, err error)
}

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

// ClientStorage defines the client to communicate to the storage to persist core logic transactions.
type ClientStorage interface {
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
