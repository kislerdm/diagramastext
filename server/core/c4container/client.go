package c4container

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/kislerdm/diagramastext/server/core"
	"github.com/kislerdm/diagramastext/server/core/utils"
)

// Graph defines the diagram graph.
type Graph struct {
	Title  string  `json:"title,omitempty"`
	Footer string  `json:"footer,omitempty"`
	Nodes  []*Node `json:"nodes"`
	Links  []*Link `json:"links,omitempty"`
}

// Node diagram's definition node.
type Node struct {
	ID         string `json:"id"`
	Label      string `json:"label,omitempty"`
	Group      string `json:"group,omitempty"`
	Technology string `json:"technology,omitempty"`
	External   bool   `json:"external,omitempty"`
	IsQueue    bool   `json:"is_queue,omitempty"`
	IsDatabase bool   `json:"is_database,omitempty"`
}

// Link diagram's definition link.
type Link struct {
	From       string `json:"from"`
	To         string `json:"to"`
	Direction  string `json:"direction,omitempty"`
	Label      string `json:"label,omitempty"`
	Technology string `json:"technology,omitempty"`
}

// Client client to generate a diagram artifact, e.g. svg image.
type Client interface {
	Do(context.Context, Graph) ([]byte, error)
}

// HttpClient http base client.
type HttpClient interface {
	Do(req *http.Request) (resp *http.Response, err error)
}

type handler struct {
	clientCore             core.Handler
	clientDiagramRendering Client
}

func (h handler) TextToDiagram(ctx context.Context, req core.Request) ([]byte, error) {
	graphPrediction, err := h.clientCore.InferModel(
		ctx, core.Inquiry{
			Request:           req,
			Model:             defineModel(req),
			PrefixTransformFn: addPromptRequestConditionC4Containers,
		},
	)
	if err != nil {
		return nil, err
	}

	var graph Graph
	if err := json.Unmarshal(graphPrediction, &graph); err != nil {
		return nil, err
	}

	diagram, err := h.clientDiagramRendering.Do(ctx, graph)
	if err != nil {
		return nil, err
	}

	if err := utils.ValidateSVG(diagram); err != nil {
		return nil, err
	}

	return diagram, nil
}

func defineModel(req core.Request) string {
	if req.IsRegisteredUser {
		// FIXME: change for fine-tuned model after it's trained
		return "code-davinci-002"
	}
	return "code-davinci-002"
}

func NewFromConfig(cfg core.Config) (core.Client, error) {
	c, err := core.NewFromConfig(cfg)
	if err != nil {
		return nil, err
	}
	return &handler{
		clientCore:             c,
		clientDiagramRendering: NewClient(),
	}, nil
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
