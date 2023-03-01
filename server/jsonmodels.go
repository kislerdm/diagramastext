package server

// Request invocation request object.
type Request struct {
	Prompt string `json:"prompt"`
}

// ResponseSVG response with SVG encoded diagram.
type ResponseSVG struct {
	SVG string `json:"svg"`
}
