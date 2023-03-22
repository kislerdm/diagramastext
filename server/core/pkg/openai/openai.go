// Package openai defines the client to communicate to OpenAI server over http to use the "completions" endpoint.
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
)

// NewOpenAIClient initiates the OpenAI client.
func NewOpenAIClient(cfg Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &Client{
		token:        cfg.Token,
		organization: cfg.Organization,
		maxTokens:    cfg.MaxTokens,
		httpClient:   cfg.HTTPClient,
	}, nil
}

// Config configuration of the OpenAI client.
// see:
//   - https://platform.openai.com/docs/api-reference/authentication
//   - https://platform.openai.com/docs/api-reference/completions
type Config struct {
	// https://platform.openai.com/docs/api-reference/completions/create#completions/create-max_tokens
	MaxTokens int

	// https://platform.openai.com/docs/api-reference/authentication
	Token string

	Organization string

	HTTPClient HTTPClient
}

func (cfg Config) Validate() error {
	if cfg.HTTPClient == nil {
		return errors.New("http client must be set")
	}
	if cfg.Token == "" {
		return errors.New(
			"'Token' must be specified, see: https://platform.openai.com/docs/api-reference/authentication",
		)
	}
	return nil
}

const (
	defaultMaxTokens   = 1000
	defaultTemperature = 0.2
	defaultTopP        = 1
)

// Client defines the OpenAI client object.
type Client struct {
	httpClient   HTTPClient
	token        string
	organization string
	maxTokens    int
}

func (c Client) getMaxTokens(model string) int {
	if c.maxTokens <= 0 || c.maxTokens > modelContextMaxTokes(model) {
		return defaultMaxTokens
	}
	return c.maxTokens
}

func (c Client) Do(ctx context.Context, userPrompt string, systemContent string, model string) ([]byte, error) {
	if err := c.validatePrompt(model, userPrompt, systemContent); err != nil {
		return nil, err
	}

	req, err := c.request(ctx, model, userPrompt, systemContent)
	if err != nil {
		return nil, err
	}

	respBytes, err := c.requestHandler(req)
	if err != nil {
		return nil, err
	}

	return decodeResponse(respBytes, model)
}

type payload interface {
	openAIRequestCompletions | openAIRequestCompletionsChat
}

func newReader[T payload](v T) (io.Reader, error) {
	var w bytes.Buffer
	err := json.NewEncoder(&w).Encode(v)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (c Client) request(ctx context.Context, model, userPrompt, systemContent string) (*http.Request, error) {
	base := openAIRequestBase{
		Model:            model,
		Stop:             []string{"\n"},
		TopP:             defaultTopP,
		MaxTokens:        c.getMaxTokens(model),
		Temperature:      defaultTemperature,
		FrequencyPenalty: 0,
		PresencePenalty:  0,
		N:                1,
	}

	var (
		payload io.Reader
		err     error
	)

	switch model {
	case "gpt-3.5-turbo":
		payload, err = newReader(
			openAIRequestCompletionsChat{
				openAIRequestBase: base,
				Messages: []openAIRequestChatMessage{
					{
						Role:    "system",
						Content: systemContent,
					},
					{
						Role:    "user",
						Content: userPrompt,
					},
				},
			},
		)
	default:
		payload, err = newReader(
			openAIRequestCompletions{
				openAIRequestBase: base,
				Prompt:            systemContent + "\n" + userPrompt + "\n",
				BestOf:            2,
			},
		)
	}

	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, baseURL(model)+"completions", payload)
	return req, nil
}

func (c Client) validatePrompt(model, userPrompt, systemContent string) error {
	if len(userPrompt)+len(systemContent)+c.getMaxTokens(model) > modelContextMaxTokes(model) {
		return errors.New(
			"prompt exceeds the model's context length." +
				"see: https://platform.openai.com/docs/api-reference/completions/create#completions/create-max_tokens",
		)
	}
	return nil
}

func (c Client) setHeader(req *http.Request) {
	req.Header.Add("Authorization", "Bearer "+c.token)
	req.Header.Add("Content-Type", "application/json")
	if c.organization != "" {
		req.Header.Add("OpenAI-Organization", c.organization)
	}
}

func (c Client) requestHandler(req *http.Request) ([]byte, error) {
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

func baseURL(model string) string {
	switch model {
	case "gpt-3.5-turbo":
		return "https://api.openai.com/v1/chat/"
	default:
		return "https://api.openai.com/v1/"
	}
}

func modelContextMaxTokes(model string) int {
	switch model {
	case "gpt-3.5-turbo":
		return 4096
	case "code-davinci-002":
		return 8001
	default:
		return 2049
	}
}

func decodeResponse(respBytes []byte, model string) ([]byte, error) {
	switch model {
	case "gpt-3.5-turbo":
		return decodeChatCompletionsResult(respBytes)
	default:
		return decodeCompletionsResult(respBytes)
	}
}

func decodeChatCompletionsResult(respBytes []byte) ([]byte, error) {
	var resp openAIResponseChat
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, errors.New("unsuccessful prediction")
	}

	s := cleanRawResponse(resp.Choices[0].Message.Content)

	return []byte(s), nil
}

func decodeCompletionsResult(respBytes []byte) ([]byte, error) {
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

// HTTPClient http client to interact with the server.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type openAIRequestBase struct {
	Model            string   `json:"model"`
	Stop             []string `json:"stop,omitempty"`
	TopP             float32  `json:"top_p"`
	MaxTokens        int      `json:"max_tokens,omitempty"`
	Temperature      float32  `json:"temperature,omitempty"`
	FrequencyPenalty float32  `json:"frequency_penalty"`
	PresencePenalty  float32  `json:"presence_penalty"`
	N                int      `json:"n"`
}

type openAIRequestCompletions struct {
	openAIRequestBase
	Prompt string `json:"prompt"`
	BestOf uint8  `json:"best_of"`
}

type openAIRequestChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIRequestCompletionsChat struct {
	openAIRequestBase
	Messages []openAIRequestChatMessage `json:"messages"`
}

type openAIResponseBase struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type openAIResponse struct {
	openAIResponseBase
	Model   string `json:"model"`
	Choices []struct {
		Text         string `json:"text"`
		Index        int    `json:"index"`
		Logprobs     int    `json:"logprobs"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

type openAIResponseChat struct {
	openAIResponseBase
	Choices []struct {
		Index        int    `json:"index"`
		FinishReason string `json:"finish_reason"`
		Message      struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type openAIErrorResponse struct {
	Error *struct {
		Code    *int    `json:"code,omitempty"`
		Message string  `json:"message"`
		Param   *string `json:"param,omitempty"`
		Type    string  `json:"type"`
	} `json:"error,omitempty"`
}
