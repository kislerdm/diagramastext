package c4container

import (
	"context"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/kislerdm/diagramastext/server/core/diagram"
)

func TestNewC4ContainersHandlerInitHappyPath(t *testing.T) {
	type args struct {
		clientModelInference       diagram.ModelInference
		clientRepositoryPrediction diagram.RepositoryPrediction
		httpClient                 diagram.HTTPClient
	}

	mustNewResult := func(v []byte) diagram.Output {
		o, err := diagram.NewResultSVG(v)
		if err != nil {
			panic(err)
		}
		return o
	}

	tests := []struct {
		name    string
		args    args
		input   diagram.Input
		want    diagram.Output
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				clientModelInference: diagram.MockModelInference{
					V: []byte(`{"nodes":[{"id":"0"}]}`),
				},
				clientRepositoryPrediction: diagram.MockRepositoryPrediction{},
				httpClient: diagram.MockHTTPClient{
					V: &http.Response{
						StatusCode: http.StatusOK,
						Body: io.NopCloser(
							strings.NewReader(
								`<?xml version="1.0" encoding="us-ascii" standalone="no"?>
<svg xmlns="http://www.w3.org/2000/svg" contentstyletype="text/css" height="179px" preserveAspectRatio="none" version="1.1" viewBox="0 0 375 179" width="375px" zoomAndPan="magnify">
<defs></defs>
<g>
	<g id="elem_n0">
		<rect fill="#438DD5" height="52.5938" rx="2.5" ry="2.5" style="stroke:#3C7FC0;stroke-width:0.5;" width="125" x="7" y="11.8301"></rect>
		<text fill="#FFFFFF" font-family="sans-serif" font-size="16" font-weight="bold" lengthAdjust="spacing" textLength="40" x="17" y="36.6816">Web
		</text>
		<text fill="#FFFFFF" font-family="sans-serif" font-size="16" font-weight="bold" lengthAdjust="spacing" textLength="6" x="57" y="36.6816"></text>
		<text fill="#FFFFFF" font-family="sans-serif" font-size="16" font-weight="bold" lengthAdjust="spacing" textLength="59" x="63" y="36.6816">Server
		</text>
		<text fill="#FFFFFF" font-family="sans-serif" font-size="12" font-style="italic" lengthAdjust="spacing" textLength="26" x="56.5" y="51.5938">[Go]
		</text>
	</g>
</g>
</svg>`,
							),
						),
					},
				},
			},
			input: diagram.MockInput{
				Prompt:    "foobar",
				RequestID: "xxxx",
				User: &diagram.User{
					ID: "NA",
				},
			},
			want: mustNewResult(
				[]byte(`<?xml version="1.0" encoding="us-ascii" standalone="no"?>
<svg xmlns="http://www.w3.org/2000/svg" contentstyletype="text/css" height="179px" preserveAspectRatio="none" version="1.1" viewBox="0 0 375 179" width="375px" zoomAndPan="magnify">
<defs></defs>
<g>
	<g id="elem_n0">
		<rect fill="#438DD5" height="52.5938" rx="2.5" ry="2.5" style="stroke:#3C7FC0;stroke-width:0.5;" width="125" x="7" y="11.8301"></rect>
		<text fill="#FFFFFF" font-family="sans-serif" font-size="16" font-weight="bold" lengthAdjust="spacing" textLength="40" x="17" y="36.6816">Web
		</text>
		<text fill="#FFFFFF" font-family="sans-serif" font-size="16" font-weight="bold" lengthAdjust="spacing" textLength="6" x="57" y="36.6816"></text>
		<text fill="#FFFFFF" font-family="sans-serif" font-size="16" font-weight="bold" lengthAdjust="spacing" textLength="59" x="63" y="36.6816">Server
		</text>
		<text fill="#FFFFFF" font-family="sans-serif" font-size="12" font-style="italic" lengthAdjust="spacing" textLength="26" x="56.5" y="51.5938">[Go]
		</text>
	</g>
</g>
</svg>`),
			),
			wantErr: false,
		},
		{
			name: "unhappy path: invalid input",
			args: args{
				clientModelInference:       diagram.MockModelInference{},
				clientRepositoryPrediction: diagram.MockRepositoryPrediction{},
				httpClient:                 diagram.MockHTTPClient{},
			},
			input: diagram.MockInput{
				Err: errors.New("foobar"),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "unhappy path: failed to predict",
			args: args{
				clientModelInference: diagram.MockModelInference{
					Err: errors.New("foobar"),
				},
				clientRepositoryPrediction: diagram.MockRepositoryPrediction{},
				httpClient:                 diagram.MockHTTPClient{},
			},
			input: diagram.MockInput{
				Prompt: "foobar",
				User: &diagram.User{
					ID: "NA",
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "unhappy path: failed to deserialize prediction",
			args: args{
				clientModelInference: diagram.MockModelInference{
					V: []byte(`{"nodes":[{"id":"0"}]`),
				},
				clientRepositoryPrediction: diagram.MockRepositoryPrediction{},
				httpClient:                 diagram.MockHTTPClient{},
			},
			input: diagram.MockInput{
				Prompt: "foobar",
				User: &diagram.User{
					ID: "NA",
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "unhappy path: failed to render diagram",
			args: args{
				clientModelInference: diagram.MockModelInference{
					V: []byte(`{"nodes":[{"id":"0"}]}`),
				},
				clientRepositoryPrediction: diagram.MockRepositoryPrediction{},
				httpClient: diagram.MockHTTPClient{
					Err: errors.New("foobar"),
				},
			},
			input: diagram.MockInput{
				Prompt: "foobar",
				User: &diagram.User{
					ID: "NA",
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
				c, err := NewC4ContainersHandler(
					tt.args.clientModelInference, tt.args.clientRepositoryPrediction, tt.args.httpClient,
				)
				if err != nil {
					t.Fatal(err)
				}

				got, err := c(context.TODO(), tt.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("NewC4ContainersHandler() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("NewC4ContainersHandler() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestNewC4ContainersHandlerInitUnhappyPath(t *testing.T) {
	t.Parallel()

	t.Run(
		"model inference client not provided", func(t *testing.T) {
			// GIVEN
			var clientModelInference diagram.ModelInference
			var clientRepositoryPrediction diagram.RepositoryPrediction
			var httpClient diagram.HTTPClient = diagram.MockHTTPClient{}

			// WHEN
			c, err := NewC4ContainersHandler(clientModelInference, clientRepositoryPrediction, httpClient)

			// THEN
			if c != nil {
				t.Fatalf("unexpected client")
			}

			if err == nil || err.Error() !=
				"diagram/c4container/c4container.go:49: model inference client must be provided" {
				t.Fatalf("unexpected error")
			}
		},
	)

	t.Run(
		"http client not provided", func(t *testing.T) {
			// GIVEN
			var clientModelInference diagram.ModelInference = diagram.MockModelInference{}
			var clientRepositoryPrediction diagram.RepositoryPrediction
			var httpClient diagram.HTTPClient

			// WHEN
			c, err := NewC4ContainersHandler(clientModelInference, clientRepositoryPrediction, httpClient)

			// THEN
			if c != nil {
				t.Fatalf("unexpected client")
			}

			if err == nil || err.Error() != "diagram/c4container/c4container.go:52: http client must be provided" {
				t.Fatalf("unexpected error")
			}
		},
	)
}

func Test_defineBestOf(t *testing.T) {
	t.Parallel()

	t.Run(
		"not registered user", func(t *testing.T) {
			// GIVEN
			user := &diagram.User{ID: "NA"}

			// WHEN
			got := defineBestOf(user)

			// THEN
			if got != notRegisteredBestOf {
				t.Fatalf("unexpected bestOf")
			}
		},
	)

	t.Run(
		"registered user", func(t *testing.T) {
			// GIVEN
			user := &diagram.User{ID: "foobar", IsRegistered: true}

			// WHEN
			got := defineBestOf(user)

			// THEN
			if got != registeredBestOf {
				t.Fatalf("unexpected bestOf")
			}
		},
	)
}

func Test_defineModel(t *testing.T) {
	t.Parallel()

	t.Run(
		"not registered user", func(t *testing.T) {
			// GIVEN
			user := &diagram.User{ID: "NA"}

			// WHEN
			got := defineModel(user)

			// THEN
			if got != notRegisteredModel {
				t.Fatalf("unexpected Model")
			}
		},
	)

	t.Run(
		"registered user", func(t *testing.T) {
			// GIVEN
			user := &diagram.User{ID: "foobar", IsRegistered: true}

			// WHEN
			got := defineModel(user)

			// THEN
			if got != registeredModel {
				t.Fatalf("unexpected Model")
			}
		},
	)
}
