package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	coreHandler "github.com/kislerdm/diagramastext/core/handler"
	"github.com/kislerdm/diagramastext/core/utils"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/kislerdm/diagramastext/core"
)

func main() {
	clientOpenAI, err := core.NewOpenAIClient(
		core.ConfigOpenAI{
			Token:       os.Getenv("OPENAI_API_KEY"),
			MaxTokens:   utils.MustParseInt(os.Getenv("OPENAI_MAX_TOKENS")),
			Temperature: utils.MustParseFloat32(os.Getenv("OPENAI_TEMPERATURE")),
		},
	)
	if err != nil {
		log.Fatalln(err)
	}

	clientPlantUML := core.NewPlantUMLClient()

	var corsHeaders corsHeaders
	if v := os.Getenv("CORS_HEADERS"); v != "" {
		_ = json.Unmarshal([]byte(v), &corsHeaders)
	}

	lambda.Start(handler(clientOpenAI, clientPlantUML, corsHeaders))
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

func handler(
	clientModel core.ClientInputToGraph, clientDiagram core.ClientGraphToDiagram, corsHeaders corsHeaders,
) func(
	ctx context.Context, req events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	return func(
		ctx context.Context, req events.APIGatewayProxyRequest,
	) (events.APIGatewayProxyResponse, error) {
		prompt, err := coreHandler.ReadPrompt([]byte(req.Body))
		if err != nil {
			return corsHeaders.setHeaders(
				events.APIGatewayProxyResponse{
					StatusCode: http.StatusUnprocessableEntity,
					Body:       "could not recognise the prompt format",
				},
			), err
		}

		if err := coreHandler.ValidatePrompt(prompt); err != nil {
			return corsHeaders.setHeaders(
				events.APIGatewayProxyResponse{
					StatusCode: http.StatusBadRequest,
					Body:       err.Error(),
				},
			), err
		}

		graph, err := clientModel.Do(ctx, prompt)
		if err != nil {
			return corsHeaders.setHeaders(parseClientError(err)), err
		}

		svg, err := clientDiagram.Do(ctx, graph)
		if err != nil {
			return corsHeaders.setHeaders(parseClientError(err)), err
		}

		return corsHeaders.setHeaders(
			events.APIGatewayProxyResponse{
				StatusCode: http.StatusOK,
				Body:       string(svg.MustMarshal()),
			},
		), nil
	}
}

func parseClientError(err error) events.APIGatewayProxyResponse {
	v := coreHandler.ParseClientError(err)
	return events.APIGatewayProxyResponse{
		StatusCode: v.StatusCode,
		Body:       string(v.Body),
	}
}
