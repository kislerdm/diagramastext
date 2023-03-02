package modelinference

import (
	"net/http"
	"reflect"
	"testing"
	"time"
)

const mockToken = "sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

func TestNewOpenAIClient(t *testing.T) {
	type args struct {
		cfg    ConfigOpenAI
		optFns []func(client *clientOpenAI)
	}

	tests := []struct {
		name    string
		args    args
		want    Client
		wantErr bool
	}{
		{
			name: "happy path: default client",
			args: args{
				cfg: ConfigOpenAI{
					Token: mockToken,
				},
			},
			want: &clientOpenAI{
				httpClient: &http.Client{
					Timeout: defaultTimeoutOpenAI,
				},
				payload: openAIRequest{
					Model:       defaultModelOpenAI,
					Stop:        []string{"\n"},
					MaxTokens:   defaultMaxTokens,
					Temperature: defaultTemperature,
					TopP:        defaultTopP,
					BestOf:      defaultBestOf,
				},
				token:   mockToken,
				baseURL: baseURLOpenAI,
			},
			wantErr: false,
		},
		{
			name: "happy path: overwrite http client",
			args: args{
				cfg: ConfigOpenAI{
					Token: mockToken,
				},
				optFns: []func(client *clientOpenAI){
					WithHTTPClientOpenAI(
						&http.Client{
							Timeout: 10 * time.Minute,
						},
					),
				},
			},
			want: &clientOpenAI{
				httpClient: &http.Client{
					Timeout: 10 * time.Minute,
				},
				payload: openAIRequest{
					Model:       defaultModelOpenAI,
					Stop:        []string{"\n"},
					MaxTokens:   defaultMaxTokens,
					Temperature: defaultTemperature,
					TopP:        defaultTopP,
					BestOf:      defaultBestOf,
				},
				token:   mockToken,
				baseURL: baseURLOpenAI,
			},
			wantErr: false,
		},
		{
			name: "happy path: max tokens outside the limits",
			args: args{
				cfg: ConfigOpenAI{
					Token:     mockToken,
					MaxTokens: 10000,
				},
			},
			want: &clientOpenAI{
				httpClient: &http.Client{
					Timeout: defaultTimeoutOpenAI,
				},
				payload: openAIRequest{
					Model:       defaultModelOpenAI,
					Stop:        []string{"\n"},
					MaxTokens:   defaultMaxTokens,
					Temperature: defaultTemperature,
					TopP:        defaultTopP,
					BestOf:      defaultBestOf,
				},
				token:   mockToken,
				baseURL: baseURLOpenAI,
			},
			wantErr: false,
		},
		{
			name: "unhappy path: missing token",
			args: args{
				cfg: ConfigOpenAI{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	t.Parallel()
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := NewOpenAIClient(tt.args.cfg, tt.args.optFns...)
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
				token:        mockToken,
				organization: "foobar",
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
			if req.Header.Get("Organization") != "foobar" {
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
