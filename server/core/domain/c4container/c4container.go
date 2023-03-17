package c4container

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/kislerdm/diagramastext/server/core/adapter"
	"github.com/kislerdm/diagramastext/server/core/port"
)

// c4ContainersGraph defines the containers and relations for C4 container diagram's graph.
type c4ContainersGraph struct {
	Containers []*container `json:"nodes"`
	Rels       []*rel       `json:"links"`
	Title      string       `json:"title,omitempty"`
	Footer     string       `json:"footer,omitempty"`
	WithLegend bool         `json:"with_legend,omitempty"`
}

// container C4 container definition.
type container struct {
	ID          string `json:"id"`
	Label       string `json:"label,omitempty"`
	Technology  string `json:"technology,omitempty"`
	Description string `json:"description,omitempty"`
	System      string `json:"group,omitempty"`
	IsQueue     bool   `json:"is_queue,omitempty"`
	IsDatabase  bool   `json:"is_database,omitempty"`
	IsExternal  bool   `json:"is_external,omitempty"`
	IsUser      bool   `json:"is_user,omitempty"`
}

// rel containers relations.
type rel struct {
	From       string `json:"from"`
	To         string `json:"to"`
	Label      string `json:"label,omitempty"`
	Direction  string `json:"direction,omitempty"`
	Technology string `json:"technology,omitempty"`
}

// NewC4ContainersHandler initialises the handler to generate C4 containers diagram.
func NewC4ContainersHandler(
	clientModelInference port.ModelInference, clientRepositoryPrediction port.RepositoryPrediction,
	httpClient port.HTTPClient,
) port.DiagramHandler {
	if httpClient == nil {
		httpClient = adapter.NewHTTPClient(
			adapter.HTTPClientConfig{
				Timeout: 30 * time.Second,
				Backoff: adapter.Backoff{
					MaxIterations:             2,
					BackoffTimeMinMillisecond: 50,
					BackoffTimeMaxMillisecond: 200,
				},
			},
		)
	}
	return func(ctx context.Context, input port.Input) (port.Output, error) {
		if err := input.Validate(); err != nil {
			return nil, err
		}

		modelConfig := port.ModelInferenceConfig{
			Prompt: addPromptRequestCondition(input.GetPrompt()),
			Model:  defineModel(input.GetUser()),
			BestOf: defineBestOf(input.GetUser()),
		}

		diagramPrediction, err := clientModelInference.Do(ctx, modelConfig)
		if err != nil {
			return nil, err
		}

		if err := clientRepositoryPrediction.WriteInputPrompt(
			ctx, input.GetRequestID(), input.GetUser().ID, input.GetPrompt(),
		); err == nil {
			_ = clientRepositoryPrediction.WriteModelResult(
				ctx, input.GetRequestID(), input.GetUser().ID, string(diagramPrediction),
			)
		}

		var diagramGraph c4ContainersGraph
		if err := json.Unmarshal(diagramPrediction, &diagramGraph); err != nil {
			return nil, err
		}

		diagramPostRendering, err := renderDiagram(ctx, httpClient, &diagramGraph)
		if err != nil {
			return nil, err
		}

		return adapter.NewResultSVG(diagramPostRendering)
	}
}

func defineBestOf(user *port.User) uint8 {
	if user.IsRegistered {
		return 3
	}
	return 2
}

func defineModel(user *port.User) string {
	if user.IsRegistered {
		// FIXME: change for fine-tuned model after it's trained
		return "code-davinci-002"
	}
	return "code-davinci-002"
}

func addPromptRequestCondition(prompt string) string {
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
