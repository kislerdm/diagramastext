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
	Prompt string
	User   port.User
}

const (
	promptLengthMin               = 3
	promptLengthMaxBaseUser       = 100
	promptLengthMaxRegisteredUser = 300
)

func (v inquiry) GetUser() port.User {
	return v.User
}

func (v inquiry) Validate() error {
	prompt := strings.ReplaceAll(v.Prompt, "\n", "")
	if v.GetUser().IsRegistered {
		return validatePromptLength(prompt, promptLengthMaxRegisteredUser)
	}
	return validatePromptLength(prompt, promptLengthMaxBaseUser)
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

// NewInquiryDriverHTTP creates the inquiry to be processed using the input from a http request.
func NewInquiryDriverHTTP(body io.Reader, headers http.Header) (port.Input, error) {
	var req struct {
		Prompt string `json:"prompt"`
	}
	if err := json.NewDecoder(body).Decode(&req); err != nil {
		return nil, err
	}

	o := &inquiry{
		Prompt: req.Prompt,
		User:   userProfileFromHTTPHeaders(headers),
	}

	if err := o.Validate(); err != nil {
		return nil, err
	}

	return o, nil
}

func userProfileFromHTTPHeaders(headers http.Header) port.User {
	// FIXME: change when the auth layer is implemented
	return port.User{ID: "NA"}
}
