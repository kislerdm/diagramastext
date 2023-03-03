package openai

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

	errs "github.com/kislerdm/diagramastext/server/core/errors"
)

// HttpClient http base client.
type HttpClient interface {
	Do(req *http.Request) (resp *http.Response, err error)
}

// Request model inference request.
type Request struct {
	BestOf uint8
	Prompt string
	Model  string
}

// Client interface defining the client to infer a model to convert user's prompt to a serialised data structure.
type Client interface {
	Do(ctx context.Context, req Request) ([]byte, error)
}

/*
	Defines client to communicate to OpenAI over http
*/

// ConfigOpenAI configuration of the OpenAI client.
// see:
//   - https://platform.openai.com/docs/api-reference/authentication
//   - https://platform.openai.com/docs/api-reference/completions
type ConfigOpenAI struct {
	// https://platform.openai.com/docs/api-reference/completions/create#completions/create-max_tokens
	MaxTokens int

	// https://platform.openai.com/docs/api-reference/authentication
	Token string

	// https://platform.openai.com/docs/api-reference/requesting-organization
	Organization string
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
	BestOf           uint8    `json:"best_of"`
}

type clientOpenAI struct {
	httpClient   HttpClient
	payload      openAIRequest
	token        string
	organization string
	baseURL      string
}

// NewClient initiates the client to communicate with the plantuml server.
func NewClient(cfg ConfigOpenAI, optFns ...func(client *clientOpenAI)) (Client, error) {
	c := clientOpenAI{
		httpClient:   nil,
		token:        cfg.Token,
		organization: cfg.Organization,
		payload: openAIRequest{
			Model:            defaultModelOpenAI,
			Prompt:           "",
			Stop:             []string{"\n"},
			MaxTokens:        cfg.MaxTokens,
			TopP:             defaultTopP,
			Temperature:      defaultTemperature,
			FrequencyPenalty: 0,
			PresencePenalty:  0,
			BestOf:           defaultBestOf,
		},
		baseURL: baseURLOpenAI,
	}

	for _, fn := range optFns {
		fn(&c)
	}

	if err := resolveConfigurations(&c); err != nil {
		return nil, errs.Error{
			Service: errs.ServiceOpenAI,
			Message: err.Error(),
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

	if c.payload.MaxTokens <= 0 || c.payload.MaxTokens > 2048 {
		c.payload.MaxTokens = defaultMaxTokens
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
	defaultModelOpenAI   = "code-davinci-002"
	defaultMaxTokens     = 200
	defaultTemperature   = 0.2
	defaultTopP          = 1
	defaultBestOf        = 2
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

func (c *clientOpenAI) Do(ctx context.Context, req Request) ([]byte, error) {
	if err := c.validatePrompt(req.Prompt); err != nil {
		return nil, err
	}

	payload := c.payload

	payload.Prompt = req.Prompt
	if req.BestOf > 0 {
		payload.BestOf = req.BestOf
	}

	if req.Model != "" {
		payload.Model = req.Model
	}

	respBytes, err := c.requestHandler(ctx, payload)
	if err != nil {
		return nil, err
	}

	return c.decodeResponse(ctx, respBytes)
}

func (c *clientOpenAI) modelContextMaxTokes() int {
	switch c.payload.Model {
	case "code-davinci-002":
		return 8000
	case "code-cushman-001":
		return 2048
	default:
		return 2048
	}
}

func (c *clientOpenAI) validatePrompt(prompt string) error {
	if len(prompt)+c.payload.MaxTokens > c.modelContextMaxTokes() {
		return errors.New(
			"prompt exceeds the model's context length." +
				"see: https://platform.openai.com/docs/api-reference/completions/create#completions/create-max_tokens",
		)
	}
	return nil
}

func (c *clientOpenAI) setHeader(req *http.Request) {
	req.Header.Add("Authorization", "Bearer "+c.token)
	req.Header.Add("Content-Type", "application/json")
	if c.organization != "" {
		req.Header.Add("Organization", c.organization)
	}
}

func (c *clientOpenAI) decodeResponse(ctx context.Context, respBytes []byte) ([]byte, error) {
	var resp openAIResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, errs.Error{
			Service: errs.ServiceOpenAI,
			Message: err.Error(),
		}
	}

	if len(resp.Choices) == 0 {
		return nil, errs.Error{
			Service: errs.ServiceOpenAI,
			Message: "openAI could not convert the input prompt",
		}
	}

	s := cleanRawResponse(resp.Choices[0].Text)

	return []byte(s), nil
}

func cleanRawResponse(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, ",")
	if s[:1] != "{" {
		s = "{" + s
	}
	if s[len(s)-1:] != "}" {
		s += "}"
	}
	return s
}

func (c *clientOpenAI) requestHandler(ctx context.Context, payload openAIRequest) ([]byte, error) {
	var w bytes.Buffer
	err := json.NewEncoder(&w).Encode(payload)
	if err != nil {
		return nil, errs.Error{
			Service: errs.ServiceOpenAI,
			Message: err.Error(),
		}
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"completions", &w)

	c.setHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errs.Error{
			Service: errs.ServiceOpenAI,
			Message: err.Error(),
		}
	}

	if resp.StatusCode > 209 {
		o := errs.Error{
			Service:                   errs.ServiceOpenAI,
			Message:                   "error status code: " + strconv.Itoa(resp.StatusCode),
			ServiceResponseStatusCode: resp.StatusCode,
		}

		var e openAIErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&e); err == nil {
			if v := e.Error; v != nil {
				o.Message = v.Message
			}
		}

		return nil, o
	}

	buf, err := io.ReadAll(resp.Body)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return nil, errs.Error{
			Service: errs.ServiceOpenAI,
			Message: err.Error(),
		}
	}

	return buf, nil
}

type openAIErrorResponse struct {
	Error *struct {
		Code    *int    `json:"code,omitempty"`
		Message string  `json:"message"`
		Param   *string `json:"param,omitempty"`
		Type    string  `json:"type"`
	} `json:"error,omitempty"`
}
