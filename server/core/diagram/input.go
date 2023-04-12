package diagram

import (
	"errors"
	"strconv"
	"strings"

	"github.com/kislerdm/diagramastext/server/core/internal/utils"
)

type User struct {
	ID           string
	APIToken     string
	IsRegistered bool
}

// Input defines the entrypoint interface.
type Input interface {
	Validate() error
	GetUser() *User
	GetPrompt() string
	GetRequestID() string
}

type MockInput struct {
	Err       error
	Prompt    string
	RequestID string
	User      *User
}

func (v MockInput) Validate() error {
	return v.Err
}

func (v MockInput) GetUser() *User {
	return v.User
}

func (v MockInput) GetPrompt() string {
	return strings.ReplaceAll(v.Prompt, "\n", "")
}

func (v MockInput) GetRequestID() string {
	return v.RequestID
}

type inquiry struct {
	Prompt    string
	RequestID string
	User      *User
}

const (
	promptLengthMin               = 3
	promptLengthMaxBaseUser       = 100
	promptLengthMaxRegisteredUser = 300
)

func (v inquiry) GetPrompt() string {
	return v.Prompt
}

func (v inquiry) GetRequestID() string {
	return v.RequestID
}

func (v inquiry) GetUser() *User {
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

// NewInput initialises the `Input` object.
func NewInput(prompt string, user *User) (Input, error) {
	o := &inquiry{
		Prompt:    prompt,
		User:      user,
		RequestID: utils.NewUUID(),
	}

	if err := o.Validate(); err != nil {
		return nil, err
	}

	return o, nil
}
