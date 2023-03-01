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
	"strings"
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

	prompt = addPromptRequestConditionC4Containers(prompt)

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

func addPromptRequestConditionC4Containers(prompt string) string {
	// FIXME: revert back for test
	// FIXME: replace with embed after understanding encoding diff:
	// FIXME: we tried to replace with embedded files already, but openAI was returning 400
	// FIXME: it's likely to do with data encoding -> TBD
	const promptInputComment = `Given prompts and corresponding graphs as json define new graph based on new prompt. ` +
		`Every node has id,label,group,technology as strings,external,is_queue and is_database as bool. ` +
		`Every link connects nodes using their id:from,to. It also has label,technology and direction as strings. ` +
		`Every json has title and footer as string.` + "\n" +
		`Draw c4 container diagram with four containers,thee of which are external and belong to the system X.
	{"nodes":[{"id":"0"},{"id":"1","group":"X","external":true},{"id":"2","group":"X","external":true},` +
		`{"id":"3","group":"X","external":true}]}` + "\n" +
		`three connected boxes
	{"nodes":[{"id":"0"},{"id":"1"},{"id":"2"}],` +
		`"links":[{"from":"0","to":"1"},{"from":"1","to":"2"},{"from":"2","to":"0"}]}` + "\n" +
		`c4 containers:golang web server authenticating users read from external mysql database
	{"nodes":[{"id":"0","label":"Web Server","technology":"Go","description":"Authenticates users"},` + "\n" +
		`{"id":"1","label":"Database","technology":"MySQL","external":true,"is_database":true}]` + "\n" +
		`"links":[{"from":"0","to":"1","direction":"LR"}]}` + "\n" +
		`Five containers in three groups. First container is a Flink Application which performs feature engineering ` +
		`using JSON encoded user behavioural clickstream consumed from AWS Kinesis Stream over HTTP. ` +
		`It publishes AVRO encoded results to the kafka topic over TCP and infers the machine learning model by ` +
		`sending JSON data over rest API. The Flink application is deployed to AWS KDA of the Business Domain account. ` +
		`Kafka topic is part of the Streaming Platform,which sinks the data to the Datalake,AWS S3 bucket. ` +
		`The model is deployed to the MLPlatform. MLPlatform,clickstream and datalake belong to the Data Platform. ` +
		`All but Flink application are external.` + "\n" +
		`{"nodes":[{"id":"0","label":"Go producer","technology":"Go","description":"Publishes to Kafka"},` +
		`{"id":"1","label":"Kafka","technology":"Kafka","is_queue":true},` +
		`{"id":"2","label":"Java consumer","technology":"Java","description":"Consumes from Kafka"}],` +
		`"links":[{"from":"0","to":"1","label":"publishes to Kafka","technology":"TCP/Protobuf"},` +
		`{"from":"2","to":"1","label":"consumes from Kafka","technology":"TCP/Protobuf"}]}` + "\n" +
		`draw c4 diagram with python backend reading from postgres over tcp` + "\n" +
		`{"nodes":[{"id":"0","label":"Postgres","technology":"Postgres","is_database":true},` +
		`{"id":"1","label":"Backend","technology":"Python"}],` +
		`"links":[{"from":"1","to":"0","label":"reads from postgres","technology":"TCP"}]}` + "\n" +
		`draw c4 diagram with java backend reading from dynamoDB over tcp` + "\n" +
		`{"nodes":[{"id":"0","label":"DynamoDB","technology":"DynamoDB","is_database":true},` +
		`{"id":"1","label":"Backend","technology":"Java"}],` +
		`"links":[{"from":"1","to":"0","label":"reads from dynamoDB","technology":"TCP"}]}` + "\n" +
		`c4 diagram with kotlin backend reading from mysql and publishing to kafka avro encoded events` + "\n" +
		`{"nodes":[{"id":"0","label":"Backend","technology":"Kotlin"},` +
		`{"id":"1","label":"Kafka","technology":"Kafka","is_queue":true},` +
		`{"id":"2","label":"Database","technology":"MySQL","is_database":true}],` +
		`"links":[{"from":"0","to":"2","label":"reads from database","technology":"TCP"},` +
		`{"from":"0","to":"2","label":"publishes to kafka","technology":"TCP/AVRO"}]`

	return promptInputComment + "\n" + strings.ReplaceAll(prompt, "\n", "") + "\n"
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
