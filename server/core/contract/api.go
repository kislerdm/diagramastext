package contract

// Request API request object.
type Request struct {
	// Prompt user's prompt describing the diagram.
	Prompt string `json:"prompt"`
}

// Response API response object.
type Response struct {
	// SVG XML-encoded SVG diagram.
	SVG string `json:"svg"`
}
