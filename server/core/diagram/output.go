package diagram

import (
	"encoding/json"

	"github.com/kislerdm/diagramastext/server/core/internal/utils"
)

// Output defines the exit point's interface.
type Output interface {
	Serialize() ([]byte, error)
}

type MockOutput struct {
	V   []byte
	Err error
}

func (m MockOutput) Serialize() ([]byte, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.V, nil
}

type responseSVG struct {
	// SVG XML-encoded SVG diagram.
	SVG string `json:"svg"`
}

func (r responseSVG) Serialize() ([]byte, error) {
	return json.Marshal(r)
}

// NewResultSVG create a response object with the SVG diagram.
func NewResultSVG(v []byte) (Output, error) {
	if err := utils.ValidateSVG(v); err != nil {
		return nil, err
	}
	return &responseSVG{SVG: string(v)}, nil
}
