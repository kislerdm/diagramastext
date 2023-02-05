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
	// FIXME(?): add svg validation.
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

	diagram, err := h.clientPlantUML.GenerateDiagram(graph)
	if err != nil {
		return nil, err
	}

	return responseC4Diagram{SVG: diagram}, nil
}

// NewC4Handler initialises the Handler to generate C4 diagram.
func NewC4Handler(clientOpenAI ClientOpenAI, clientPlantUML ClientPlantUML) (Handler, error) {
	return handlerC4Diagram{
		clientOpenAI:   clientOpenAI,
		clientPlantUML: clientPlantUML,
	}, nil
}
