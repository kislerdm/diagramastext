package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/kislerdm/diagramastext/core"
)

func handler(clientModel core.ClientInputToGraph, clientDiagram core.ClientGraphToDiagram) func(
	ctx context.Context, req events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	return func(
		ctx context.Context, req events.APIGatewayProxyRequest,
	) (events.APIGatewayProxyResponse, error) {
		prompt, err := readPrompt(req)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusBadRequest,
			}, err
		}
		graph, err := clientModel.Do(ctx, prompt)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       "could not recognise diagram description",
			}, err
		}

		svg, err := clientDiagram.Do(ctx, graph)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       "could not generate diagram using provided description",
			}, err
		}

		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Body:       string(svg.MustMarshal()),
		}, err
	}
}

type request struct {
	Prompt string `json:"prompt"`
}

func readPrompt(req events.APIGatewayProxyRequest) (string, error) {
	var r request
	if err := json.Unmarshal([]byte(req.Body), &r); err != nil {
		return "", err
	}
	return r.Prompt, nil
}
