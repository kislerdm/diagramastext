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
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/kislerdm/diagramastext/core"
	coreHandler "github.com/kislerdm/diagramastext/core/handler"
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
		corsHeaders   corsHeaders
		clientStorage core.ClientStorage
	}
	type args struct {
		ctx context.Context
		req events.APIGatewayProxyRequest
	}

	expectedHandler := map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Headers": "'Content-Type,X-Amz-Date,x-api-key,Authorization,X-Api-Key,X-Amz-Security-Token'",
		"Access-Control-Allow-Methods": "'POST,OPTIONS'",
	}

	ctx := lambdacontext.NewContext(context.TODO(), &lambdacontext.LambdaContext{AwsRequestID: "foobar"})

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
				corsHeaders:   expectedHandler,
				clientStorage: core.MockClientStorage{},
			},
			args: args{
				ctx: ctx,
				req: events.APIGatewayProxyRequest{
					Body: `{"prompt": "` + randomString(coreHandler.PromptLengthMin+1) + `"}`,
				},
			},
			want: events.APIGatewayProxyResponse{
				Headers:    expectedHandler,
				StatusCode: http.StatusOK,
				Body:       string(core.ResponseC4Diagram{SVG: "<svg></svg>"}.MustMarshal()),
			},
			wantErr: false,
		},
		{
			name: "unhappy path: faulty prompt",
			fields: fields{
				corsHeaders:   expectedHandler,
				clientStorage: core.MockClientStorage{},
			},
			args: args{
				ctx: ctx,
				req: events.APIGatewayProxyRequest{
					Body: `{"prompt":`,
				},
			},
			want: events.APIGatewayProxyResponse{
				Headers:    expectedHandler,
				StatusCode: http.StatusUnprocessableEntity,
				Body:       "could not recognise the prompt format",
			},
			wantErr: true,
		},
		{
			name: "unhappy path: invalid prompt",
			fields: fields{
				corsHeaders:   expectedHandler,
				clientStorage: core.MockClientStorage{},
			},
			args: args{
				ctx: context.TODO(),
				req: events.APIGatewayProxyRequest{
					Body: `{"prompt":"` + randomString(coreHandler.PromptLengthMin-1) + `"}`,
				},
			},
			want: events.APIGatewayProxyResponse{
				Headers:    expectedHandler,
				StatusCode: http.StatusBadRequest,
				Body: "prompt length must be between " + strconv.Itoa(coreHandler.PromptLengthMin) + " and " +
					strconv.Itoa(coreHandler.PromptLengthMax) + " characters",
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
				corsHeaders:   expectedHandler,
				clientStorage: core.MockClientStorage{},
			},
			args: args{
				ctx: ctx,
				req: events.APIGatewayProxyRequest{
					Body: `{"prompt": "` + randomString(100) + `"}`,
				},
			},
			want: events.APIGatewayProxyResponse{
				Headers:    expectedHandler,
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
				corsHeaders:   expectedHandler,
				clientStorage: core.MockClientStorage{},
			},
			args: args{
				ctx: ctx,
				req: events.APIGatewayProxyRequest{
					Body: `{"prompt": "` + randomString(100) + `"}`,
				},
			},
			want: events.APIGatewayProxyResponse{
				Headers:    expectedHandler,
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
				got, gotErr := handler(
					tt.fields.clientModel, tt.fields.clientDiagram, tt.fields.corsHeaders, tt.fields.clientStorage,
				)(tt.args.ctx, tt.args.req)
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

func Test_corsHeaders_setHeaders(t *testing.T) {
	type args struct {
		resp events.APIGatewayProxyResponse
	}
	tests := []struct {
		name string
		h    corsHeaders
		args args
		want events.APIGatewayProxyResponse
	}{
		{
			name: "no cors handlers",
			h:    nil,
			args: args{
				resp: events.APIGatewayProxyResponse{},
			},
			want: events.APIGatewayProxyResponse{},
		},
		{
			name: "cors: wildcard origin and methods",
			h: corsHeaders{
				"Access-Control-Allow-Origin":  "'*'",
				"Access-Control-Allow-Methods": "'POST,OPTIONS'",
			},
			args: args{
				resp: events.APIGatewayProxyResponse{},
			},
			want: events.APIGatewayProxyResponse{
				Headers: map[string]string{
					"Access-Control-Allow-Origin":  "*",
					"Access-Control-Allow-Methods": "'POST,OPTIONS'",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := tt.h.setHeaders(tt.args.resp); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("setHeaders() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}
