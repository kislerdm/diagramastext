//go:build !unittest
// +build !unittest

package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	coreHandler "github.com/kislerdm/diagramastext/server/handler"
	"github.com/kislerdm/diagramastext/server/pkg/core"
	"github.com/kislerdm/diagramastext/server/pkg/rendering/plantuml"
	"github.com/kislerdm/diagramastext/server/utils"

	"github.com/kislerdm/diagramastext/server"
)

type httpHandler struct {
	clientModel   server.ClientInputToGraph
	clientDiagram server.ClientGraphToDiagram
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
	if r.Method == http.MethodOptions {
		h.response(w, nil, http.StatusOK, nil)
		return
	}

	defer func() { _ = r.Body.Close() }()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.response(w, []byte("could not recognise the prompt format"), http.StatusUnprocessableEntity, err)
		return
	}

	// FIXME: add proper sink to preserve user's requests for model fine-tuning
	log.Println(string(body))

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
		core.WithSinkFn(
			// FIXME: add proper sink to preserve user's requests for model fine-tuning
			func(s string) {
				log.Println(s)
			},
		),
	)
	if err != nil {
		log.Fatalln(err)
	}

	handler := httpHandler{
		clientModel:   clientOpenAI,
		clientDiagram: plantuml.NewClient(),
		reportErrorFn: func(err error) { log.Println(err) },
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
