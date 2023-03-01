package modelinference

import (
	"errors"
	"net/http"
	"strconv"
)

// Request request expected by the `modelinference` client.
type Request struct {
	Prompt string `json:"prompt"`
}

// Graph defines the diagram graph.
type Graph struct {
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

// HttpClient http base client.
type HttpClient interface {
	Do(req *http.Request) (resp *http.Response, err error)
}

const (
	PromptLengthMin = 3
	PromptLengthMax = 500
)

// ValidatePrompt validates input prompt.
func ValidatePrompt(prompt string) error {
	if len(prompt) < PromptLengthMin || len(prompt) > PromptLengthMax {
		return errors.New(
			"prompt length must be between " + strconv.Itoa(PromptLengthMin) + " and " +
				strconv.Itoa(PromptLengthMax) + " characters",
		)
	}
	return nil
}
