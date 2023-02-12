package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	coreHandler "github.com/kislerdm/diagramastext/core/handler"
	"github.com/kislerdm/diagramastext/core/utils"

	"github.com/kislerdm/diagramastext/core"
)

type httpHandler struct {
	clientModel   core.ClientInputToGraph
	clientDiagram core.ClientGraphToDiagram
	reportErrorFn func(err error)
	corsHeaders   corsHeaders
}

func (h httpHandler) response(w http.ResponseWriter, body []byte, status int, err error) {
	if err != nil {
		h.reportErrorFn(err)
	}
	h.corsHeaders.setHeaders(w.Header())
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

func (h httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() { _ = r.Body.Close() }()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.response(w, []byte("could not recognise the prompt format"), http.StatusUnprocessableEntity, err)
		return
	}

	prompt, err := coreHandler.ReadPrompt(body)
	if err != nil {
		h.response(w, []byte("could not recognise the prompt format"), http.StatusUnprocessableEntity, err)
		return
	}

	if err := coreHandler.ValidatePrompt(prompt); err != nil {
		h.response(w, []byte(err.Error()), http.StatusUnprocessableEntity, err)
		return
	}

	graph, err := h.clientModel.Do(context.Background(), prompt)
	if err != nil {
		e := coreHandler.ParseClientError(err)
		h.response(w, e.Body, e.StatusCode, err)
		return
	}

	svg, err := h.clientDiagram.Do(context.Background(), graph)
	if err != nil {
		e := coreHandler.ParseClientError(err)
		h.response(w, e.Body, e.StatusCode, err)
		return
	}

	h.response(w, svg.MustMarshal(), http.StatusOK, nil)
}

func main() {
	clientOpenAI, err := core.NewOpenAIClient(
		core.ConfigOpenAI{
			Token:       os.Getenv("OPENAI_API_KEY"),
			MaxTokens:   utils.MustParseInt(os.Getenv("OPENAI_MAX_TOKENS")),
			Temperature: utils.MustParseFloat32(os.Getenv("OPENAI_TEMPERATURE")),
		},
	)
	if err != nil {
		log.Fatalln(err)
	}

	handler := httpHandler{
		clientModel:   clientOpenAI,
		clientDiagram: core.NewPlantUMLClient(),
	}

	if v := os.Getenv("CORS_HEADERS"); v != "" {
		_ = json.Unmarshal([]byte(v), &handler.corsHeaders)
	}

	port := 9000
	if v := utils.MustParseInt(os.Getenv("PORT")); v > 0 {
		port = v
	}

	if err := http.ListenAndServe(":"+strconv.Itoa(port), handler); err != nil {
		log.Println(err)
	}
}

type corsHeaders map[string]string

func (h corsHeaders) setHeaders(header http.Header) {
	for k, v := range h {
		header.Set(k, v)

		if k == "Access-Control-Allow-Origin" && (v == "" || v == "'*'") {
			header.Set(k, "*")
		}
	}
}
