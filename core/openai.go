package core

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path"
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
	Model       string  `json:"model"`
	Prompt      string  `json:"prompt"`
	Suffix      string  `json:"suffix,omitempty"`
	Stop        string  `json:"stop,omitempty"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
	Temperature float32 `json:"temperature,omitempty"`
	Echo        bool    `json:"echo,omitempty"`
	N           int     `json:"n,omitempty"`
	Stream      bool    `json:"stream,omitempty"`
	BestOf      int     `json:"best_of,omitempty"`
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
			Model:       cfg.Model,
			Prompt:      "",
			Suffix:      "",
			Stop:        "",
			MaxTokens:   cfg.MaxTokens,
			Temperature: cfg.Temperature,
			Echo:        false,
			N:           1,
			Stream:      false,
			BestOf:      1,
		},
		baseURL: baseURLOpenAI,
	}

	for _, fn := range optFns {
		fn(&c)
	}

	if err := resolveConfigurations(&c); err != nil {
		return nil, err
	}

	resolveHTTPClientOpenAI(&c)

	return &c, nil
}

func resolveConfigurations(c *clientOpenAI) error {
	if c.token == "" {
		return errors.New(
			"'Token' must be specified, see: https://platform.openai.com/docs/api-reference/authentication",
		)
	}

	if c.payload.Model == "" {
		c.payload.Model = "code-cushman-001"
	}

	if c.payload.MaxTokens <= 0 || c.payload.MaxTokens > 2048 {
		c.payload.MaxTokens = 1024
	}

	if c.payload.Temperature <= 0 {
		c.payload.Temperature = 0.1
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

func (c *clientOpenAI) Do(ctx context.Context, prompt string) (DiagramGraph, error) {
	payload := c.payload
	payload.Prompt = prompt

	respBytes, err := c.requestHandler(ctx, payload)
	if err != nil {
		return DiagramGraph{}, err
	}

	var resp openAIResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return DiagramGraph{}, err
	}

	if len(resp.Choices) == 0 {
		return DiagramGraph{}, errors.New("openAI could not convert the input prompt")
	}

	s := strings.TrimSpace(resp.Choices[0].Text)
	var o DiagramGraph
	if err := json.Unmarshal(*(*[]byte)(unsafe.Pointer(&s)), &o); err != nil {
		return DiagramGraph{}, err
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
		return nil, err
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, path.Join(c.baseURL, "completions"), &w)

	c.setHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode > 209 {
		return nil, errors.New("error status code: " + strconv.Itoa(resp.StatusCode))
	}

	buf, err := io.ReadAll(resp.Body)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	return buf, nil
}
