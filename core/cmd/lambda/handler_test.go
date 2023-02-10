package main

import (
	"context"
	"math/rand"
	"net/http"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/kislerdm/diagramastext/core"
)

func Test_readPrompt(t *testing.T) {
	type args struct {
		req events.APIGatewayProxyRequest
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				req: events.APIGatewayProxyRequest{
					Body: `{"prompt":"foobar"}`,
				},
			},
			want:    "foobar",
			wantErr: false,
		},
		{
			name: "unhappy path: base64 encoded",
			args: args{
				req: events.APIGatewayProxyRequest{
					Body:            `eyJwcm9tcHQiOiJmb29iYXIifQ==`,
					IsBase64Encoded: true,
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "unhappy path: corrupt data",
			args: args{
				req: events.APIGatewayProxyRequest{
					Body: `{"prompt":"foobar"`,
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "unhappy path: corrupt data",
			args: args{
				req: events.APIGatewayProxyRequest{
					Body: `{"prompt":"foobar"`,
				},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := readPrompt(tt.args.req)
				if (err != nil) != tt.wantErr {
					t.Errorf("readPrompt() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("readPrompt() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_validatePrompt(t *testing.T) {
	type args struct {
		prompt string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid",
			args: args{
				prompt: "c4 diagram with Go backend reading postgres",
			},
			wantErr: false,
		},
		{
			name: "invalid: short",
			args: args{
				prompt: randomString(promptLengthMin - 1),
			},
			wantErr: true,
		},
		{
			name: "invalid: long",
			args: args{
				prompt: randomString(promptLengthMax + 1),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if err := validatePrompt(tt.args.prompt); (err != nil) != tt.wantErr {
					t.Errorf("validatePrompt() error = %v, wantErr %v", err, tt.wantErr)
				}
			},
		)
	}
}

func randomString(length int) string {
	const charset = "abcdef"
	var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	var b = make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func Test_parseClientError(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want events.APIGatewayProxyResponse
	}{
		{
			name: "too many requests",
			args: args{
				err: core.Error{
					ServiceResponseStatusCode: http.StatusTooManyRequests,
				},
			},
			want: events.APIGatewayProxyResponse{
				StatusCode: http.StatusTooManyRequests,
				Body:       "service experiences high load, please try later",
			},
		},
		{
			name: "opanAI failed to predict",
			args: args{
				err: core.Error{
					Service:                   core.ServiceOpenAI,
					ServiceResponseStatusCode: http.StatusInternalServerError,
				},
			},
			want: events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       "could not recognise diagram description",
			},
		},
		{
			name: "plantUML failed to predict",
			args: args{
				err: core.Error{
					Service:                   core.ServiePlantUML,
					ServiceResponseStatusCode: http.StatusInternalServerError,
				},
			},
			want: events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       "could not generate diagram using provided description",
			},
		},
		{
			name: "unknown",
			args: args{
				err: core.Error{},
			},
			want: events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       "unknown",
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := parseClientError(tt.args.err); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("parseClientError() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

type mockClientInputToGraph struct {
	Diagram core.DiagramGraph
	Err     error
}

func (m mockClientInputToGraph) Do(ctx context.Context, s string) (core.DiagramGraph, error) {
	_ = ctx
	_ = s
	return m.Diagram, m.Err
}

type mockClientGraphToDiagram struct {
	Resp core.ResponseDiagram
	Err  error
}

func (m mockClientGraphToDiagram) Do(ctx context.Context, graph core.DiagramGraph) (core.ResponseDiagram, error) {
	_ = ctx
	_ = graph
	return m.Resp, m.Err
}

func Test_handler(t *testing.T) {
	type fields struct {
		clientModel   core.ClientInputToGraph
		clientDiagram core.ClientGraphToDiagram
	}
	type args struct {
		ctx context.Context
		req events.APIGatewayProxyRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    events.APIGatewayProxyResponse
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				clientModel: mockClientInputToGraph{
					Diagram: core.DiagramGraph{
						Nodes: []*core.Node{{ID: "0"}},
					},
					Err: nil,
				},
				clientDiagram: mockClientGraphToDiagram{
					Resp: core.ResponseC4Diagram{SVG: "<svg></svg>"},
					Err:  nil,
				},
			},
			args: args{
				ctx: context.TODO(),
				req: events.APIGatewayProxyRequest{
					Body: `{"prompt": "` + randomString(promptLengthMin+1) + `"}`,
				},
			},
			want: events.APIGatewayProxyResponse{
				StatusCode: http.StatusOK,
				Body:       string(core.ResponseC4Diagram{SVG: "<svg></svg>"}.MustMarshal()),
			},
			wantErr: false,
		},
		{
			name: "unhappy path: faulty prompt",
			args: args{
				ctx: context.TODO(),
				req: events.APIGatewayProxyRequest{
					Body: `{"prompt":`,
				},
			},
			want: events.APIGatewayProxyResponse{
				StatusCode: http.StatusUnprocessableEntity,
				Body:       "could not recognise the prompt format",
			},
			wantErr: true,
		},
		{
			name: "unhappy path: invalid prompt",
			args: args{
				ctx: context.TODO(),
				req: events.APIGatewayProxyRequest{
					Body: `{"prompt":"` + randomString(promptLengthMin-1) + `"}`,
				},
			},
			want: events.APIGatewayProxyResponse{
				StatusCode: http.StatusBadRequest,
				Body: "prompt length must be between " + strconv.Itoa(promptLengthMin) + " and " +
					strconv.Itoa(promptLengthMax) + " characters",
			},
			wantErr: true,
		},
		{
			name: "unhappy path: openAI error",
			fields: fields{
				clientModel: mockClientInputToGraph{
					Err: core.Error{
						Service:                   core.ServiceOpenAI,
						Stage:                     core.StageResponse,
						Message:                   "foobar",
						ServiceResponseStatusCode: http.StatusTooManyRequests,
					},
				},
			},
			args: args{
				ctx: context.TODO(),
				req: events.APIGatewayProxyRequest{
					Body: `{"prompt": "` + randomString(promptLengthMin+1) + `"}`,
				},
			},
			want: events.APIGatewayProxyResponse{
				StatusCode: http.StatusTooManyRequests,
				Body:       "service experiences high load, please try later",
			},
			wantErr: true,
		},
		{
			name: "unhappy path: plantUML error",
			fields: fields{
				clientModel: mockClientInputToGraph{
					Diagram: core.DiagramGraph{
						Nodes: []*core.Node{{ID: "0"}},
					},
					Err: nil,
				},
				clientDiagram: mockClientGraphToDiagram{
					Err: core.Error{
						Service:                   core.ServiePlantUML,
						Stage:                     core.StageResponse,
						Message:                   "foobar",
						ServiceResponseStatusCode: http.StatusTooManyRequests,
					},
				},
			},
			args: args{
				ctx: context.TODO(),
				req: events.APIGatewayProxyRequest{
					Body: `{"prompt": "` + randomString(promptLengthMin+1) + `"}`,
				},
			},
			want: events.APIGatewayProxyResponse{
				StatusCode: http.StatusTooManyRequests,
				Body:       "service experiences high load, please try later",
			},
			wantErr: true,
		},
	}

	t.Parallel()

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, gotErr := handler(tt.fields.clientModel, tt.fields.clientDiagram)(tt.args.ctx, tt.args.req)
				if (gotErr != nil) != tt.wantErr {
					t.Errorf("handler execution error = %v, wantErr %v", gotErr, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("handler execution got: %v, want: %v", got, tt.want)
				}
			},
		)
	}
}
