package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/kislerdm/diagramastext/server"
	errs "github.com/kislerdm/diagramastext/server/errors"
)

func main() {
	client, err := server.NewC4DiagramHandler(context.Background(), os.Getenv("ACCESS_CREDENTIALS_ARN"))
	if err != nil {
		log.Fatal(err)
	}

	defer func() { _ = client.Stop(context.Background()) }()

	var corsHeaders corsHeaders
	if v := os.Getenv("CORS_HEADERS"); v != "" {
		if err := json.Unmarshal([]byte(v), &corsHeaders); err != nil {
			log.Fatal(err)
		}
	}

	lambda.Start(handler(client, corsHeaders))
}

type corsHeaders map[string]string

func (h corsHeaders) setHeaders(resp events.APIGatewayProxyResponse) events.APIGatewayProxyResponse {
	if h == nil {
		return resp
	}

	if resp.Headers == nil {
		resp.Headers = map[string]string{}
	}

	for k, v := range h {
		resp.Headers[k] = v

		if k == "Access-Control-Allow-Origin" && (v == "" || v == "'*'") {
			resp.Headers[k] = "*"
		}
	}

	return resp
}

func handler(client server.Handler, corsHeaders corsHeaders) func(
	ctx context.Context, req events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	return func(
		ctx context.Context, req events.APIGatewayProxyRequest,
	) (events.APIGatewayProxyResponse, error) {
		var input server.Request
		if err := json.Unmarshal([]byte(req.Body), &input); err != nil {
			return corsHeaders.setHeaders(
				events.APIGatewayProxyResponse{
					StatusCode: http.StatusUnprocessableEntity,
					Body:       "could not recognise the prompt format",
				},
			), err
		}

		callID := server.CallID{
			RequestID: readRequestID(ctx),
			UserID:    readUserID(req.Headers),
		}

		diagram, err := client.GenerateSVG(ctx, input.Prompt, callID)
		if err != nil {
			return corsHeaders.setHeaders(parseClientError(err)), err
		}

		output, _ := json.Marshal(server.ResponseSVG{SVG: string(diagram)})

		return corsHeaders.setHeaders(
			events.APIGatewayProxyResponse{
				StatusCode: http.StatusOK,
				Body:       string(output),
			},
		), nil
	}
}

func readRequestID(ctx context.Context) string {
	c, _ := lambdacontext.FromContext(ctx)
	return c.AwsRequestID
}

func readUserID(h map[string]string) string {
	// FIXME: extract UserID from the headers when authN is implemented
	return "NA"
}

func parseClientError(err error) events.APIGatewayProxyResponse {
	code := http.StatusInternalServerError

	if e, ok := err.(errs.Error); ok && e.ServiceResponseStatusCode != 0 {
		code = e.ServiceResponseStatusCode
	}

	return events.APIGatewayProxyResponse{
		StatusCode: code,
		Body:       err.Error(),
	}
}
