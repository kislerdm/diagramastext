package sdk

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	errs "github.com/kislerdm/diagramastext/server/errors"
	"github.com/kislerdm/diagramastext/server/modelinference"
	"github.com/kislerdm/diagramastext/server/rendering/c4container"
	"github.com/kislerdm/diagramastext/server/secretsmanager"
	"github.com/kislerdm/diagramastext/server/storage"
	"github.com/kislerdm/diagramastext/server/utils"
)

func validateSVG(v []byte) error {
	type svg struct {
		SVG string `xml:"svg"`
	}
	var probe svg
	if err := xml.Unmarshal(v, &probe); err != nil {
		return err
	}
	return nil
}

const (
	PromptLengthMin = 3
	PromptLengthMax = 768
)

func validatePrompt(prompt string) error {
	if len(prompt) < PromptLengthMin || len(prompt) > PromptLengthMax {
		return errors.New(
			"prompt length must be between " + strconv.Itoa(PromptLengthMin) + " and " +
				strconv.Itoa(PromptLengthMax) + " characters",
		)
	}
	return nil
}

// Handler handles diagram generation end-to-end.
type Handler interface {
	// GenerateSVG generates the diagram as SVG given user's input.
	GenerateSVG(ctx context.Context, prompt string, callID storage.CallID) ([]byte, error)

	// Stop closes connections for all clients.
	Stop(ctx context.Context) error
}

// handlerC4Containers generates C4 Container diagrams.
type handlerC4Containers struct {
	clientModel     modelinference.Client
	clientRendering c4container.Client
	clientStorage   storage.Client

	Logger *log.Logger
}

func (h handlerC4Containers) Stop(ctx context.Context) error {
	return h.clientStorage.Close(ctx)
}

func (h handlerC4Containers) GenerateSVG(ctx context.Context, prompt string, callID storage.CallID) ([]byte, error) {
	if err := validatePrompt(prompt); err != nil {
		return nil, errs.Error{
			Stage:                     errs.CombineStages(errs.StageRequest, errs.StageValidation),
			Message:                   err.Error(),
			ServiceResponseStatusCode: http.StatusUnprocessableEntity,
		}
	}

	// FIXME: decide on execution path when db write fails
	if h.clientStorage != nil {
		if err := h.clientStorage.WritePrompt(
			ctx, storage.UserInput{
				CallID:    callID,
				Prompt:    prompt,
				Timestamp: time.Now().UTC(),
			},
		); err != nil {
			h.Logger.Print("[ERROR] " + err.Error())
			h.Logger.Printf("prompt: %s", prompt)
		}
	}

	graphPrediction, err := h.clientModel.Do(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// FIXME: decide on execution path when db write fails
	if h.clientStorage != nil {
		if err := h.clientStorage.WriteModelPrediction(
			ctx, storage.ModelOutput{
				CallID:    callID,
				Response:  string(graphPrediction),
				Timestamp: time.Now().UTC(),
			},
		); err != nil {
			h.Logger.Print("[ERROR] " + err.Error())
			if v, err := json.Marshal(graphPrediction); err != nil {
				h.Logger.Printf("response: %s", string(v))
			}
		}
	}

	var graph c4container.Graph
	if err := json.Unmarshal(graphPrediction, &graph); err != nil {
		return nil, errs.Error{
			Stage:   errs.StageDeserialization,
			Message: err.Error(),
		}
	}

	diagram, err := h.clientRendering.Do(ctx, graph)
	if err != nil {
		return nil, err
	}

	if err := validateSVG(diagram); err != nil {
		return nil, errs.Error{
			Stage:   errs.CombineStages(errs.StageResponse, errs.StageValidation),
			Message: err.Error(),
		}
	}

	return diagram, nil
}

type StorageConfig struct {
	DBHost     string `json:"db_host"`
	DBName     string `json:"db_name"`
	DBUser     string `json:"db_user"`
	DBPassword string `json:"db_password"`
}

type handlerCfg struct {
	modelConfig   modelinference.ConfigOpenAI
	storageConfig StorageConfig
}

type secret struct {
	StorageConfig
	OpenAiAPIKey string `json:"openai_api_key"`
}

func configure(ctx context.Context, client secretsmanager.Client, secretARN string) (handlerCfg, error) {
	envCfg := handlerCfg{
		modelConfig: modelinference.ConfigOpenAI{
			Token:       os.Getenv("OPENAI_API_KEY"),
			MaxTokens:   utils.MustParseInt(os.Getenv("OPENAI_MAX_TOKENS")),
			Temperature: utils.MustParseFloat32(os.Getenv("OPENAI_TEMPERATURE")),
			Model:       os.Getenv("OPENAI_MODEL"),
		},
		storageConfig: StorageConfig{
			DBHost:     os.Getenv("DB_HOST"),
			DBName:     os.Getenv("DB_DBNAME"),
			DBUser:     os.Getenv("DB_USER"),
			DBPassword: os.Getenv("DB_PASSWORD"),
		},
	}

	if client == nil {
		return envCfg, nil
	}

	var s secret

	if err := client.ReadLatestSecret(ctx, secretARN, &s); err != nil {
		return envCfg, nil
	}

	return handlerCfg{
		modelConfig: modelinference.ConfigOpenAI{
			Token: s.OpenAiAPIKey,
		},
		storageConfig: StorageConfig{
			DBHost:     s.DBHost,
			DBName:     s.DBName,
			DBUser:     s.DBUser,
			DBPassword: s.DBPassword,
		},
	}, nil
}

// NewC4DiagramHandler initialises the sdk to generate C4 Container diagram.
func NewC4DiagramHandler(ctx context.Context, secretARN string) (Handler, error) {
	var secretsManagerClient secretsmanager.Client

	// FIXME: add aws config error handling(?)
	if awsCfg, err := config.LoadDefaultConfig(ctx); err != nil {
		secretsManagerClient = secretsmanager.NewAWSSecretManagerFromConfig(awsCfg)
	}

	cfg, err := configure(ctx, secretsManagerClient, secretARN)
	if err != nil {
		return nil, err
	}

	clientOpenAI, err := modelinference.NewOpenAIClient(cfg.modelConfig)
	if err != nil {
		return nil, err
	}

	clientStorage, err := storage.NewPgClient(
		ctx, cfg.storageConfig.DBHost, cfg.storageConfig.DBName, cfg.storageConfig.DBUser, cfg.storageConfig.DBPassword,
	)
	if err != nil {
		log.Print(err.Error())
	}

	return handlerC4Containers{
		clientModel:     clientOpenAI,
		clientRendering: c4container.NewClient(),
		clientStorage:   clientStorage,
		Logger:          nil,
	}, nil
}
