package adapter

import (
	"io"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/kislerdm/diagramastext/server/core/port"
)

func TestNewHTTPClient(t *testing.T) {
	type args struct {
		cfg HTTPClientConfig
	}
	tests := []struct {
		name string
		args args
		want port.HTTPClient
	}{
		{
			name: "happy path: default",
			args: args{
				cfg: HTTPClientConfig{
					Timeout: -1,
					Backoff: Backoff{
						MaxIterations: 1,
					},
				},
			},
			want: &httpclient{
				httpClient: &http.Client{Timeout: defaultTimeout},
				backoff: Backoff{
					MaxIterations:             1,
					BackoffTimeMinMillisecond: defaultBackoffTimeMinMillisecond,
					BackoffTimeMaxMillisecond: defaultBackoffTimeMaxMillisecond,
				},
				backoffCounter: map[*http.Request]uint8{},
				mu:             &sync.RWMutex{},
			},
		},
		{
			name: "no backoff",
			args: args{
				cfg: HTTPClientConfig{
					Timeout: -1,
					Backoff: Backoff{
						MaxIterations:             0,
						BackoffTimeMinMillisecond: -1,
						BackoffTimeMaxMillisecond: -1,
					},
				},
			},
			want: &httpclient{
				httpClient: &http.Client{Timeout: defaultTimeout},
				backoff: Backoff{
					MaxIterations:             0,
					BackoffTimeMinMillisecond: defaultBackoffTimeMinMillisecond,
					BackoffTimeMaxMillisecond: defaultBackoffTimeMaxMillisecond,
				},
				backoffCounter: map[*http.Request]uint8{},
				mu:             &sync.RWMutex{},
			},
		},
		{
			name: "flip min and max backoff",
			args: args{
				cfg: HTTPClientConfig{
					Timeout: -1,
					Backoff: Backoff{
						MaxIterations:             1,
						BackoffTimeMinMillisecond: 2,
						BackoffTimeMaxMillisecond: 1,
					},
				},
			},
			want: &httpclient{
				httpClient: &http.Client{Timeout: defaultTimeout},
				backoff: Backoff{
					MaxIterations:             1,
					BackoffTimeMinMillisecond: 1,
					BackoffTimeMaxMillisecond: 2,
				},
				backoffCounter: map[*http.Request]uint8{},
				mu:             &sync.RWMutex{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := NewHTTPClient(tt.args.cfg); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("NewHTTPClient() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

type mockHttpClient struct {
	V       *http.Response
	Err     error
	Counter uint8
}

func (c *mockHttpClient) Do(_ *http.Request) (*http.Response, error) {
	c.Counter++
	if c.Err != nil {
		return nil, c.Err
	}
	return c.V, nil
}

func Test_client_Do(t *testing.T) {
	t.Parallel()

	t.Run(
		"happy path: max iterations reached", func(t *testing.T) {
			// GIVEN
			const maxIterations = 2

			cl := mockHttpClient{
				V: &http.Response{
					StatusCode: http.StatusTooManyRequests,
					Body:       io.NopCloser(strings.NewReader(`foobar`)),
				},
			}

			c := httpclient{
				httpClient: &cl,
				backoff: Backoff{
					MaxIterations:             maxIterations,
					BackoffTimeMinMillisecond: 1,
					BackoffTimeMaxMillisecond: 1,
				},
				backoffCounter: map[*http.Request]uint8{},
				mu:             &sync.RWMutex{},
			}

			// WHEN
			resp, err := c.Do(&http.Request{Method: http.MethodGet})

			// THEN
			if err != nil {
				t.Errorf("unexpected error")
			}
			if !reflect.DeepEqual(
				resp, &port.HTTPResponse{
					StatusCode: http.StatusTooManyRequests,
					Body:       io.NopCloser(strings.NewReader(`foobar`)),
				},
			) {
				t.Errorf("unexpected response")
			}

			if cl.Counter != maxIterations {
				t.Errorf("unexpected number iterations")
			}
		},
	)

	t.Run(
		"happy path: a single iterations", func(t *testing.T) {
			// GIVEN
			cl := mockHttpClient{
				V: &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`foobar`)),
				},
			}

			c := httpclient{
				httpClient: &cl,
				backoff: Backoff{
					MaxIterations:             2,
					BackoffTimeMinMillisecond: 1,
					BackoffTimeMaxMillisecond: 1,
				},
				backoffCounter: map[*http.Request]uint8{},
				mu:             &sync.RWMutex{},
			}

			// WHEN
			resp, err := c.Do(&http.Request{Method: http.MethodGet})

			// THEN
			if err != nil {
				t.Errorf("unexpected error")
			}
			if !reflect.DeepEqual(
				resp, &port.HTTPResponse{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`foobar`)),
				},
			) {
				t.Errorf("unexpected response")
			}

			if cl.Counter != 1 {
				t.Errorf("unexpected number iterations")
			}
		},
	)
}
