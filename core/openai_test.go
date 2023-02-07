package core

import (
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestNewOpenAIClient(t *testing.T) {
	type args struct {
		cfg    ConfigOpenAI
		optFns []func(client *clientOpenAI)
	}

	const mockToken = "sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

	tests := []struct {
		name    string
		args    args
		want    ClientInputToGraph
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
				},
				token:   mockToken,
				baseURL: baseURLOpenAI,
			},
			wantErr: false,
		},
		{
			name: "happy path: overwrite model, max tokens and temperature",
			args: args{
				cfg: ConfigOpenAI{
					Token:       mockToken,
					Model:       "code-davinci-002",
					Temperature: 0.5,
					MaxTokens:   100,
				},
			},
			want: &clientOpenAI{
				httpClient: &http.Client{
					Timeout: defaultTimeoutOpenAI,
				},
				payload: openAIRequest{
					Model:       "code-davinci-002",
					Stop:        []string{"\n"},
					MaxTokens:   100,
					Temperature: 0.5,
					TopP:        defaultTopP,
				},
				token:   mockToken,
				baseURL: baseURLOpenAI,
			},
			wantErr: false,
		},
		{
			name: "happy path: temperature outside the limits",
			args: args{
				cfg: ConfigOpenAI{
					Token:       mockToken,
					Temperature: 100,
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
