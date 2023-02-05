package core

import (
	"reflect"
	"testing"
)

type mockClientOpenAI struct{}

func (m mockClientOpenAI) InferModel(s string) (DiagramGraph, error) {
	return DiagramGraph{}, nil
}

type mockClientPlantUML struct{}

func (m mockClientPlantUML) GenerateDiagram(code string) ([]byte, error) {
	return nil, nil
}

func TestNewC4Handler(t *testing.T) {
	t.Run(
		"happy path", func(t *testing.T) {

			// GIVEN
			// the OpenAI and PlantUML clients

			clientOpenAI := mockClientOpenAI{}
			clientPlantUML := mockClientPlantUML{}

			// WHEN
			// initiating the handler object
			got, _ := NewC4Handler(clientOpenAI, clientPlantUML)

			// THEN
			want := handlerC4Diagram{
				clientOpenAI:   clientOpenAI,
				clientPlantUML: clientPlantUML,
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("NewC4Handler() got = %v, want %v", got, want)
			}
		},
	)
}
