package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/kislerdm/diagramastext/server"
	"github.com/kislerdm/diagramastext/server/pkg/core"
	"github.com/kislerdm/diagramastext/server/pkg/rendering/plantuml"
	"github.com/kislerdm/diagramastext/server/pkg/secretsmanager"
	"github.com/kislerdm/diagramastext/server/pkg/storage"

	coreHandler "github.com/kislerdm/diagramastext/server/handler"
	"github.com/kislerdm/diagramastext/server/utils"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
)

type secret struct {
	OpenAiAPIKey string `json:"openai_api_key"`
	DBHost       string `json:"db_host"`
	DBName       string `json:"db_name"`
	DBUser       string `json:"db_user"`
	DBPassword   string `json:"db_password"`
}

func configureInterfaceClients(ctx context.Context, client secretsmanager.Client, secretARN string) (
	server.ClientInputToGraph, storage.Client, error,
) {
	var s secret

	if err := client.ReadLatestSecret(ctx, secretARN, &s); err != nil {
		s = secret{
			OpenAiAPIKey: os.Getenv("OPENAI_API_KEY"),
			DBHost:       os.Getenv("DB_HOST"),
			DBName:       os.Getenv("DB_DBNAME"),
			DBUser:       os.Getenv("DB_USER"),
			DBPassword:   os.Getenv("DB_PASSWORD"),
		}
	}

	clientOpenAI, err := core.NewOpenAIClient(
		core.ConfigOpenAI{
			Token:       s.OpenAiAPIKey,
			MaxTokens:   utils.MustParseInt(os.Getenv("OPENAI_MAX_TOKENS")),
			Temperature: utils.MustParseFloat32(os.Getenv("OPENAI_TEMPERATURE")),
			Model:       os.Getenv("OPENAI_MODEL"),
		},
	)
	if err != nil {
		return nil, nil, server.Error{
			Service: server.ServiceOpenAI,
			Stage:   server.StageInit,
			Message: err.Error(),
		}
	}

	clientStorage, err := storage.NewPgClient(ctx, s.DBHost, s.DBName, s.DBUser, s.DBPassword)
	if err != nil {
		return nil, nil, server.Error{
			Service: server.ServiceStorage,
			Stage:   server.StageInit,
			Message: err.Error(),
		}
	}

	return clientOpenAI, clientStorage, nil
}

func main() {
	ctx, cancelFn := context.WithTimeout(context.Background(), time.Second*20)
	defer cancelFn()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatal(
			server.Error{
				Service: "aws-config",
				Stage:   server.StageInit,
				Message: err.Error(),
			},
		)
	}

	clientOpenAI, clientStorage, err := configureInterfaceClients(
		ctx, secretsmanager.NewAWSSecretManagerFromConfig(cfg), os.Getenv("ACCESS_CREDENTIALS_ARN"),
	)
	switch err.(type) {
	case nil:
	case server.Error:
		// NOTE: no need to terminate on cold start if no connection to db can be established
		// It is an avoidable UX disruption because we only use db to persist prompts for models finetune yet.
		// FIXME: to remove when the "history" feature is rolled out, i.e. after v0.0.3
		if err.(server.Error).Service == server.ServiceStorage {
			log.Print(err)
		}
	default:
		log.Fatal(err)
	}

	ctx, cancelFn = context.WithTimeout(context.Background(), time.Second*10)
	defer cancelFn()
	defer func() { _ = clientStorage.Close(ctx) }()

	clientPlantUML := plantuml.NewClient()

	var corsHeaders corsHeaders
	if v := os.Getenv("CORS_HEADERS"); v != "" {
		_ = json.Unmarshal([]byte(v), &corsHeaders)
	}

	lambda.Start(handler(clientOpenAI, clientPlantUML, corsHeaders, clientStorage))
}

type corsHeaders map[string]string

func (h corsHeaders) setHeaders(resp events.APIGatewayProxyResponse) events.APIGatewayProxyResponse {
	if h == nil {
		return resp
	}

	if resp.Headers == nil {
		resp.Headers = map[string]string{}
	}

	for k, v := range h {
		resp.Headers[k] = v

		if k == "Access-Control-Allow-Origin" && (v == "" || v == "'*'") {
			resp.Headers[k] = "*"
		}
	}

	return resp
}

func handler(
	clientModel server.ClientInputToGraph, clientDiagram server.ClientGraphToDiagram, corsHeaders corsHeaders,
	clientStorage storage.Client,
) func(
	ctx context.Context, req events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	return func(
		ctx context.Context, req events.APIGatewayProxyRequest,
	) (events.APIGatewayProxyResponse, error) {
		prompt, err := coreHandler.ReadPrompt([]byte(req.Body))
		if err != nil {
			return corsHeaders.setHeaders(
				events.APIGatewayProxyResponse{
					StatusCode: http.StatusUnprocessableEntity,
					Body:       "could not recognise the prompt format",
				},
			), err
		}

		if err := coreHandler.ValidatePrompt(prompt); err != nil {
			return corsHeaders.setHeaders(
				events.APIGatewayProxyResponse{
					StatusCode: http.StatusBadRequest,
					Body:       err.Error(),
				},
			), err
		}

		userID := readUserID(req.Headers)
		requestID := readRequestID(ctx)

		userInput := storage.UserInput{
			CallID: storage.CallID{
				RequestID: requestID,
				UserID:    userID,
			},
			Prompt:    prompt,
			Timestamp: time.Now().UTC(),
		}

		if err := clientStorage.WritePrompt(ctx, userInput); err != nil {
			log.Print("WritePrompt() error " + err.Error())
			if v, err := json.Marshal(userInput); err != nil {
				log.Printf("prompt: %s", string(v))
			}
		}

		graph, err := clientModel.Do(ctx, prompt)
		if err != nil {
			return corsHeaders.setHeaders(parseClientError(err)), err
		}

		prediction, _ := json.Marshal(graph)
		predictionOutput := storage.ModelOutput{
			CallID: storage.CallID{
				RequestID: requestID,
				UserID:    userID,
			},
			Response:  string(prediction),
			Timestamp: time.Now().UTC(),
		}
		if err := clientStorage.WriteModelPrediction(ctx, predictionOutput); err != nil {
			log.Print("WriteModelPrediction() error " + err.Error())
			if v, err := json.Marshal(predictionOutput); err != nil {
				log.Printf("response: %s", string(v))
			}
		}

		svg, err := clientDiagram.Do(ctx, graph)
		if err != nil {
			return corsHeaders.setHeaders(parseClientError(err)), err
		}

		return corsHeaders.setHeaders(
			events.APIGatewayProxyResponse{
				StatusCode: http.StatusOK,
				Body:       string(svg.MustMarshal()),
			},
		), nil
	}
}

func readRequestID(ctx context.Context) string {
	c, _ := lambdacontext.FromContext(ctx)
	return c.AwsRequestID
}

func readUserID(h map[string]string) string {
	// FIXME: extract UserID from the headers when authN is implemented
	return "00000000-0000-0000-0000-000000000000"
}

func parseClientError(err error) events.APIGatewayProxyResponse {
	v := coreHandler.ParseClientError(err)
	return events.APIGatewayProxyResponse{
		StatusCode: v.StatusCode,
		Body:       string(v.Body),
	}
}
