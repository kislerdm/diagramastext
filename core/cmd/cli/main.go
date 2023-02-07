package main

import (
	"context"
	"log"

	"github.com/kislerdm/diagramastext/core"
)

func main() {
	client := core.NewPlantUMLClient()

	svg, err := client.Do(context.Background(), core.DiagramGraph{Nodes: []*core.Node{{ID: "0"}}})
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(string(svg.MustMarshal()))
}
