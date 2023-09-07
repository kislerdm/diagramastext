package diagram

import (
	"errors"
	"strconv"
	"strings"

	"github.com/kislerdm/diagramastext/server/core/internal/utils"
)

// Input defines the entrypoint interface.
type Input interface {
	Validate() error
	GetUserID() string
	GetUserAPIToken() string
	GetPrompt() string
	GetRequestID() string
}

type MockInput struct {
	Err       error
	Prompt    string
	RequestID string
	UserID    string
	APIToken  string
}

func (v MockInput) Validate() error {
	return v.Err
}

func (v MockInput) GetUserID() string {
	return v.UserID
}

func (v MockInput) GetUserAPIToken() string {
	return v.APIToken
}

func (v MockInput) GetPrompt() string {
	return strings.ReplaceAll(v.Prompt, "\n", "")
}

func (v MockInput) GetRequestID() string {
	return v.RequestID
}

type inquiry struct {
	Prompt          string
	RequestID       string
	UserID          string
	APIToken        string
	PromptLengthMax uint16
}

const promptLengthMin = 3

func (v inquiry) GetPrompt() string {
	return v.Prompt
}

func (v inquiry) GetRequestID() string {
	return v.RequestID
}

func (v inquiry) GetUserID() string {
	return v.UserID
}

func (v inquiry) GetUserAPIToken() string {
	return v.APIToken
}

func (v inquiry) Validate() error {
	max := int(v.PromptLengthMax)

	prompt := strings.ReplaceAll(v.Prompt, "\n", "")

	if len(prompt) < promptLengthMin || len(prompt) > max {
		return errors.New(
			"prompt length must be between " + strconv.Itoa(promptLengthMin) + " and " +
				strconv.Itoa(max) + " characters",
		)
	}

	return nil
}

// NewInput initialises the `Input` object.
func NewInput(prompt string, userID string, apiToken string, promptLengthMax uint16) (Input, error) {
	o := &inquiry{
		Prompt:          prompt,
		UserID:          userID,
		PromptLengthMax: promptLengthMax,
		APIToken:        apiToken,
		RequestID:       utils.NewUUID(),
	}

	if err := o.Validate(); err != nil {
		return nil, err
	}

	return o, nil
}
