package adapter

import (
	"encoding/json"

	"github.com/kislerdm/diagramastext/server/core/port"
	"github.com/kislerdm/diagramastext/server/core/utils"
)

type responseSVG struct {
	// SVG XML-encoded SVG diagram.
	SVG string `json:"svg"`
}

func (r responseSVG) Serialize() ([]byte, error) {
	return json.Marshal(r.SVG)
}

// NewResultSVG create a response object with the SVG diagram.
func NewResultSVG(v []byte) (port.Output, error) {
	if err := utils.ValidateSVG(v); err != nil {
		return nil, err
	}
	return &responseSVG{SVG: string(v)}, nil
}
