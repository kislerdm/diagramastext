package adapter

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/kislerdm/diagramastext/server/core/port"
)

type inquiry struct {
	Prompt                 string
	UserID                 string
	IsRegisteredUser       bool
	OptOutFromSavingPrompt bool
}

const (
	promptLengthMin = 3

	promptLengthMaxBase       = 100
	promptLengthMaxRegistered = 300
)

func (v inquiry) Validate() error {
	prompt := strings.ReplaceAll(v.Prompt, "\n", "")
	if v.IsRegisteredUser {
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

// NewInquiryHTTPDriver creates the inquiry to be processed using the input from a http request.
func NewInquiryHTTPDriver(body io.Reader, headers http.Header) (port.Input, error) {
	var req struct {
		Prompt string `json:"prompt"`
	}
	if err := json.NewDecoder(body).Decode(&req); err != nil {
		return nil, err
	}

	o := inquiry{
		Prompt: req.Prompt,
	}

	userProfileFromHTTPHeaders(&o, headers)

	return o, nil
}

func userProfileFromHTTPHeaders(inquery *inquiry, headers http.Header) {
	// FIXME: change when the auth layer is implemented
	inquery.UserID = "NA"
}
