package core

import "encoding/json"

// ResponseDiagram response object.
type ResponseDiagram interface {
	// ToJSON encodes the result as JSON.
	ToJSON() []byte
}

// Handler orchestrator the logic invoked upon user's request.
type Handler interface {
	// GenerateDiagram generates the diagram artifact as SVG.
	GenerateDiagram(string) (ResponseDiagram, error)
}

// DiagramGraph defines the diagram graph.
type DiagramGraph struct {
	Title  string  `json:"title"`
	Footer string  `json:"footer"`
	Nodes  []*Node `json:"nodes"`
	Links  []*Link `json:"links"`
}

// Node diagram's definition node.
type Node struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	Group      string `json:"group"`
	Technology string `json:"technology"`
	External   bool   `json:"external"`
	IsQueue    bool   `json:"is_queue"`
	IsDatabase bool   `json:"is_database"`
}

// Link diagram's definition link.
type Link struct {
	From       string `json:"from"`
	To         string `json:"to"`
	Direction  string `json:"direction"`
	Label      string `json:"label"`
	Technology string `json:"technology"`
}

// ClientOpenAI client to communicate with OpenAI.
type ClientOpenAI interface {
	// InferModel infers the model.
	InferModel(string) (DiagramGraph, error)
}

type handlerC4Diagram struct {
	clientOpenAI   ClientOpenAI
	clientPlantUML ClientPlantUML
}

type responseC4Diagram struct {
	SVG []byte `json:"svg"`
}

func (r responseC4Diagram) ToJSON() []byte {
	o, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	return o
}

func (h handlerC4Diagram) GenerateDiagram(s string) (ResponseDiagram, error) {
	graph, err := h.clientOpenAI.InferModel(s)
	if err != nil {
		return nil, err
	}

	plantUMLCode, err := diagramGraph2plantUMLCode(graph)
	if err != nil {
		return nil, err
	}

	plantUMLReqString, err := code2Path(plantUMLCode)
	if err != nil {
		return nil, err
	}

	diagram, err := h.clientPlantUML.GenerateDiagram(plantUMLReqString)
	if err != nil {
		return nil, err
	}

	return responseC4Diagram{SVG: diagram}, nil
}

// diagramGraph2plantUMLCode function to "transpile" the diagram definition graph to plantUML code as string.
func diagramGraph2plantUMLCode(graph DiagramGraph) (string, error) {
	panic("todo")
}

// NewC4Handler initialises the Handler to generate C4 diagram.
func NewC4Handler(clientOpenAI ClientOpenAI, clientPlantUML ClientPlantUML) (Handler, error) {
	return handlerC4Diagram{
		clientOpenAI:   clientOpenAI,
		clientPlantUML: clientPlantUML,
	}, nil
}
