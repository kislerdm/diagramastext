package openai

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"math/rand"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"
)

func randomString(length int) string {
	const charset = "abcdef"
	var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	var b = make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

const mockToken = "sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

func Test_clientOpenAI_setHeader(t *testing.T) {
	t.Parallel()
	t.Run(
		"auth headers, no organization specified", func(t *testing.T) {
			// GIVEN
			c := Client{
				token: mockToken,
			}
			req := http.Request{
				Header: make(map[string][]string),
			}

			// WHEN
			c.setHeader(&req)

			// THEN
			if req.Header.Get("Authorization") != "Bearer "+mockToken {
				t.Errorf("header Authorization must be set")
				return
			}
			if req.Header.Get("Content-Type") != "application/json" {
				t.Errorf("header Content-Type must be set as application/json")
				return
			}
		},
	)
	t.Run(
		"auth headers, organization specified", func(t *testing.T) {
			// GIVEN
			const wantOrganization = "foobar"
			c := Client{
				token:        mockToken,
				organization: wantOrganization,
			}
			req := http.Request{
				Header: make(map[string][]string),
			}

			// WHEN
			c.setHeader(&req)

			// THEN
			if req.Header.Get("Authorization") != "Bearer "+mockToken {
				t.Errorf("header Authorization must be set")
				return
			}
			if req.Header.Get("Content-Type") != "application/json" {
				t.Errorf("header Content-Type must be set as application/json")
				return
			}
			if req.Header.Get("OpenAI-Organization") != wantOrganization {
				t.Errorf("header OpenAI-Organization must be set as " + wantOrganization)
				return
			}
		},
	)
}

func Test_cleanRawResponse(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "faulty json end",
			args: args{
				s: `{"nodes":[{"id":"0","label":"Go Web Server","technology":"Go","description":"Authenticates users"},{"id":"1","label":"Kafka","technology":"Kafka","database":true},{"id":"2"},{"id":"3","label":"Database","technology":"MySQL","database":true}],`,
			},
			want: `{"nodes":[{"id":"0","label":"Go Web Server","technology":"Go","description":"Authenticates users"},{"id":"1","label":"Kafka","technology":"Kafka","database":true},{"id":"2"},{"id":"3","label":"Database","technology":"MySQL","database":true}]}`,
		},
		{
			name: `"nodes":[{"id":"0"}]`,
			args: args{
				s: "\n" + `"nodes":[{"id":"0"}]` + "\n",
			},
			want: `{"nodes":[{"id":"0"}]}`,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := cleanRawResponse(tt.args.s); got != tt.want {
					t.Errorf("cleanRawResponse() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_clientOpenAI_decodeResponseGPT35Turbo(t *testing.T) {
	const model = "gpt-3.5-turbo"

	t.Parallel()

	t.Run(
		"happy path", func(t *testing.T) {
			// GIVEN
			responseBytes := []byte(`{"id":"chatcmpl-731P1AqUCWr2iVKnmlMIMCQoufu40","object":"chat.completion","created":1680954731,"model":"gpt-3.5-turbo-0301","usage":{"prompt_tokens":781,"completion_tokens":66,"total_tokens":847},"choices":[{"message":{"role":"assistant","content":"C4 diagram with a Go CLI interacting with a Rust backend:\n\n{\"nodes\":[{\"id\":\"0\",\"label\":\"Go CLI\",\"technology\":\"Go\"},{\"id\":\"1\",\"label\":\"Rust Backend\",\"technology\":\"Rust\"}],\"links\":[{\"from\":\"0\",\"to\":\"1\",\"label\":\"interacts with\",\"technology\":\"HTTP\"}]}"},"finish_reason":"stop","index":0}]}`)
			// WHEN
			gotRaw, got, usageTokensPrompt, usageTokensCompletions, err := decodeResponse(responseBytes, model)
			if err != nil {
				t.Fatal(err)
			}

			// THEN
			if !reflect.DeepEqual(
				got,
				[]byte(`{"nodes":[{"id":"0","label":"Go CLI","technology":"Go"},{"id":"1","label":"Rust Backend","technology":"Rust"}],"links":[{"from":"0","to":"1","label":"interacts with","technology":"HTTP"}]}`),
			) {
				t.Fatal("unexpected response")
			}

			if gotRaw != `"C4 diagram with a Go CLI interacting with a Rust backend:\n\n{\"nodes\":[{\"id\":\"0\",\"label\":\"Go CLI\",\"technology\":\"Go\"},{\"id\":\"1\",\"label\":\"Rust Backend\",\"technology\":\"Rust\"}],\"links\":[{\"from\":\"0\",\"to\":\"1\",\"label\":\"interacts with\",\"technology\":\"HTTP\"}]}"` {
				t.Fatal("unexpected raw response")
			}

			if usageTokensPrompt != 781 {
				t.Fatal("unexpected prompt tokens response")
			}

			if usageTokensCompletions != 66 {
				t.Fatal("unexpected completion tokens response")
			}
		},
	)

	t.Run(
		"unhappy path: empty response", func(t *testing.T) {
			// GIVEN
			responseBytes := []byte(`{"id":"0"}`)

			// WHEN
			_, _, _, _, err := decodeResponse(responseBytes, model)

			// THEN
			if !reflect.DeepEqual(err, errors.New("unsuccessful prediction")) {
				t.Fatal("unexpected error: unsuccessful prediction")
			}
		},
	)

	t.Run(
		"unhappy path: unmarshalling", func(t *testing.T) {
			// GIVEN
			responseBytes := []byte(`{"id":"0"`)

			// WHEN
			_, _, _, _, err := decodeResponse(responseBytes, model)

			// THEN
			if err == nil {
				t.Errorf("unmarshalling errors is expected")
			}
		},
	)
}

func Test_clientOpenAI_decodeResponseCodeDavinci(t *testing.T) {
	const model = "code-davinci-002"

	t.Parallel()

	t.Run(
		"happy path", func(t *testing.T) {
			// GIVEN
			responseBytes, err := json.Marshal(
				openAIResponse{
					openAIResponseBase: openAIResponseBase{
						ID:     "foo",
						Object: "chat.completion",
						Usage: struct {
							PromptTokens     uint16 `json:"prompt_tokens"`
							CompletionTokens uint16 `json:"completion_tokens"`
							TotalTokens      int    `json:"total_tokens"`
						}{
							100,
							10,
							110,
						},
					},
					Model: model,
					Choices: []struct {
						Text         string `json:"text"`
						Index        int    `json:"index"`
						Logprobs     int    `json:"logprobs"`
						FinishReason string `json:"finish_reason"`
					}{
						{
							Index: 0,
							Text:  `{"nodes":["id":"0"]}`,
						},
					},
				},
			)
			if err != nil {
				t.Fatal(err)
			}
			// WHEN
			gotRaw, got, usageTokensPrompt, usageTokensCompletions, err := decodeResponse(responseBytes, model)
			if err != nil {
				t.Fatal(err)
			}

			// THEN
			if !reflect.DeepEqual(got, []byte(`{"nodes":["id":"0"]}`)) {
				t.Fatal("unexpected response")
			}

			if gotRaw != `{"nodes":["id":"0"]}` {
				t.Fatal("unexpected raw response")
			}

			if usageTokensPrompt != 100 {
				t.Fatal("unexpected prompt tokens response")
			}

			if usageTokensCompletions != 10 {
				t.Fatal("unexpected completion tokens response")
			}
		},
	)

	t.Run(
		"unhappy path: empty response", func(t *testing.T) {
			// GIVEN
			responseBytes := []byte(`{"id":"0"}`)

			// WHEN
			_, _, _, _, err := decodeResponse(responseBytes, model)

			// THEN
			if !reflect.DeepEqual(err, errors.New("unsuccessful prediction")) {
				t.Fatal("unexpected error: unsuccessful prediction")
			}
		},
	)

	t.Run(
		"unhappy path: unmarshalling", func(t *testing.T) {
			// GIVEN
			responseBytes := []byte(`{"id":"0"`)

			// WHEN
			_, _, _, _, err := decodeResponse(responseBytes, model)

			// THEN
			if err == nil {
				t.Errorf("unmarshalling errors is expected")
			}
		},
	)
}

func TestNewOpenAIClient(t *testing.T) {
	type args struct {
		cfg Config
	}
	tests := []struct {
		name    string
		args    args
		want    *Client
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				cfg: Config{Token: mockToken, MaxTokens: 100, HTTPClient: http.DefaultClient},
			},
			want: &Client{
				httpClient: http.DefaultClient,
				token:      mockToken,
				maxTokens:  100,
			},
			wantErr: false,
		},
		{
			name: "happy path: negative maxTokens",
			args: args{
				cfg: Config{Token: mockToken, MaxTokens: -100, HTTPClient: http.DefaultClient},
			},
			want: &Client{
				httpClient: http.DefaultClient,
				token:      mockToken,
				maxTokens:  -100,
			},
			wantErr: false,
		},
		{
			name: "unhappy path: invalid config, no token",
			args: args{
				cfg: Config{HTTPClient: http.DefaultClient},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "unhappy path: invalid config, no http client",
			args: args{
				cfg: Config{Token: mockToken},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := NewOpenAIClient(tt.args.cfg)
				if (err != nil) != tt.wantErr {
					t.Errorf("NewOpenAIClient() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("NewOpenAIClient() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

type mockHTTPClient struct {
	V   *http.Response
	Err error
}

func (m mockHTTPClient) Do(_ *http.Request) (*http.Response, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.V, nil
}

func Test_clientOpenAI_Do(t *testing.T) {
	type fields struct {
		httpClient HTTPClient
		token      string
		baseURL    string
		maxTokens  int
	}
	type args struct {
		ctx                              context.Context
		model, userPrompt, systemContent string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantRaw string
		wantErr bool
	}{
		{
			name: "happy path: gpt-3.5-turbo",
			fields: fields{
				httpClient: mockHTTPClient{
					V: &http.Response{
						Body: io.NopCloser(
							strings.NewReader(
								`{"id":"0","choices":[{"message":{"content":"{\"nodes\":[{\"id\":\"0\"}]}"},"finish_reason":"stop"}]}`,
							),
						),
						StatusCode: http.StatusOK,
					},
				},
				token:     mockToken,
				maxTokens: 100,
			},
			args: args{
				ctx:           context.TODO(),
				model:         "gpt-3.5-turbo",
				systemContent: "foo",
				userPrompt:    "bar",
			},
			wantRaw: `"{\"nodes\":[{\"id\":\"0\"}]}"`,
			want:    []byte(`{"nodes":[{"id":"0"}]}`),
		},
		{
			name: "happy path: code-davinci-002",
			fields: fields{
				httpClient: mockHTTPClient{
					V: &http.Response{
						Body: io.NopCloser(
							strings.NewReader(
								`{"id":"0","model":"code-davinci-002","choices":[{"text":"{\"nodes\":[{\"id\":\"0\"}]}"}]}`,
							),
						),
						StatusCode: http.StatusOK,
					},
				},
				token:     mockToken,
				maxTokens: 100,
			},
			args: args{
				ctx:           context.TODO(),
				model:         "code-davinci-002",
				systemContent: "foo",
				userPrompt:    "bar",
			},
			wantRaw: `{"nodes":[{"id":"0"}]}`,
			want:    []byte(`{"nodes":[{"id":"0"}]}`),
		},
		{
			name: "happy path: gpt-3.5-turbo, default maxTokens",
			fields: fields{
				httpClient: mockHTTPClient{
					V: &http.Response{
						Body: io.NopCloser(
							strings.NewReader(
								`{"id":"0","choices":[{"message":{"content":"{\"nodes\":[{\"id\":\"0\"}]}"},"finish_reason":"stop"}]}`,
							),
						),
						StatusCode: http.StatusOK,
					},
				},
				token:     mockToken,
				maxTokens: -100,
			},
			args: args{
				ctx:           context.TODO(),
				model:         "gpt-3.5-turbo",
				systemContent: "foo",
				userPrompt:    "bar",
			},
			wantRaw: `"{\"nodes\":[{\"id\":\"0\"}]}"`,
			want:    []byte(`{"nodes":[{"id":"0"}]}`),
		},
		{
			name: "happy path: gpt-3.5-turbo, chat assistance's explanation included",
			fields: fields{
				httpClient: mockHTTPClient{
					V: &http.Response{
						Body: io.NopCloser(
							strings.NewReader(
								`{"id":"chatcmpl-71dU5szgGdOOtLiTDqB5yr0gum3Kz","object":"chat.completion","created":1680624461,"model":"gpt-3.5-turbo-0301","usage":{"prompt_tokens":781,"completion_tokens":173,"total_tokens":954},` +
									`"choices":[{"message":{"role":"assistant","content":"Here's the C4 diagram for a Python web server reading from an external Postgres database:` +
									"\n\n```\n{\n  " +
									`\"title\": \"Python Web Server Reading from External Postgres Database\",\n  \"nodes\": [\n    {\"id\": \"0\", \"label\": \"Web Server\", \"technology\": \"Python\"},\n    {\"id\": \"1\", \"label\": \"Postgres\", \"technology\": \"Postgres\", \"external\": true, \"is_database\": true}\n  ],\n  \"links\": [\n    {\"from\": \"0\", \"to\": \"1\", \"label\": \"reads from Postgres\", \"technology\": \"TCP\"}\n  ],\n  \"footer\": \"C4 Model\"` +
									"\n}\n```\n\nThe diagram shows two nodes: a Python web server and an external Postgres database. The web server reads data from the Postgres database over TCP." + `"},"finish_reason":"stop","index":0}]}`,
							),
						),
						StatusCode: http.StatusOK,
					},
				},
				token:     mockToken,
				maxTokens: -100,
			},
			args: args{
				ctx:           context.TODO(),
				model:         "gpt-3.5-turbo",
				systemContent: "foo",
				userPrompt:    "bar",
			},
			wantRaw: "\"Here's the C4 diagram for a Python web server reading from an external Postgres database:\n\n```\n{\n  \\\"title\\\": \\\"Python Web Server Reading from External Postgres Database\\\",\\n  \\\"nodes\\\": [\\n    {\\\"id\\\": \\\"0\\\", \\\"label\\\": \\\"Web Server\\\", \\\"technology\\\": \\\"Python\\\"},\\n    {\\\"id\\\": \\\"1\\\", \\\"label\\\": \\\"Postgres\\\", \\\"technology\\\": \\\"Postgres\\\", \\\"external\\\": true, \\\"is_database\\\": true}\\n  ],\\n  \\\"links\\\": [\\n    {\\\"from\\\": \\\"0\\\", \\\"to\\\": \\\"1\\\", \\\"label\\\": \\\"reads from Postgres\\\", \\\"technology\\\": \\\"TCP\\\"}\\n  ],\\n  \\\"footer\\\": \\\"C4 Model\\\"\n}\n```\n\nThe diagram shows two nodes: a Python web server and an external Postgres database. The web server reads data from the Postgres database over TCP.\"",
			want:    []byte(`{"title":"Python Web Server Reading from External Postgres Database","nodes":[{"id":"0","label":"Web Server","technology":"Python"},{"id":"1","label":"Postgres","technology":"Postgres","external":true,"is_database":true}],"links":[{"from":"0","to":"1","label":"reads from Postgres","technology":"TCP"}],"footer":"C4 Model"}`),
		},
		{
			name: "unhappy path: invalid prompt",
			fields: fields{
				maxTokens: 10,
			},
			args: args{
				ctx:           context.TODO(),
				model:         "foobar",
				userPrompt:    randomString(10000),
				systemContent: "foo",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "unhappy path: high rate",
			fields: fields{
				httpClient: mockHTTPClient{
					V: &http.Response{
						Body: io.NopCloser(
							strings.NewReader(
								`{"error":{"code":429,"message":"foobar"}}`,
							),
						),
						StatusCode: http.StatusTooManyRequests,
					},
				},
				token:     mockToken,
				maxTokens: 10,
			},
			args: args{
				ctx:           context.TODO(),
				userPrompt:    "foobar",
				systemContent: "qux",
			},
			want:    nil,
			wantErr: true,
		},
	}

	t.Parallel()

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c := Client{
					httpClient: tt.fields.httpClient,
					token:      tt.fields.token,
					maxTokens:  tt.fields.maxTokens,
				}
				gotRaw, got, _, _, err := c.Do(
					tt.args.ctx, tt.args.userPrompt, tt.args.systemContent, tt.args.model,
				)
				if (err != nil) != tt.wantErr {
					t.Errorf("Do() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if gotRaw != tt.wantRaw {
					t.Errorf("Do() gotRaw = %s, want %s", gotRaw, tt.wantRaw)
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("Do() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_cleanRawChatResponse(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "json surrounded by text from both sides",
			args: args{"foo" + chatDescriptionSeparator + `{"bar":"zip"}` + chatDescriptionSeparator + "qux"},
			want: `{"bar":"zip"}`,
		},
		{
			name: "json surrounded by text from left",
			args: args{"foo" + chatDescriptionSeparator + `{"bar":"zip"}`},
			want: `{"bar":"zip"}`,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := cleanRawChatResponse(tt.args.s); got != tt.want {
					t.Errorf("cleanRawChatResponse() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_cleanRawBytesChatResponse(t *testing.T) {
	type args struct {
		respBytes []byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "pure json string",
			args: args{[]byte(`{\"nodes\":[{\"id\":\"0\",\"label\":\"Go CLI\",\"technology\":\"Go\"},{\"id\":\"1\",\"label\":\"Rust Backend\",\"technology\":\"Rust\"}],\"links\":[{\"from\":\"0\",\"to\":\"1\",\"label\":\"interacts with\",\"technology\":\"HTTP\"}]}`)},
			want: []byte(`{\"nodes\":[{\"id\":\"0\",\"label\":\"Go CLI\",\"technology\":\"Go\"},{\"id\":\"1\",\"label\":\"Rust Backend\",\"technology\":\"Rust\"}],\"links\":[{\"from\":\"0\",\"to\":\"1\",\"label\":\"interacts with\",\"technology\":\"HTTP\"}]}`),
		},
		{
			name: "json string preceding text",
			args: args{[]byte(`{"id":"chatcmpl-731P1AqUCWr2iVKnmlMIMCQoufu40","object":"chat.completion","created":1680954731,"model":"gpt-3.5-turbo-0301","usage":{"prompt_tokens":781,"completion_tokens":66,"total_tokens":847},"choices":[{"message":{"role":"assistant","content":"C4 diagram with a Go CLI interacting with a Rust backend:\n\n{\"nodes\":[{\"id\":\"0\",\"label\":\"Go CLI\",\"technology\":\"Go\"},{\"id\":\"1\",\"label\":\"Rust Backend\",\"technology\":\"Rust\"}],\"links\":[{\"from\":\"0\",\"to\":\"1\",\"label\":\"interacts with\",\"technology\":\"HTTP\"}]}"},"finish_reason":"stop","index":0}]}`)},
			want: []byte(`{"id":"chatcmpl-731P1AqUCWr2iVKnmlMIMCQoufu40","object":"chat.completion","created":1680954731,"model":"gpt-3.5-turbo-0301","usage":{"prompt_tokens":781,"completion_tokens":66,"total_tokens":847},"choices":[{"message":{"role":"assistant","content":"C4 diagram with a Go CLI interacting with a Rust backend` +
				chatDescriptionSeparator +
				`{\"nodes\":[{\"id\":\"0\",\"label\":\"Go CLI\",\"technology\":\"Go\"},{\"id\":\"1\",\"label\":\"Rust Backend\",\"technology\":\"Rust\"}],\"links\":[{\"from\":\"0\",\"to\":\"1\",\"label\":\"interacts with\",\"technology\":\"HTTP\"}]}"},"finish_reason":"stop","index":0}]}`),
		},
		{
			name: "json surrounded by text",
			args: args{
				[]byte(
					`{"id":"chatcmpl-71dU5szgGdOOtLiTDqB5yr0gum3Kz","object":"chat.completion","created":1680624461,"model":"gpt-3.5-turbo-0301","usage":{"prompt_tokens":781,"completion_tokens":173,"total_tokens":954},` +
						`"choices":[{"message":{"role":"assistant","content":"Here's the C4 diagram for a Python web server reading from an external Postgres database:` +
						"\n\n```\n{\n  " +
						`\"title\": \"Python Web Server Reading from External Postgres Database\",\n  \"nodes\": [\n    {\"id\": \"0\", \"label\": \"Web Server\", \"technology\": \"Python\"},\n    {\"id\": \"1\", \"label\": \"Postgres\", \"technology\": \"Postgres\", \"external\": true, \"is_database\": true}\n  ],\n  \"links\": [\n    {\"from\": \"0\", \"to\": \"1\", \"label\": \"reads from Postgres\", \"technology\": \"TCP\"}\n  ],\n  \"footer\": \"C4 Model\"` +
						"\n}\n```\n\nThe diagram shows two nodes: a Python web server and an external Postgres database. The web server reads data from the Postgres database over TCP." + `"},"finish_reason":"stop","index":0}]}`,
				),
			},
			want: []byte(
				`{"id":"chatcmpl-71dU5szgGdOOtLiTDqB5yr0gum3Kz","object":"chat.completion","created":1680624461,"model":"gpt-3.5-turbo-0301","usage":{"prompt_tokens":781,"completion_tokens":173,"total_tokens":954},` +
					`"choices":[{"message":{"role":"assistant","content":` +
					`"Here's the C4 diagram for a Python web server reading from an external Postgres database:` + chatDescriptionSeparator +
					`{  \"title\": \"Python Web Server Reading from External Postgres Database\",  \"nodes\": [    {\"id\": \"0\", \"label\": \"Web Server\", \"technology\": \"Python\"},    {\"id\": \"1\", \"label\": \"Postgres\", \"technology\": \"Postgres\", \"external\": true, \"is_database\": true}  ],  \"links\": [    {\"from\": \"0\", \"to\": \"1\", \"label\": \"reads from Postgres\", \"technology\": \"TCP\"}  ],  \"footer\": \"C4 Model\"}` +
					chatDescriptionSeparator + `The diagram shows two nodes: a Python web server and an external Postgres database. The web server reads data from the Postgres database over TCP."}` +
					`,"finish_reason":"stop","index":0}]}`,
			),
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := cleanRawBytesChatResponse(tt.args.respBytes); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("cleanRawBytesChatResponse() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_removeInnerSpaces(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "trivial",
			args: args{`{"foo":"bar"}`},
			want: `{"foo":"bar"}`,
		},
		{
			name: "spaces after key-value separator",
			args: args{`{"foo":   "bar"}`},
			want: `{"foo":"bar"}`,
		},
		{
			name: "spaces after the value quotes",
			args: args{`{"foo":"   bar"}`},
			want: `{"foo":"   bar"}`,
		},
		{
			name: "quoted text",
			args: args{`{"foo":" \"   bar  "}`},
			want: `{"foo":" \"   bar  "}`,
		},
		{
			name: "real example",
			args: args{`{  "title": "Python Web Server Reading from External Postgres Database",  "nodes": [    {"id": "0", "label": "Web Server", "technology": "Python"},    {"id": "1", "label": "Postgres", "technology": "Postgres", "external": true, "is_database": true}  ],  "links": [    {"from": "0", "to": "1", "label": "reads from Postgres", "technology": "TCP"}  ],  "footer": "C4 Model"}`},
			want: `{"title":"Python Web Server Reading from External Postgres Database","nodes":[{"id":"0","label":"Web Server","technology":"Python"},{"id":"1","label":"Postgres","technology":"Postgres","external":true,"is_database":true}],"links":[{"from":"0","to":"1","label":"reads from Postgres","technology":"TCP"}],"footer":"C4 Model"}`,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := removeInnerSpaces(tt.args.s); got != tt.want {
					t.Errorf("removeInnerSpaces() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}
