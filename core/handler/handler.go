// Package handler defines the client-server communication entrypoint handler.
package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/kislerdm/diagramastext/core"
)

// ResponseError error object to use as response.
type ResponseError struct {
	StatusCode int
	Body       []byte
}

// ParseClientError parses core error.
func ParseClientError(err error) ResponseError {
	msg := "unknown"
	if e, ok := err.(core.Error); ok {
		if e.ServiceResponseStatusCode == http.StatusTooManyRequests {
			return ResponseError{
				StatusCode: e.ServiceResponseStatusCode,
				Body:       []byte("service experiences high load, please try later"),
			}
		}
		switch e.Service {
		case core.ServiceOpenAI:
			msg = "could not recognise diagram description"
		case core.ServiePlantUML:
			msg = "could not generate diagram using provided description"
		}
	}
	return ResponseError{
		StatusCode: http.StatusInternalServerError,
		Body:       []byte(msg),
	}
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

// ReadPrompt reads the user's input prompt.
func ReadPrompt(requestBody []byte) (string, error) {
	type request struct {
		Prompt string `json:"prompt"`
	}

	var r request

	if err := json.Unmarshal(requestBody, &r); err != nil {
		return "", err
	}
	return r.Prompt, nil
}
