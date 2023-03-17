package adapter

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

	"github.com/kislerdm/diagramastext/server/core/port"
)

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
}

func (cfg ConfigOpenAI) Validate() error {
	if cfg.Token == "" {
		return errors.New(
			"'Token' must be specified, see: https://platform.openai.com/docs/api-reference/authentication",
		)
	}
	return nil
}

type clientOpenAI struct {
	httpClient port.HTTPClient
	token      string
	baseURL    string
	maxTokens  int
}

const (
	baseURLOpenAI        = "https://api.openai.com/v1/"
	defaultTimeoutOpenAI = 3 * time.Minute
	defaultMaxTokens     = 200
	defaultTemperature   = 0.2
	defaultTopP          = 1
)

var defaultHttpClient = NewHTTPClient(
	HTTPClientConfig{
		Timeout: defaultTimeoutOpenAI,
		Backoff: Backoff{
			MaxIterations:             2,
			BackoffTimeMinMillisecond: 50,
			BackoffTimeMaxMillisecond: 300,
		},
	},
)

// NewOpenAIClient initiates the OpenAI client.
func NewOpenAIClient(cfg ConfigOpenAI) (port.ModelInference, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	if cfg.Token == "mock" {
		return port.MockModelInference{}, nil
	}

	c := clientOpenAI{
		token:      cfg.Token,
		maxTokens:  cfg.MaxTokens,
		baseURL:    baseURLOpenAI,
		httpClient: defaultHttpClient,
	}

	resolveConfigurations(&c)

	return &c, nil
}

func resolveConfigurations(c *clientOpenAI) {
	if c.maxTokens <= 0 || c.maxTokens > 2048 {
		c.maxTokens = defaultMaxTokens
	}
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

func (c clientOpenAI) Do(ctx context.Context, cfg port.ModelInferenceConfig) ([]byte, error) {
	if err := c.validatePrompt(cfg.Model, cfg.Prompt); err != nil {
		return nil, err
	}

	payload := openAIRequest{
		Model:            cfg.Model,
		Prompt:           cfg.Prompt,
		Stop:             []string{"\n"},
		MaxTokens:        c.maxTokens,
		TopP:             defaultTopP,
		Temperature:      defaultTemperature,
		FrequencyPenalty: 0,
		PresencePenalty:  0,
		BestOf:           cfg.BestOf,
	}

	respBytes, err := c.requestHandler(ctx, payload)
	if err != nil {
		return nil, err
	}

	return decodeResponse(respBytes)
}

func modelContextMaxTokes(model string) int {
	switch model {
	case "code-davinci-002":
		return 8000
	case "code-cushman-001":
		return 2048
	default:
		return 2048
	}
}

func (c clientOpenAI) validatePrompt(model, prompt string) error {
	if len(prompt)+c.maxTokens > modelContextMaxTokes(model) {
		return errors.New(
			"prompt exceeds the model's context length." +
				"see: https://platform.openai.com/docs/api-reference/completions/create#completions/create-max_tokens",
		)
	}
	return nil
}

func (c clientOpenAI) setHeader(req *http.Request) {
	req.Header.Add("Authorization", "Bearer "+c.token)
	req.Header.Add("Content-Type", "application/json")
}

type openAIErrorResponse struct {
	Error *struct {
		Code    *int    `json:"code,omitempty"`
		Message string  `json:"message"`
		Param   *string `json:"param,omitempty"`
		Type    string  `json:"type"`
	} `json:"error,omitempty"`
}

func (c clientOpenAI) requestHandler(ctx context.Context, payload openAIRequest) ([]byte, error) {
	var w bytes.Buffer
	err := json.NewEncoder(&w).Encode(payload)
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"completions", &w)
	c.setHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode > 209 {
		var e openAIErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&e); err == nil {
			if v := e.Error; v != nil {
				return nil, errors.New(v.Message)
			}
		}
		return nil, errors.New("error status code: " + strconv.Itoa(resp.StatusCode))
	}

	buf, err := io.ReadAll(resp.Body)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func decodeResponse(respBytes []byte) ([]byte, error) {
	var resp openAIResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, errors.New("unsuccessful prediction")
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
