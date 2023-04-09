package c4container

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/kislerdm/diagramastext/server/core/diagram"
	diagramErrors "github.com/kislerdm/diagramastext/server/core/errors"
)

const placeholderUserID = "00000000-0000-0000-0000-000000000000"

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
		wantErr error
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
					ID: placeholderUserID,
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
			wantErr: nil,
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
			wantErr: errors.New("foobar"),
		},
		{
			name: "unhappy path: inference error",
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
					ID: placeholderUserID,
				},
			},
			want:    nil,
			wantErr: errors.New("diagram/c4container/c4container.go:83: foobar"),
		},
		{
			name: "unhappy path: failed to predict",
			args: args{
				clientModelInference: diagram.MockModelInference{
					V: []byte(`{"error":"foobar"}`),
				},
				clientRepositoryPrediction: diagram.MockRepositoryPrediction{},
				httpClient:                 diagram.MockHTTPClient{},
			},
			input: diagram.MockInput{
				Prompt: "foobar",
				User: &diagram.User{
					ID: placeholderUserID,
				},
			},
			want:    nil,
			wantErr: diagramErrors.NewPredictionError([]byte(`{"error":"foobar"}`)),
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
					ID: placeholderUserID,
				},
			},
			want:    nil,
			wantErr: errors.New("unexpected end of JSON input"),
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
					ID: placeholderUserID,
				},
			},
			want:    nil,
			wantErr: errors.New("diagram/c4container/plantuml.go:41: foobar"),
		},
	}

	t.Parallel()

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c, err := NewC4ContainersHTTPHandler(
					tt.args.clientModelInference, tt.args.clientRepositoryPrediction, tt.args.httpClient,
				)
				if err != nil {
					t.Fatal(err)
				}

				got, err := c(context.TODO(), tt.input)
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("NewC4ContainersHTTPHandler() got = %v, want %v", got, tt.want)
				}

				var expectedError bool
				switch err.(type) {
				case nil:
					expectedError = tt.wantErr == nil
				case *json.SyntaxError:
					expectedError = tt.wantErr != nil && err.Error() == tt.wantErr.Error()
				case *diagramErrors.Error:
					expectedError = tt.wantErr != nil && diagramErrors.IsError(err, tt.wantErr.Error())
				default:
					expectedError = reflect.DeepEqual(err, tt.wantErr)
				}
				if !expectedError {
					t.Errorf("NewC4ContainersHTTPHandler() error = %v, wantErr %v", err, tt.wantErr)
					return
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
			c, err := NewC4ContainersHTTPHandler(clientModelInference, clientRepositoryPrediction, httpClient)

			// THEN
			if c != nil {
				t.Fatalf("unexpected client")
			}

			if err == nil || err.Error() !=
				"diagram/c4container/c4container.go:62: model inference client must be provided" {
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
			c, err := NewC4ContainersHTTPHandler(clientModelInference, clientRepositoryPrediction, httpClient)

			// THEN
			if c != nil {
				t.Fatalf("unexpected client")
			}

			if err == nil || err.Error() != "diagram/c4container/c4container.go:65: http client must be provided" {
				t.Fatalf("unexpected error")
			}
		},
	)
}

func Test_defineModel(t *testing.T) {
	t.Parallel()

	t.Run(
		"not registered user", func(t *testing.T) {
			// GIVEN
			user := &diagram.User{ID: placeholderUserID}

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

func Test_UnmarshalGraph(t *testing.T) {
	t.Parallel()

	t.Run(
		"default legend behaviour", func(t *testing.T) {
			// GIVEN
			graphPrediction := []byte(`{"nodes":[{"id":"0"}]}`)
			want := c4ContainersGraph{
				Containers: []*container{{ID: "0"}},
				WithLegend: true,
			}

			// WHEN
			var got c4ContainersGraph
			err := json.Unmarshal(graphPrediction, &got)

			// THEN
			if err != nil {
				t.Error("unexpected error")
				return
			}

			if !reflect.DeepEqual(got, want) {
				t.Errorf("got: %+v, want: %+v", got, want)
				return
			}
		},
	)

	t.Run(
		"legend is off explicitly", func(t *testing.T) {
			// GIVEN
			graphPrediction := []byte(`{"nodes":[{"id":"0"}],"legend":false}`)
			want := c4ContainersGraph{
				Containers: []*container{{ID: "0"}},
				WithLegend: false,
			}

			// WHEN
			var got c4ContainersGraph
			err := json.Unmarshal(graphPrediction, &got)

			// THEN
			if err != nil {
				t.Error("unexpected error")
				return
			}

			if !reflect.DeepEqual(got, want) {
				t.Errorf("got: %+v, want: %+v", got, want)
				return
			}
		},
	)

	t.Run(
		"legend is on explicitly", func(t *testing.T) {
			// GIVEN
			graphPrediction := []byte(`{"nodes":[{"id":"0"}],"legend":true}`)
			want := c4ContainersGraph{
				Containers: []*container{{ID: "0"}},
				WithLegend: true,
			}

			// WHEN
			var got c4ContainersGraph
			err := json.Unmarshal(graphPrediction, &got)

			// THEN
			if err != nil {
				t.Error("unexpected error")
				return
			}

			if !reflect.DeepEqual(got, want) {
				t.Errorf("got: %+v, want: %+v", got, want)
				return
			}
		},
	)
}

type mockRepositoryPrediction struct {
	InputPromptWritten     uint8
	ModelPredictionWritten uint8
	SuccessFlagWritten     uint8
}

func (m *mockRepositoryPrediction) WriteInputPrompt(_ context.Context, _, _, _ string) error {
	m.InputPromptWritten++
	return nil
}

func (m *mockRepositoryPrediction) WriteModelResult(_ context.Context, _, _, _, _, _ string, _, _ uint16) error {
	m.ModelPredictionWritten++
	return nil
}

func (m *mockRepositoryPrediction) WriteSuccessFlag(_ context.Context, _, _, _ string) error {
	m.SuccessFlagWritten++
	return nil
}

func (m *mockRepositoryPrediction) Close(_ context.Context) error {
	return nil
}

func TestC4ContainerHandlerRepositoryPredictionPersistence(t *testing.T) {
	t.Parallel()

	t.Run(
		"happy path", func(t *testing.T) {
			// GIVEN
			repositoryPredictionClient := &mockRepositoryPrediction{}

			modelInferenceClient := diagram.MockModelInference{
				V: []byte(`{"nodes":[{"id":"0"}]}`),
			}

			httpClient := diagram.MockHTTPClient{
				V: &http.Response{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(
						strings.NewReader(
							`<?xml version="1.0" encoding="us-ascii" standalone="no"?>
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10" width="100%" height="100%">
<defs></defs><g><g id="elem_n0"><rect fill="#438DD5" width="52.5938" rx="2.5" ry="2.5"></rect></g></g></svg>`,
						),
					),
				},
			}

			userInput := diagram.MockInput{
				Prompt:    "foobar",
				RequestID: "1410904f-f646-488f-ae08-cc341dfb321c",
				User: &diagram.User{
					ID: placeholderUserID,
				},
			}

			handler, err := NewC4ContainersHTTPHandler(modelInferenceClient, repositoryPredictionClient, httpClient)

			if err != nil {
				t.Fatalf("unexpected init error")
			}

			// WHEN
			if _, err := handler(context.TODO(), userInput); err != nil {
				t.Errorf("unexpected handling error: %+v", err)
			}

			// THEN
			if repositoryPredictionClient.InputPromptWritten != 1 {
				t.Errorf(
					"prompt persisted unexpectedly: got = %v\nwant = %v",
					repositoryPredictionClient.InputPromptWritten, 1,
				)
			}
			if repositoryPredictionClient.ModelPredictionWritten != 1 {
				t.Errorf(
					"model's prediction persisted unexpectedly: got = %v\nwant = %v",
					repositoryPredictionClient.ModelPredictionWritten, 1,
				)
			}
			if repositoryPredictionClient.SuccessFlagWritten != 1 {
				t.Errorf(
					"successful generation flag persisted unexpectedly: got = %v\nwant = %v",
					repositoryPredictionClient.SuccessFlagWritten, 1,
				)
			}
		},
	)

	t.Run(
		"unhappy path: model inference failed", func(t *testing.T) {
			// GIVEN
			repositoryPredictionClient := &mockRepositoryPrediction{}

			const errMsg = "foobar"
			modelInferenceClient := diagram.MockModelInference{
				Err: errors.New(errMsg),
			}

			httpClient := diagram.MockHTTPClient{}

			userInput := diagram.MockInput{
				Prompt:    "foobar",
				RequestID: "1410904f-f646-488f-ae08-cc341dfb321c",
				User: &diagram.User{
					ID: placeholderUserID,
				},
			}

			handler, err := NewC4ContainersHTTPHandler(modelInferenceClient, repositoryPredictionClient, httpClient)

			if err != nil {
				t.Fatalf("unexpected init error")
			}

			// WHEN
			if _, err := handler(context.TODO(), userInput); !diagramErrors.IsError(
				err, err.Error(),
			) || !strings.HasSuffix(err.Error(), errMsg) {
				t.Error("unexpected handling error")
			}

			// THEN
			if repositoryPredictionClient.InputPromptWritten != 1 {
				t.Errorf(
					"prompt persisted unexpectedly: got = %v\nwant = %v",
					repositoryPredictionClient.InputPromptWritten, 1,
				)
			}
			if repositoryPredictionClient.ModelPredictionWritten != 0 {
				t.Errorf(
					"model's prediction persisted unexpectedly: got = %v\nwant = %v",
					repositoryPredictionClient.ModelPredictionWritten, 0,
				)
			}
			if repositoryPredictionClient.SuccessFlagWritten != 0 {
				t.Errorf(
					"successful generation flag persisted unexpectedly: got = %v\nwant = %v",
					repositoryPredictionClient.SuccessFlagWritten, 0,
				)
			}
		},
	)

	t.Run(
		"unhappy path: diagram generation failed", func(t *testing.T) {
			// GIVEN
			repositoryPredictionClient := &mockRepositoryPrediction{}

			modelInferenceClient := diagram.MockModelInference{
				V: []byte(`{"nodes":[{"id":"0"}]}`),
			}

			const errMsg = "foobar"
			httpClient := diagram.MockHTTPClient{
				Err: errors.New(errMsg),
			}

			userInput := diagram.MockInput{
				Prompt:    "foobar",
				RequestID: "1410904f-f646-488f-ae08-cc341dfb321c",
				User: &diagram.User{
					ID: placeholderUserID,
				},
			}

			handler, err := NewC4ContainersHTTPHandler(modelInferenceClient, repositoryPredictionClient, httpClient)

			if err != nil {
				t.Fatalf("unexpected init error")
			}

			// WHEN
			if _, err := handler(context.TODO(), userInput); !diagramErrors.IsError(
				err, err.Error(),
			) || !strings.HasSuffix(err.Error(), errMsg) {
				t.Error("unexpected handling error")
			}

			// THEN
			if repositoryPredictionClient.InputPromptWritten != 1 {
				t.Errorf(
					"prompt persisted unexpectedly: got = %v\nwant = %v",
					repositoryPredictionClient.InputPromptWritten, 1,
				)
			}
			if repositoryPredictionClient.ModelPredictionWritten != 1 {
				t.Errorf(
					"model's prediction persisted unexpectedly: got = %v\nwant = %v",
					repositoryPredictionClient.ModelPredictionWritten, 1,
				)
			}
			if repositoryPredictionClient.SuccessFlagWritten != 0 {
				t.Errorf(
					"successful generation flag persisted unexpectedly: got = %v\nwant = %v",
					repositoryPredictionClient.SuccessFlagWritten, 0,
				)
			}
		},
	)
}
