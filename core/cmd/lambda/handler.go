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
		prompt, err := readPrompt(req)
		if err != nil {
			return corsHeaders.setHeaders(
				events.APIGatewayProxyResponse{
					StatusCode: http.StatusUnprocessableEntity,
					Body:       "could not recognise the prompt format",
				},
			), err
		}

		if err := validatePrompt(prompt); err != nil {
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
		), err
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
	promptLengthMax = 768
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
