package core

import (
	"encoding/json"
	"net/http"
)

// DiagramGraph defines the diagram graph.
type DiagramGraph struct {
	Title  string  `json:"title"`
	Footer string  `json:"footer"`
	Nodes  []*Node `json:"nodes"`
	Links  []*Link `json:"links"`
}

// Node diagram's definition node.
type Node struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	Group      string `json:"group"`
	Technology string `json:"technology"`
	External   bool   `json:"external"`
	IsQueue    bool   `json:"is_queue"`
	IsDatabase bool   `json:"is_database"`
}

// Link diagram's definition link.
type Link struct {
	From       string `json:"from"`
	To         string `json:"to"`
	Direction  string `json:"direction"`
	Label      string `json:"label"`
	Technology string `json:"technology"`
}

// ResponseDiagram response object.
type ResponseDiagram interface {
	// ToJSON encodes the result as JSON.
	ToJSON() []byte
}

// ResponseC4Diagram resulting C4 diagram.
type ResponseC4Diagram struct {
	SVG string `json:"svg"`
}

func (r ResponseC4Diagram) ToJSON() []byte {
	// FIXME(?): add svg validation.
	o, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	return o
}

// ClientInputToGraph client to convert user input inquiry to the DiagramGraph.
type ClientInputToGraph interface {
	Do(string) (DiagramGraph, error)
}

// ClientGraphToDiagram client to convert DiagramGraph to diagram artifact, e.g. svg image.
type ClientGraphToDiagram interface {
	Do(graph DiagramGraph) ([]byte, error)
}

// HttpClient http base client.
type HttpClient interface {
	Get(url string) (resp *http.Response, err error)
}
