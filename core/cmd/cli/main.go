//go:build !unittest
// +build !unittest

package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/kislerdm/diagramastext/core/storage"

	"github.com/kislerdm/diagramastext/core"
)

func main() {
	client, err := core.NewOpenAIClient(
		core.ConfigOpenAI{
			Token:       os.Getenv("OPENAI_API_KEY"),
			MaxTokens:   768,
			Temperature: 0,
		},
	)
	if err != nil {
		log.Fatalln(err)
	}

	clientStorage, err := storage.NewClient(
		context.Background(),
		os.Getenv("NEON_HOST"),
		os.Getenv("NEON_DBNAME"),
		os.Getenv("NEON_USER"),
		os.Getenv("NEON_PASSWORD"),
	)
	if err != nil {
		log.Fatalln(err)
	}

	const prompt = `c4 diagram with kotlin backend reading postgres`
	//if err := clientStorage.WritePrompt(context.Background(), userInput); err != nil {
	//	log.Print("WritePrompt() error " + err.Error())
	//}

	graph, err := client.Do(context.Background(), prompt)
	if err != nil {
		log.Fatalln(err)
	}

	resp, _ := json.Marshal(graph)

	predictionOutput := core.ModelOutput{
		CallID: core.CallID{
			RequestID: "6863f332-1809-40ec-8bbd-878ccb11ff06",
			UserID:    "00000000-0000-0000-0000-000000000001",
		},
		Response:  string(resp),
		Timestamp: time.Now().UTC(),
	}
	if err := clientStorage.WriteModelPrediction(context.Background(), predictionOutput); err != nil {
		log.Print("WriteModelPrediction() error " + err.Error())
	}

	plantUMLClient := core.NewPlantUMLClient()
	svg, err := plantUMLClient.Do(context.Background(), graph)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(string(svg.MustMarshal()))
}
