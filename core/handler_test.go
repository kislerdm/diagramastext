package core

import (
	"encoding/json"
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

func Test_responseC4Diagram_ToJSON(t *testing.T) {
	mustMarshal := func(v []byte) []byte {
		o, _ := json.Marshal(
			map[string][]byte{
				"svg": v,
			},
		)
		return o
	}

	type fields struct {
		SVG []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   []byte
	}{
		{
			name: "happy path",
			fields: fields{
				SVG: []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"></svg>`),
			},
			want: mustMarshal([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"></svg>`)),
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				r := responseC4Diagram{
					SVG: tt.fields.SVG,
				}
				if got := r.ToJSON(); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ToJSON() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}
