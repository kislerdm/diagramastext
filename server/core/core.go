package core

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aws/smithy-go/rand"
	"github.com/kislerdm/diagramastext/server/core/openai"
	"github.com/kislerdm/diagramastext/server/core/storage"
)

const (
	promptLengthMin = 3

	promptLengthMaxBase       = 100
	promptLengthMaxRegistered = 300

	bestOfBase       = 2
	bestOfRegistered = 3
)

func validatePromptLength(prompt string, max int) error {
	if len(prompt) < promptLengthMin || len(prompt) > max {
		return errors.New(
			"prompt length must be between " + strconv.Itoa(promptLengthMin) + " and " +
				strconv.Itoa(max) + " characters",
		)
	}
	return nil
}

// Request invocation request object.
type Request struct {
	Prompt                 string
	UserID                 string
	IsRegisteredUser       bool
	OptOutFromSavingPrompt bool
}

func (r Request) getBestOf() uint8 {
	if r.IsRegisteredUser {
		return bestOfRegistered
	}
	return bestOfBase
}

// ValidatePrompt validate the request.
func (r Request) validatePrompt() error {
	prompt := strings.ReplaceAll(r.Prompt, "\n", "")
	if r.IsRegisteredUser {
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

type Inquiry struct {
	Request

	Model             string
	PrefixTransformFn func(string) string
}

// ModelInferenceClient the logic to infer the model and predict a diagram graph.
type ModelInferenceClient interface {
	Do(ctx context.Context, req Inquiry) ([]byte, error)
}

type client struct {
	clientModel   openai.Client
	clientStorage storage.Client
}

func (h client) Do(ctx context.Context, req Inquiry) ([]byte, error) {
	if err := req.validatePrompt(); err != nil {
		return nil, err
	}

	cID := storage.CallID{
		RequestID: generateRequestID(),
		UserID:    req.UserID,
	}

	// FIXME: decide on execution path when db write fails
	if h.clientStorage != nil && !req.OptOutFromSavingPrompt {
		if err := h.clientStorage.WritePrompt(
			ctx, storage.UserInput{
				CallID:    cID,
				Prompt:    req.Prompt,
				Timestamp: time.Now().UTC(),
			},
		); err != nil {
			log.Print(err)
		}
	}

	prompt := req.Prompt
	if req.PrefixTransformFn != nil {
		prompt = req.PrefixTransformFn(prompt)
	}

	graphPrediction, err := h.clientModel.Do(
		ctx, openai.Request{
			BestOf: req.getBestOf(),
			Prompt: prompt,
			Model:  req.Model,
		},
	)
	if err != nil {
		return nil, err
	}

	// FIXME: decide on execution path when db write fails
	if h.clientStorage != nil && !req.OptOutFromSavingPrompt {
		if err := h.clientStorage.WriteModelPrediction(
			ctx, storage.ModelOutput{
				CallID:    cID,
				Response:  string(graphPrediction),
				Timestamp: time.Now().UTC(),
			},
		); err != nil {
			if v, err := json.Marshal(graphPrediction); err != nil {
				log.Printf("response: %s", string(v))
			}
		}
	}

	return graphPrediction, err
}

func generateRequestID() string {
	o, _ := rand.NewUUID(rand.Reader).GetUUID()
	return o
}

// NewModelInferenceClientFromConfig initialises the client to infer the model.
func NewModelInferenceClientFromConfig(cfg Config) (ModelInferenceClient, error) {
	clientOpenAI, err := openai.NewClient(cfg.ModelInferenceConfig)
	if err != nil {
		return nil, err
	}

	clientStorage, err := storage.NewPgClient(
		context.Background(), cfg.StorageConfig.DBHost, cfg.StorageConfig.DBName, cfg.StorageConfig.DBUser,
		cfg.StorageConfig.DBPassword,
	)
	if err != nil {
		log.Print(err.Error())
	}

	return &client{
		clientModel:   clientOpenAI,
		clientStorage: clientStorage,
	}, nil
}

// DiagramRenderingHandler functional entrypoint to render the text prompt to a diagram.
type DiagramRenderingHandler func(ctx context.Context, inferenceClient ModelInferenceClient, req Request) (
	[]byte, error,
)
