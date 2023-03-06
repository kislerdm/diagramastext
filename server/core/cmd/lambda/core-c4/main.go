package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	cloudConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/kislerdm/diagramastext/server/core"
	"github.com/kislerdm/diagramastext/server/core/c4container"
	"github.com/kislerdm/diagramastext/server/core/contract"
	"github.com/kislerdm/diagramastext/server/core/secretsmanager"
)

func main() {
	awsConfig, err := cloudConfig.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	clientSecretsmanager := secretsmanager.NewAWSSecretManagerFromConfig(awsConfig)

	entrypoint, err := core.InitEntrypoint(context.Background(), clientSecretsmanager)
	if err != nil {
		log.Fatal(err)
	}

	c4handler, err := c4container.NewHandler(nil)
	if err != nil {
		log.Fatal(err)
	}

	var corsHeaders corsHeaders
	if v := os.Getenv("CORS_HEADERS"); v != "" {
		if err := json.Unmarshal([]byte(v), &corsHeaders); err != nil {
			log.Fatal(err)
		}
	}

	lambda.Start(handler(entrypoint, c4handler, corsHeaders))
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
	handler contract.Entrypoint, diagramHandler contract.DiagramHandler, corsHeaders corsHeaders,
) func(
	ctx context.Context, req events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	return func(
		ctx context.Context, req events.APIGatewayProxyRequest,
	) (events.APIGatewayProxyResponse, error) {
		var input contract.Request
		if err := json.Unmarshal([]byte(req.Body), &input); err != nil {
			log.Print(err)
			return corsHeaders.setHeaders(
				events.APIGatewayProxyResponse{
					StatusCode: http.StatusUnprocessableEntity,
					Body:       "could not recognise the prompt format",
				},
			), err
		}

		inquiry := contract.Inquiry{
			Request: &input,
			UserProfile: &contract.UserProfile{
				UserID:                 readUserID(req.Headers),
				IsRegistered:           isRegisteredUser(req.Headers),
				OptOutFromSavingPrompt: isOptOutFromSavingPrompt(req.Headers),
			},
		}

		diagram, err := handler(ctx, inquiry, diagramHandler)
		if err != nil {
			log.Print(err)
			return corsHeaders.setHeaders(
				events.APIGatewayProxyResponse{
					StatusCode: http.StatusInternalServerError,
					Body:       "failed generating the diagram",
				},
			), err
		}

		output, _ := json.Marshal(contract.Response{SVG: string(diagram)})

		return corsHeaders.setHeaders(
			events.APIGatewayProxyResponse{
				StatusCode: http.StatusOK,
				Body:       string(output),
			},
		), nil
	}
}

func isOptOutFromSavingPrompt(headers map[string]string) bool {
	// FIXME: extract registration from JWT when authN is implemented
	return false
}

func isRegisteredUser(headers map[string]string) bool {
	// FIXME: extract registration from JWT when authN is implemented
	return false
}

func readUserID(headers map[string]string) string {
	// FIXME: extract UserID from the headers when authN is implemented
	return "NA"
}
