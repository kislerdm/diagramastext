package server

import (
	"context"
	"net/http"
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
