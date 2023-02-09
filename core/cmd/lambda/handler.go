package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

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
				StatusCode: http.StatusUnprocessableEntity,
			}, err
		}

		if err := validatePrompt(prompt); err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusBadRequest,
			}, err
		}

		graph, err := clientModel.Do(ctx, prompt)
		if err != nil {
			return parseClientError(err), err
		}

		svg, err := clientDiagram.Do(ctx, graph)
		if err != nil {
			return parseClientError(err), err
		}

		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Body:       string(svg.MustMarshal()),
		}, err
	}
}

func parseClientError(err error) events.APIGatewayProxyResponse {
	msg := "unknown"
	if e, ok := err.(core.Error); ok {
		if e.ServiceResponseStatusCode == http.StatusTooManyRequests {
			return events.APIGatewayProxyResponse{
				StatusCode: e.ServiceResponseStatusCode,
				Body:       "service experiences high load, please try later",
			}
		}
		switch e.Service {
		case core.ServiceOpenAI:
			msg = "could not recognise diagram description"
		case core.ServiePlantUML:
			msg = "could not generate diagram using provided description"
		}
	}
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       msg,
	}
}

const (
	promptLengthMin = 3
	promptLengthMax = 1000
)

func validatePrompt(prompt string) error {
	if len(prompt) < promptLengthMin || len(prompt) > promptLengthMax {
		return errors.New(
			"prompt length must be between " + strconv.Itoa(promptLengthMin) + " and " +
				strconv.Itoa(promptLengthMax) + " characters",
		)
	}
	return nil
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
