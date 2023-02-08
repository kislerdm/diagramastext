package core

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

/*
	Defines client to communicate to OpenAI over http
*/

// ConfigOpenAI configuration of the OpenAI client.
// see:
//   - https://platform.openai.com/docs/api-reference/authentication
//   - https://platform.openai.com/docs/api-reference/completions
type ConfigOpenAI struct {
	// https://platform.openai.com/docs/api-reference/authentication
	Token string

	// https://platform.openai.com/docs/api-reference/requesting-organization
	Organization string

	// https://platform.openai.com/docs/api-reference/completions/create#completions/create-model
	Model string

	// https://platform.openai.com/docs/api-reference/completions/create#completions/create-max_tokens
	MaxTokens int

	// https://platform.openai.com/docs/api-reference/completions/create#completions/create-temperature
	Temperature float32
}

type openAIRequest struct {
	Model            string   `json:"model"`
	Prompt           string   `json:"prompt"`
	Stop             []string `json:"stop,omitempty"`
	MaxTokens        int      `json:"max_tokens,omitempty"`
	Temperature      float32  `json:"temperature,omitempty"`
	TopP             float32  `json:"top_p"`
	FrequencyPenalty float32  `json:"frequency_penalty"`
	PresencePenalty  float32  `json:"presence_penalty"`
}

type clientOpenAI struct {
	httpClient   HttpClient
	payload      openAIRequest
	token        string
	organization string
	baseURL      string
}

// NewOpenAIClient initiates the client to communicate with the plantuml server.
func NewOpenAIClient(cfg ConfigOpenAI, optFns ...func(client *clientOpenAI)) (ClientInputToGraph, error) {
	c := clientOpenAI{
		httpClient:   nil,
		token:        cfg.Token,
		organization: cfg.Organization,
		payload: openAIRequest{
			Model:            cfg.Model,
			Prompt:           "",
			Stop:             []string{"\n"},
			MaxTokens:        cfg.MaxTokens,
			TopP:             defaultTopP,
			Temperature:      cfg.Temperature,
			FrequencyPenalty: 0,
			PresencePenalty:  0,
		},
		baseURL: baseURLOpenAI,
	}

	for _, fn := range optFns {
		fn(&c)
	}

	if err := resolveConfigurations(&c); err != nil {
		return nil, Error{
			Service:                   ServiceOpenAI,
			Stage:                     StageInit,
			ServiceResponseStatusCode: 0,
			Message:                   err.Error(),
		}
	}

	resolveHTTPClientOpenAI(&c)

	return &c, nil
}

// WithHTTPClientOpenAI overwrite the OpenAI HTTP client.
func WithHTTPClientOpenAI(c HttpClient) func(o *clientOpenAI) {
	return func(o *clientOpenAI) {
		o.httpClient = c
	}
}

func resolveConfigurations(c *clientOpenAI) error {
	if c.token == "" {
		return errors.New(
			"'Token' must be specified, see: https://platform.openai.com/docs/api-reference/authentication",
		)
	}

	if c.payload.Model == "" {
		c.payload.Model = defaultModelOpenAI
	}

	if c.payload.MaxTokens <= 0 || c.payload.MaxTokens > 2048 {
		c.payload.MaxTokens = defaultMaxTokens
	}

	if c.payload.Temperature <= 0 || c.payload.Temperature > 1 {
		c.payload.Temperature = defaultTemperature
	}

	return nil
}

func resolveHTTPClientOpenAI(c *clientOpenAI) {
	if c.httpClient == nil {
		c.httpClient = &http.Client{Timeout: defaultTimeoutOpenAI}
	}
}

const (
	baseURLOpenAI        = "https://api.openai.com/v1/"
	defaultTimeoutOpenAI = 3 * time.Minute
	defaultModelOpenAI   = "code-cushman-001"
	defaultMaxTokens     = 768
	defaultTemperature   = 0
	defaultTopP          = 1
)

type openAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Text         string `json:"text"`
		Index        int    `json:"index"`
		Logprobs     int    `json:"logprobs"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

//go:embed c4diagram.prompt
var promptInputComment string

func (c *clientOpenAI) Do(ctx context.Context, prompt string) (DiagramGraph, error) {
	payload := c.payload
	payload.Prompt = promptInputComment + "\n" + strings.ReplaceAll(prompt, "\n", "") + "\n"

	respBytes, err := c.requestHandler(ctx, payload)
	if err != nil {
		return DiagramGraph{}, Error{
			Service:                   ServiceOpenAI,
			Stage:                     StageRequest,
			Message:                   err.Error(),
			ServiceResponseStatusCode: 0,
		}
	}

	var resp openAIResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return DiagramGraph{}, Error{
			Service:                   ServiceOpenAI,
			Stage:                     StageDeserialization,
			Message:                   err.Error(),
			ServiceResponseStatusCode: 0,
		}
	}

	if len(resp.Choices) == 0 {
		return DiagramGraph{}, Error{
			Service:                   ServiceOpenAI,
			Stage:                     StageDeserialization,
			Message:                   "openAI could not convert the input prompt",
			ServiceResponseStatusCode: 0,
		}
	}

	s := strings.TrimSpace(resp.Choices[0].Text)
	var o DiagramGraph
	if err := json.Unmarshal(*(*[]byte)(unsafe.Pointer(&s)), &o); err != nil {
		return DiagramGraph{}, Error{
			Service:                   ServiceOpenAI,
			Stage:                     StageDeserialization,
			Message:                   err.Error(),
			ServiceResponseStatusCode: 0,
		}
	}

	return o, nil
}

func (c *clientOpenAI) setHeader(req *http.Request) {
	req.Header.Add("Authorization", "Bearer "+c.token)
	req.Header.Add("Content-Type", "application/json")
	if c.organization != "" {
		req.Header.Add("Organization", c.organization)
	}
}

// REFACTOR: take to a dedicated helper function.
func (c *clientOpenAI) requestHandler(ctx context.Context, payload openAIRequest) ([]byte, error) {
	var w bytes.Buffer
	err := json.NewEncoder(&w).Encode(payload)
	if err != nil {
		return nil, Error{
			Service:                   ServiceOpenAI,
			Stage:                     StageSerialization,
			Message:                   err.Error(),
			ServiceResponseStatusCode: 0,
		}
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"completions", &w)

	c.setHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, Error{
			Service:                   ServiceOpenAI,
			Stage:                     StageRequest,
			Message:                   err.Error(),
			ServiceResponseStatusCode: 0,
		}
	}

	if resp.StatusCode > 209 {
		return nil, Error{
			Service:                   ServiceOpenAI,
			Stage:                     StageRequest,
			Message:                   "error status code: " + strconv.Itoa(resp.StatusCode),
			ServiceResponseStatusCode: resp.StatusCode,
		}
	}

	buf, err := io.ReadAll(resp.Body)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return nil, Error{
			Service:                   ServiceOpenAI,
			Stage:                     StageDeserialization,
			Message:                   err.Error(),
			ServiceResponseStatusCode: 0,
		}
	}

	return buf, nil
}
