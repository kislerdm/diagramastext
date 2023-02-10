package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"

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

	var corsHeaders corsHeaders
	if v := os.Getenv("CORS_HEADERS"); v != "" {
		_ = json.Unmarshal([]byte(v), &corsHeaders)
	}

	lambda.Start(handler(clientOpenAI, clientPlantUML, corsHeaders))
}

func mustParseInt(s string) int {
	o, _ := strconv.Atoi(s)
	return o
}

func mustParseFloat32(s string) float32 {
	o, _ := strconv.ParseFloat(s, 10)
	return float32(o)
}
