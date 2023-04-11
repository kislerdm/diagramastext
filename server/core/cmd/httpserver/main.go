//go:build !unittest
// +build !unittest

package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/kislerdm/diagramastext/server/core/config"
	"github.com/kislerdm/diagramastext/server/core/httphandler"
	"github.com/kislerdm/diagramastext/server/core/pkg/gcpsecretsmanager"
	"github.com/kislerdm/diagramastext/server/core/pkg/httpclient"
	"github.com/kislerdm/diagramastext/server/core/pkg/openai"
	"github.com/kislerdm/diagramastext/server/core/pkg/postgres"
)

var (
	postgresClient *postgres.Client
	handler        http.Handler
)

func init() {
	var err error
	secretsmanagerClient, err := gcpsecretsmanager.NewSecretmanager(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	cfg := config.LoadDefaultConfig(context.Background(), secretsmanagerClient)

	modelInferenceClient, err := openai.NewOpenAIClient(
		openai.Config{
			Token:     cfg.ModelInferenceConfig.Token,
			MaxTokens: cfg.ModelInferenceConfig.MaxTokens,
			HTTPClient: httpclient.NewHTTPClient(
				httpclient.Config{
					Timeout: 2 * time.Minute,
					Backoff: httpclient.Backoff{
						MaxIterations:             2,
						BackoffTimeMinMillisecond: 50,
						BackoffTimeMaxMillisecond: 300,
					},
				},
			),
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	postgresClient, err = postgres.NewPostgresClient(
		context.Background(), postgres.Config{
			DBHost:             cfg.RepositoryPredictionConfig.DBHost,
			DBName:             cfg.RepositoryPredictionConfig.DBName,
			DBUser:             cfg.RepositoryPredictionConfig.DBUser,
			DBPassword:         cfg.RepositoryPredictionConfig.DBPassword,
			TablePrompt:        cfg.RepositoryPredictionConfig.TablePrompt,
			TablePrediction:    cfg.RepositoryPredictionConfig.TablePrediction,
			TableSuccessStatus: cfg.RepositoryPredictionConfig.TableSuccessStatus,
			TableUsers:         cfg.RepositoryPredictionConfig.TableUsers,
			TableTokens:        cfg.RepositoryPredictionConfig.TableAPITokens,
			SSLMode:            cfg.RepositoryPredictionConfig.SSLMode,
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	var corsHeaders map[string]string
	if v := os.Getenv("CORS_HEADERS"); v != "" {
		if err := json.Unmarshal([]byte(v), &corsHeaders); err != nil {
			log.Fatal(err)
		}
	}

	handler, err = httphandler.NewHTTPHandler(
		modelInferenceClient,
		postgresClient,
		httpclient.NewHTTPClient(
			httpclient.Config{
				Timeout: 1 * time.Minute,
				Backoff: httpclient.Backoff{
					MaxIterations:             2,
					BackoffTimeMinMillisecond: 10,
					BackoffTimeMaxMillisecond: 50,
				},
			},
		),
		corsHeaders,
	)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	defer func() { _ = postgresClient.Close(context.Background()) }()

	portServe := "9000"
	if v := os.Getenv("PORT"); v != "" {
		portServe = v
	}

	if err := http.ListenAndServe(":"+portServe, handler); err != nil {
		log.Println(err)
	}
}
