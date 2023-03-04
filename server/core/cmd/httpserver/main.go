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

	"github.com/kislerdm/diagramastext/server/core"
	"github.com/kislerdm/diagramastext/server/core/c4container"
	"github.com/kislerdm/diagramastext/server/core/configuration"
	"github.com/kislerdm/diagramastext/server/core/utils"
)

type httpHandler struct {
	clientModelInference    core.ModelInferenceClient
	diagramRenderingHandler core.DiagramRenderingHandler
	reportErrorFn           func(err error)
	corsHeaders             corsHeaders
}

func (h httpHandler) response(w http.ResponseWriter, body []byte, status int, err error) {
	if err != nil {
		h.reportErrorFn(err)
	}
	h.corsHeaders.setHeaders(w.Header())
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

type request struct {
	Prompt string `json:"prompt"`
}

type response struct {
	SVG string `json:"svg"`
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
	var input request
	if err := json.Unmarshal(body, &input); err != nil {
		h.response(w, []byte("could not recognise the prompt format"), http.StatusUnprocessableEntity, err)
		return
	}

	// FIXME: add proper sink to preserve user's requests for model fine-tuning
	log.Println(input.Prompt)

	diagram, err := h.diagramRenderingHandler(
		context.Background(), h.clientModelInference, core.Request{
			Prompt:                 input.Prompt,
			UserID:                 readUserID(r.Header),
			IsRegisteredUser:       isRegisteredUser(r.Header),
			OptOutFromSavingPrompt: isOptOutFromSavingPrompt(r.Header),
		},
	)
	if err != nil {
		h.response(w, []byte(err.Error()), http.StatusInternalServerError, err)
		return
	}

	o, err := json.Marshal(response{SVG: string(diagram)})
	if err != nil {
		h.response(w, []byte(err.Error()), http.StatusInternalServerError, err)
		return
	}

	h.response(w, o, http.StatusOK, nil)
}

func isOptOutFromSavingPrompt(headers http.Header) bool {
	// FIXME: extract registration from JWT when authN is implemented
	return false
}

func isRegisteredUser(headers http.Header) bool {
	// FIXME: extract registration from JWT when authN is implemented
	return false
}

func readUserID(headers http.Header) string {
	// FIXME: extract UserID from the headers when authN is implemented
	return "NA"
}

func main() {
	cfg, err := configuration.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalln(err)
	}

	c, err := core.NewModelInferenceClientFromConfig(cfg)
	if err != nil {
		log.Fatalln(err)
	}

	handler := httpHandler{
		clientModelInference:    c,
		diagramRenderingHandler: c4container.Handler,
		reportErrorFn:           func(err error) { log.Println(err) },
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
