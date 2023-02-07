package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"unsafe"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/kislerdm/diagramastext/core"
)

type Request struct {
	Prompt string `json:"prompt"`
}

func main() {
	clientOpenAI, err := core.NewOpenAIClient(
		core.ConfigOpenAI{
			Token:       os.Getenv("OPENAI_API_KEY"),
			MaxTokens:   mustParseInt(os.Getenv("OPENAI_MAX_TOKENS")),
			Temperature: mustParseFloat32(os.Getenv("OPENAI_MAX_TOKENS")),
		},
	)
	if err != nil {
		log.Fatalln(err)
	}

	clientPlantUML := core.NewPlantUMLClient()

	lambda.Start(
		func(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
			var r Request
			if err := json.Unmarshal(*(*[]byte)(unsafe.Pointer(&req.Body)), &r); err != nil {
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusBadRequest,
				}, err
			}

			graph, err := clientOpenAI.Do(ctx, r.Prompt)
			if err != nil {
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusInternalServerError,
					Body:       "could not recognise diagram description",
				}, err
			}

			svg, err := clientPlantUML.Do(context.Background(), graph)
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

func mustParseInt(s string) int {
	o, _ := strconv.Atoi(s)
	return o
}

func mustParseFloat32(s string) float32 {
	o, _ := strconv.ParseFloat(s, 10)
	return float32(o)
}
