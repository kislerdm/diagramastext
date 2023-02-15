//go:build !unittest
// +build !unittest

package main

import (
	"context"
	"log"
	"os"

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

	graph, err := client.Do(
		context.Background(),
		`c4 diagram with kotlin backend reading postgres`,
	)
	if err != nil {
		log.Fatalln(err)
	}

	plantUMLClient := core.NewPlantUMLClient()
	svg, err := plantUMLClient.Do(context.Background(), graph)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(string(svg.MustMarshal()))
}
