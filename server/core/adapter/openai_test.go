package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/kislerdm/diagramastext/server/core/port"
)

const mockToken = "sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

func Test_clientOpenAI_setHeader(t *testing.T) {
	t.Parallel()
	t.Run(
		"auth headers, no organization specified", func(t *testing.T) {
			// GIVEN
			c := clientOpenAI{
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
			c := clientOpenAI{
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
				s: `{"nodes":[{"id":"0","label":"Go Web Server","technology":"Go","description":"Authenticates users"},{"id":"1","label":"Kafka","technology":"Kafka","is_database":true},{"id":"2"},{"id":"3","label":"Database","technology":"MySQL","is_database":true}],`,
			},
			want: `{"nodes":[{"id":"0","label":"Go Web Server","technology":"Go","description":"Authenticates users"},{"id":"1","label":"Kafka","technology":"Kafka","is_database":true},{"id":"2"},{"id":"3","label":"Database","technology":"MySQL","is_database":true}]}`,
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

func Test_clientOpenAI_decodeResponse(t *testing.T) {
	t.Parallel()

	t.Run(
		"happy path", func(t *testing.T) {
			// GIVEN
			responseBytes, err := json.Marshal(
				openAIResponse{
					ID:     "foo",
					Object: "bar",
					Model:  "code-davinci-002",
					Choices: []struct {
						Text         string `json:"text"`
						Index        int    `json:"index"`
						Logprobs     int    `json:"logprobs"`
						FinishReason string `json:"finish_reason"`
					}{
						{
							Text: `"nodes":["id":"0"]`,
						},
					},
				},
			)
			if err != nil {
				t.Fatal(err)
			}
			// WHEN
			got, err := decodeResponse(responseBytes)
			if err != nil {
				t.Fatal(err)
			}

			// THEN
			if !reflect.DeepEqual(got, []byte(`{"nodes":["id":"0"]}`)) {
				t.Fatal("unexpected response")
			}
		},
	)

	t.Run(
		"unhappy path: empty response", func(t *testing.T) {
			// GIVEN
			responseBytes := []byte(`{"id":"0"}`)

			// WHEN
			_, err := decodeResponse(responseBytes)

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
			_, err := decodeResponse(responseBytes)

			// THEN
			if err == nil {
				t.Errorf("unmarshalling errors is expected")
			}
		},
	)
}

func TestNewOpenAIClient(t *testing.T) {
	type args struct {
		cfg ConfigOpenAI
	}
	tests := []struct {
		name    string
		args    args
		want    port.ModelInference
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				cfg: ConfigOpenAI{Token: mockToken, MaxTokens: 100},
			},
			want: &clientOpenAI{
				httpClient: defaultHttpClient,
				token:      mockToken,
				baseURL:    baseURLOpenAI,
				maxTokens:  100,
			},
			wantErr: false,
		},
		{
			name: "happy path: fixed max tokens",
			args: args{
				cfg: ConfigOpenAI{Token: mockToken, MaxTokens: -100},
			},
			want: &clientOpenAI{
				httpClient: defaultHttpClient,
				token:      mockToken,
				baseURL:    baseURLOpenAI,
				maxTokens:  defaultMaxTokens,
			},
			wantErr: false,
		},
		{
			name: "happy path: mock client",
			args: args{
				cfg: ConfigOpenAI{Token: "mock"},
			},
			want:    port.MockModelInference{},
			wantErr: false,
		},
		{
			name: "unhappy path: invalid config",
			args: args{
				cfg: ConfigOpenAI{},
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

func Test_clientOpenAI_Do(t *testing.T) {
	type fields struct {
		httpClient port.HTTPClient
		token      string
		baseURL    string
		maxTokens  int
	}
	type args struct {
		ctx context.Context
		cfg port.ModelInferenceConfig
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				httpClient: port.MockHTTPClient{
					V: &port.HTTPResponse{
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
				ctx: context.TODO(),
				cfg: port.ModelInferenceConfig{
					Prompt: "foobar",
					Model:  "code-davinci-002",
					BestOf: 2,
				},
			},
			want:    []byte(`{"nodes":[{"id":"0"}]}`),
			wantErr: false,
		},
		{
			name: "unhappy path: invalid prompt",
			fields: fields{
				maxTokens: 10,
			},
			args: args{
				context.TODO(),
				port.ModelInferenceConfig{
					Prompt: randomString(10000),
					Model:  "foobar",
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "unhappy path: high rate",
			fields: fields{
				httpClient: port.MockHTTPClient{
					V: &port.HTTPResponse{
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
				ctx: context.TODO(),
				cfg: port.ModelInferenceConfig{
					Prompt: "foobar",
				},
			},
			want:    nil,
			wantErr: true,
		},
	}

	t.Parallel()

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c := clientOpenAI{
					httpClient: tt.fields.httpClient,
					token:      tt.fields.token,
					baseURL:    tt.fields.baseURL,
					maxTokens:  tt.fields.maxTokens,
				}
				got, err := c.Do(tt.args.ctx, tt.args.cfg)
				if (err != nil) != tt.wantErr {
					t.Errorf("Do() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("Do() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}
