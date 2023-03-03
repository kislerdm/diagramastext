package main

import (
	"context"
	"encoding/json"
	"math/rand"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/kislerdm/diagramastext/server"
	errs "github.com/kislerdm/diagramastext/server/errors"
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

type mockHandlerC4Diagram struct {
	svg []byte
	err error
}

func (m mockHandlerC4Diagram) TextToDiagram(ctx context.Context, req server.Request) ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.svg, nil
}

func Test_handler(t *testing.T) {
	type fields struct {
		client      server.Client
		corsHeaders corsHeaders
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
				client: mockHandlerC4Diagram{
					svg: []byte(`<?xml version="1.0" encoding="us-ascii" standalone="no"?><svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink"></svg>`),
				},
				corsHeaders: expectedHandler,
			},
			args: args{
				ctx: ctx,
				req: events.APIGatewayProxyRequest{
					Body: `{"prompt": "` + randomString(10) + `"}`,
				},
			},
			want: events.APIGatewayProxyResponse{
				Headers:    expectedHandler,
				StatusCode: http.StatusOK,
				Body: string(
					mustMarshal(
						response{
							SVG: `<?xml version="1.0" encoding="us-ascii" standalone="no"?><svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink"></svg>`,
						},
					),
				),
			},
			wantErr: false,
		},
		{
			name: "unhappy path: faulty prompt",
			fields: fields{
				corsHeaders: expectedHandler,
				client:      mockHandlerC4Diagram{},
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
			name: "unhappy path: high RPS",
			fields: fields{
				corsHeaders: expectedHandler,
				client: mockHandlerC4Diagram{
					err: errs.Error{
						Service:                   errs.ServiceOpenAI,
						Message:                   "too many requests",
						ServiceResponseStatusCode: http.StatusTooManyRequests,
					},
				},
			},
			args: args{
				ctx: lambdacontext.NewContext(context.TODO(), &lambdacontext.LambdaContext{AwsRequestID: "foobar"}),
				req: events.APIGatewayProxyRequest{
					Body: `{"prompt":"` + randomString(50) + `"}`,
				},
			},
			want: events.APIGatewayProxyResponse{
				Headers:    expectedHandler,
				StatusCode: http.StatusTooManyRequests,
				Body: errs.Error{
					Service:                   errs.ServiceOpenAI,
					Message:                   "too many requests",
					ServiceResponseStatusCode: http.StatusTooManyRequests,
				}.Error(),
			},
			wantErr: true,
		},
		{
			name: "unhappy path: handler error",
			fields: fields{
				corsHeaders: expectedHandler,
				client: mockHandlerC4Diagram{
					err: errs.Error{
						Service: errs.ServiceOpenAI,
						Message: "foobar",
					},
				},
			},
			args: args{
				ctx: lambdacontext.NewContext(context.TODO(), &lambdacontext.LambdaContext{AwsRequestID: "foobar"}),
				req: events.APIGatewayProxyRequest{
					Body: `{"prompt":"` + randomString(50) + `"}`,
				},
			},
			want: events.APIGatewayProxyResponse{
				Headers:    expectedHandler,
				StatusCode: http.StatusInternalServerError,
				Body: errs.Error{
					Service: errs.ServiceOpenAI,
					Message: "foobar",
				}.Error(),
			},
			wantErr: true,
		},
	}

	t.Parallel()

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, gotErr := handler(tt.fields.client, tt.fields.corsHeaders)(tt.args.ctx, tt.args.req)
				if (gotErr != nil) != tt.wantErr {
					t.Errorf("sdk execution error = %v, wantErr %v", gotErr, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("sdk execution got: %v, want: %v", got, tt.want)
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

func mustMarshal(v interface{}) []byte {
	o, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return o
}
