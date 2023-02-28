package main

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/kislerdm/diagramastext/server"
	"github.com/kislerdm/diagramastext/server/pkg/rendering/c4container"
	"github.com/kislerdm/diagramastext/server/pkg/secretsmanager"
	"github.com/kislerdm/diagramastext/server/pkg/storage"
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
				err: server.Error{
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
				err: server.Error{
					Service:                   server.ServiceOpenAI,
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
				err: server.Error{
					Service:                   server.ServiePlantUML,
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
				err: server.Error{},
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
	Diagram server.DiagramGraph
	Err     error
}

func (m mockClientInputToGraph) Do(ctx context.Context, s string) (server.DiagramGraph, error) {
	_ = ctx
	_ = s
	return m.Diagram, m.Err
}

type mockClientGraphToDiagram struct {
	Resp server.ResponseDiagram
	Err  error
}

func (m mockClientGraphToDiagram) Do(ctx context.Context, graph server.DiagramGraph) (server.ResponseDiagram, error) {
	_ = ctx
	_ = graph
	return m.Resp, m.Err
}

func Test_handler(t *testing.T) {
	type fields struct {
		clientModel   server.ClientInputToGraph
		clientDiagram server.ClientGraphToDiagram
		corsHeaders   corsHeaders
		clientStorage storage.Client
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
					Diagram: server.DiagramGraph{
						Nodes: []*server.Node{{ID: "0"}},
					},
					Err: nil,
				},
				clientDiagram: mockClientGraphToDiagram{
					Resp: c4container.ResponseC4Diagram{SVG: "<svg></svg>"},
					Err:  nil,
				},
				corsHeaders:   expectedHandler,
				clientStorage: storage.MockClientStorage{},
			},
			args: args{
				ctx: ctx,
				req: events.APIGatewayProxyRequest{
					Body: `{"prompt": "` + randomString(server.PromptLengthMin+1) + `"}`,
				},
			},
			want: events.APIGatewayProxyResponse{
				Headers:    expectedHandler,
				StatusCode: http.StatusOK,
				Body:       string(c4container.ResponseC4Diagram{SVG: "<svg></svg>"}.MustMarshal()),
			},
			wantErr: false,
		},
		{
			name: "unhappy path: faulty prompt",
			fields: fields{
				corsHeaders:   expectedHandler,
				clientStorage: storage.MockClientStorage{},
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
				clientStorage: storage.MockClientStorage{},
			},
			args: args{
				ctx: context.TODO(),
				req: events.APIGatewayProxyRequest{
					Body: `{"prompt":"` + randomString(server.PromptLengthMin-1) + `"}`,
				},
			},
			want: events.APIGatewayProxyResponse{
				Headers:    expectedHandler,
				StatusCode: http.StatusBadRequest,
				Body: "prompt length must be between " + strconv.Itoa(server.PromptLengthMin) + " and " +
					strconv.Itoa(server.PromptLengthMax) + " characters",
			},
			wantErr: true,
		},
		{
			name: "unhappy path: openAI error",
			fields: fields{
				clientModel: mockClientInputToGraph{
					Err: server.Error{
						Service:                   server.ServiceOpenAI,
						Stage:                     server.StageResponse,
						Message:                   "foobar",
						ServiceResponseStatusCode: http.StatusTooManyRequests,
					},
				},
				corsHeaders:   expectedHandler,
				clientStorage: storage.MockClientStorage{},
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
					Diagram: server.DiagramGraph{
						Nodes: []*server.Node{{ID: "0"}},
					},
					Err: nil,
				},
				clientDiagram: mockClientGraphToDiagram{
					Err: server.Error{
						Service:                   server.ServiePlantUML,
						Stage:                     server.StageResponse,
						Message:                   "foobar",
						ServiceResponseStatusCode: http.StatusTooManyRequests,
					},
				},
				corsHeaders:   expectedHandler,
				clientStorage: storage.MockClientStorage{},
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

type mockSecretManagerClient struct {
	v []byte
}

func (m mockSecretManagerClient) ReadLatestSecret(ctx context.Context, uri string, output interface{}) error {
	if m.v == nil {
		return errors.New("no secret found")
	}
	return json.Unmarshal(m.v, output)
}

func mustMarshal(v interface{}) []byte {
	o, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return o
}

func Test_configureInterfaceClients(t *testing.T) {
	type args struct {
		ctx       context.Context
		client    secretsmanager.Client
		secretARN string
	}
	tests := []struct {
		name    string
		args    args
		envVars map[string]string
		wantErr bool
	}{
		{
			name:    "happy path",
			envVars: map[string]string{},
			args: args{
				ctx: context.TODO(),
				client: mockSecretManagerClient{
					v: mustMarshal(
						secret{
							OpenAiAPIKey: "sk-foobar",
							DBHost:       "mock",
							DBName:       "dbname",
							DBUser:       "user",
							DBPassword:   "password",
						},
					),
				},
				secretARN: "arn:aws:secretsmanager:us-east-2:027889758114:secret:foo/bar/server/lambda-C335bP",
			},
			wantErr: false,
		},
		{
			name: "happy path: environment variables",
			envVars: map[string]string{
				"OPENAI_API_KEY": "sk-foobar",
				"DB_HOST":        "mock",
				"DB_DBNAME":      "dbname",
				"DB_USER":        "user",
				"DB_PASSWORD":    "password",
			},
			args: args{
				ctx:       context.TODO(),
				client:    mockSecretManagerClient{},
				secretARN: "arn:aws:secretsmanager:us-east-2:027889758114:secret:not-exists-C335bP",
			},
			wantErr: false,
		},
		{
			name: "unhappy path: no openAI token found",
			args: args{
				ctx: context.TODO(),
				client: mockSecretManagerClient{
					v: mustMarshal(
						secret{
							DBHost:     "mock",
							DBName:     "dbname",
							DBUser:     "user",
							DBPassword: "password",
						},
					),
				},
				secretARN: "arn:aws:secretsmanager:us-east-2:027889758114:secret:no-token-C335bP",
			},
			wantErr: true,
		},
		{
			name: "unhappy path: cannot connect to db",
			args: args{
				ctx: context.TODO(),
				client: mockSecretManagerClient{
					v: mustMarshal(
						secret{
							OpenAiAPIKey: "sk-foobar",
							DBName:       "dbname",
							DBUser:       "user",
							DBPassword:   "password",
						},
					),
				},
				secretARN: "arn:aws:secretsmanager:us-east-2:027889758114:secret:no-db-host-C335bP",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				for k, v := range tt.envVars {
					t.Setenv(k, v)
				}
				_, _, err := configureInterfaceClients(
					tt.args.ctx, tt.args.client, tt.args.secretARN,
				)
				if (err != nil) != tt.wantErr {
					t.Errorf("configureInterfaceClients() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			},
		)
	}
}
