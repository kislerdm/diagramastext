package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

var	inferURL = os.Getenv("INFER_URL")

// Client represents a client for communicating with the OpenAI API.
type OpenAIClient struct {
	Token     string
	Organization string
	ModelID  string
	httpClient *http.Client
}

// NewClient returns a new instance of a Client.
func NewClient(organization, modelID string) *OpenAIClient {
	token := os.Getenv("OPENAI_API_KEY")
	if token == "" {
		log.Fatalf("OPENAI_API_KEY environment variable is not set")
	}
	return &OpenAIClient{
		Token: token,
		Organization: organization,
		ModelID: modelID,
		httpClient: http.DefaultClient,
	}
}

type DiagramGraph struct {
	Title  string
	Footer string
	Nodes  []*Node
	Links  []*Link
}

type Node struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	Group      string `json:"group"`
	Technology string `json:"technology"`
	External   bool   `json:"external"`
	IsQueue    bool   `json:"is_queue"`
	IsDatabase bool   `json:"is_database"`
}

type Link struct {
	From       string `json:"from"`
	To         string `json:"to"`
	Direction  string `json:"direction"`
	Label      string `json:"label"`
	Technology string `json:"technology"`
}


// InferModel sends the provided prompt to the OpenAI API and returns the response.
func (c *OpenAIClient) InferModel(userPrompt string) (DiagramGraph, error) {

	prefixPrompt := os.Getenv("PREFIX_PROMPT")
	description := fmt.Sprintf("description := %v\n", userPrompt)
	finalPrompt := prefixPrompt + description + "\n\n// Create a DiagramGraph object from the description above"

	reqBody := map[string]interface{}{
		"model": c.ModelID,
		"prompt": finalPrompt,
		"temperature": 0,
  		"max_tokens": 256,
		"top_p": 1,
		"frequency_penalty": 0,
		"presence_penalty": 0,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return DiagramGraph{}, err
	}

	req, err := http.NewRequest("POST", inferURL, bytes.NewBuffer(body))
	if err != nil {
		return DiagramGraph{}, err
	}

	req.Header.Add("Authorization", "Bearer "+c.Token)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", "OpenAI Go Client")
	req.Header.Add("Organization", c.Organization)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return DiagramGraph{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return DiagramGraph{}, fmt.Errorf("failed to infer model, got %s", resp.Status)
	}

	var diagramGraph DiagramGraph
	if err := json.NewDecoder(resp.Body).Decode(&diagramGraph); err != nil {
		return DiagramGraph{}, err
	}

	return diagramGraph, nil
}
