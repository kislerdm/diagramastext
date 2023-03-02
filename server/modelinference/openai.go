package modelinference

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

	errs "github.com/kislerdm/diagramastext/server/errors"
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
	BestOf           int8     `json:"best_of"`
}

type clientOpenAI struct {
	httpClient   HttpClient
	payload      openAIRequest
	token        string
	organization string
	baseURL      string
}

// NewOpenAIClient initiates the client to communicate with the plantuml server.
func NewOpenAIClient(cfg ConfigOpenAI, optFns ...func(client *clientOpenAI)) (Client, error) {
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
			Stage:   errs.StageInit,
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
	defaultTemperature   = 0.2
	defaultTopP          = 1
	defaultBestOf        = 3
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

func (c *clientOpenAI) Do(ctx context.Context, prompt string) ([]byte, error) {
	payload := c.payload
	payload.Prompt = prompt

	respBytes, err := c.requestHandler(ctx, payload)
	if err != nil {
		return nil, errs.Error{
			Service: errs.ServiceOpenAI,
			Stage:   errs.StageRequest,
			Message: err.Error(),
		}
	}

	return c.decodeResponse(ctx, respBytes)
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
			Stage:   errs.StageDeserialization,
			Message: err.Error(),
		}
	}

	if len(resp.Choices) == 0 {
		return nil, errs.Error{
			Service: errs.ServiceOpenAI,
			Stage:   errs.StageDeserialization,
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

// REFACTOR: take to a dedicated helper function.
func (c *clientOpenAI) requestHandler(ctx context.Context, payload openAIRequest) ([]byte, error) {
	var w bytes.Buffer
	err := json.NewEncoder(&w).Encode(payload)
	if err != nil {
		return nil, errs.Error{
			Service: errs.ServiceOpenAI,
			Stage:   errs.StageSerialization,
			Message: err.Error(),
		}
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"completions", &w)

	c.setHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errs.Error{
			Service: errs.ServiceOpenAI,
			Stage:   errs.StageRequest,
			Message: err.Error(),
		}
	}

	if resp.StatusCode > 209 {
		o := errs.Error{
			Service:                   errs.ServiceOpenAI,
			Stage:                     errs.StageResponse,
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
			Stage:   errs.StageDeserialization,
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
