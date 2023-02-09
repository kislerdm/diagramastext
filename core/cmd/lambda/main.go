package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/kislerdm/diagramastext/core"
)

func main() {
	clientOpenAI, err := core.NewOpenAIClient(
		core.ConfigOpenAI{
			Token:       os.Getenv("OPENAI_API_KEY"),
			MaxTokens:   mustParseInt(os.Getenv("OPENAI_MAX_TOKENS")),
			Temperature: mustParseFloat32(os.Getenv("OPENAI_TEMPERATURE")),
		},
	)
	if err != nil {
		log.Fatalln(err)
	}

	clientPlantUML := core.NewPlantUMLClient()

	lambda.Start(
		func(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
			prompt, err := readPrompt(req)
			if err != nil {
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusBadRequest,
				}, err
			}
			graph, err := clientOpenAI.Do(ctx, prompt)
			if err != nil {
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusInternalServerError,
					Body:       "could not recognise diagram description",
				}, err
			}

			svg, err := clientPlantUML.Do(ctx, graph)
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
		},
	)
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

func mustParseInt(s string) int {
	o, _ := strconv.Atoi(s)
	return o
}

func mustParseFloat32(s string) float32 {
	o, _ := strconv.ParseFloat(s, 10)
	return float32(o)
}
