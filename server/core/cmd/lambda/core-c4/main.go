package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/kislerdm/diagramastext/server/core"
	"github.com/kislerdm/diagramastext/server/core/c4container"
	"github.com/kislerdm/diagramastext/server/core/configuration"
	errs "github.com/kislerdm/diagramastext/server/core/errors"
)

func main() {
	cfg, err := configuration.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	clientModelInference, err := core.NewModelInferenceClientFromConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}

	var corsHeaders corsHeaders
	if v := os.Getenv("CORS_HEADERS"); v != "" {
		if err := json.Unmarshal([]byte(v), &corsHeaders); err != nil {
			log.Fatal(err)
		}
	}

	lambda.Start(handler(c4container.Handler, clientModelInference, corsHeaders))
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

type request struct {
	Prompt string `json:"prompt"`
}

type response struct {
	SVG string `json:"svg"`
}

func handler(handler core.DiagramRenderingHandler, client core.ModelInferenceClient, corsHeaders corsHeaders) func(
	ctx context.Context, req events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	return func(
		ctx context.Context, req events.APIGatewayProxyRequest,
	) (events.APIGatewayProxyResponse, error) {
		var input request
		if err := json.Unmarshal([]byte(req.Body), &input); err != nil {
			return corsHeaders.setHeaders(
				events.APIGatewayProxyResponse{
					StatusCode: http.StatusUnprocessableEntity,
					Body:       "could not recognise the prompt format",
				},
			), err
		}

		inquiry := core.Request{
			Prompt:                 input.Prompt,
			UserID:                 readUserID(req.Headers),
			IsRegisteredUser:       isRegisteredUser(req.Headers),
			OptOutFromSavingPrompt: isOptOutFromSavingPrompt(req.Headers),
		}

		diagram, err := handler(ctx, client, inquiry)
		if err != nil {
			return corsHeaders.setHeaders(parseClientError(err)), err
		}

		output, _ := json.Marshal(response{SVG: string(diagram)})

		return corsHeaders.setHeaders(
			events.APIGatewayProxyResponse{
				StatusCode: http.StatusOK,
				Body:       string(output),
			},
		), nil
	}
}

func isOptOutFromSavingPrompt(headers map[string]string) bool {
	// FIXME: extract registration from JWT when authN is implemented
	return false
}

func isRegisteredUser(headers map[string]string) bool {
	// FIXME: extract registration from JWT when authN is implemented
	return false
}

func readUserID(headers map[string]string) string {
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
