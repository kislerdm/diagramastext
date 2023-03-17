//go:build !unittest
// +build !unittest

package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/kislerdm/diagramastext/server/core/adapter"
	"github.com/kislerdm/diagramastext/server/core/config"
	"github.com/kislerdm/diagramastext/server/core/domain/c4container"
	"github.com/kislerdm/diagramastext/server/core/pkg/gcpsecretsmanager"
	"github.com/kislerdm/diagramastext/server/core/pkg/postgres"
	"github.com/kislerdm/diagramastext/server/core/port"
	"github.com/kislerdm/diagramastext/server/core/utils"
)

var (
	secretsmanagerClient port.RepositorySecretsVault
	modelInferenceClient port.ModelInference
	postgresClient       port.RepositoryPrediction
)

func init() {
	var err error
	secretsmanagerClient, err = gcpsecretsmanager.NewSecretmanager(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	cfg := config.LoadDefaultConfig(context.Background(), secretsmanagerClient)

	modelInferenceClient, err = adapter.NewOpenAIClient(
		adapter.ConfigOpenAI{
			Token:     cfg.ModelInferenceConfig.Token,
			MaxTokens: cfg.ModelInferenceConfig.MaxTokens,
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	postgresClient, err = postgres.NewPostgresClient(
		context.Background(), postgres.Config{
			DBHost:          cfg.RepositoryPredictionConfig.DBHost,
			DBName:          cfg.RepositoryPredictionConfig.DBName,
			DBUser:          cfg.RepositoryPredictionConfig.DBUser,
			DBPassword:      cfg.RepositoryPredictionConfig.DBPassword,
			TablePrompt:     cfg.RepositoryPredictionConfig.TablePrompt,
			TablePrediction: cfg.RepositoryPredictionConfig.TablePrediction,
		},
	)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	defer func() { _ = postgresClient.Close(context.Background()) }()

	var l = log.New(os.Stderr, "", log.Lmicroseconds|log.LUTC|log.Lshortfile)

	handler := httpHandler{
		diagramRenderingHandler: map[string]port.DiagramHandler{
			"c4": c4container.NewC4ContainersHandler(modelInferenceClient, postgresClient, nil),
		},
		reportErrorFn: func(err error) { l.Println(err) },
	}

	if v := os.Getenv("CORS_HEADERS"); v != "" {
		if err := json.Unmarshal([]byte(v), &handler.corsHeaders); err != nil {
			log.Fatal(err)
		}
	}

	portServe := 9000
	if v := utils.MustParseInt(os.Getenv("PORT")); v > 0 {
		portServe = v
	}

	if err := http.ListenAndServe(":"+strconv.Itoa(portServe), handler); err != nil {
		log.Println(err)
	}
}

type httpHandler struct {
	diagramRenderingHandler map[string]port.DiagramHandler
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

func (h httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch p := r.URL.Path; p {
	case "status":
		if r.Method == http.MethodGet {
			h.response(w, nil, http.StatusOK, nil)
			return
		}
		h.response(w, nil, http.StatusMethodNotAllowed, nil)
		return
	default:
		renderingHandler, ok := h.diagramRenderingHandler[p]
		if !ok {
			h.response(w, nil, http.StatusNotFound, nil)
			return
		}

		switch r.Method {
		case http.MethodOptions:
			h.response(w, nil, http.StatusOK, nil)
			return
		case http.MethodPost:
			defer func() { _ = r.Body.Close() }()
			input, err := adapter.NewInputDriverHTTP(r.Body, r.Header)
			if err != nil {
				h.response(w, []byte("could not recognise the request"), http.StatusUnprocessableEntity, err)
				return
			}

			diagram, err := renderingHandler(context.Background(), input)
			if err != nil {
				h.response(w, []byte("internal error"), http.StatusInternalServerError, err)
				return
			}

			bytes, err := diagram.Serialize()
			if err != nil {
				h.response(w, []byte("could not serialise the resulting diagram"), http.StatusInternalServerError, err)
				return
			}
			h.response(w, bytes, http.StatusOK, nil)
			return
		}
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
