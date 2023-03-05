package c4container

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/kislerdm/diagramastext/server/core/contract"
)

// graph defines the diagram graph.
type graph struct {
	Title  string  `json:"title,omitempty"`
	Footer string  `json:"footer,omitempty"`
	Nodes  []*node `json:"nodes"`
	Links  []*link `json:"links,omitempty"`
}

// node diagram's definition node.
type node struct {
	ID         string `json:"id"`
	Label      string `json:"label,omitempty"`
	Group      string `json:"group,omitempty"`
	Technology string `json:"technology,omitempty"`
	External   bool   `json:"external,omitempty"`
	IsQueue    bool   `json:"is_queue,omitempty"`
	IsDatabase bool   `json:"is_database,omitempty"`
}

// link diagram's definition link.
type link struct {
	From       string `json:"from"`
	To         string `json:"to"`
	Direction  string `json:"direction,omitempty"`
	Label      string `json:"label,omitempty"`
	Technology string `json:"technology,omitempty"`
}

// HttpClient http base client.
type HttpClient interface {
	Do(req *http.Request) (resp *http.Response, err error)
}

const defaultTimeout = 1 * time.Minute

// NewHandler initialises the C4 Diagram's handler.
func NewHandler(httpClient HttpClient) (contract.DiagramHandler, error) {
	h := &handler{httpClient: httpClient}
	if h.httpClient == nil {
		h.httpClient = &http.Client{Timeout: defaultTimeout}
	}
	return h, nil
}

type handler struct {
	httpClient HttpClient
}

func (h handler) GetModelRequest(inquery contract.Inquiry) contract.DiagramGraphPredictionRequest {
	return contract.DiagramGraphPredictionRequest{
		Model:  defineModel(inquery.UserProfile),
		Prompt: addPromptRequestCondition(inquery.Prompt),
	}
}

func (h handler) RenderPredictionResultsDiagramSVG(ctx context.Context, prediction []byte) ([]byte, error) {
	var graph graph
	if err := json.Unmarshal(prediction, &graph); err != nil {
		return nil, err
	}

	return renderDiagram(ctx, h.httpClient, graph)
}

func defineModel(user *contract.UserProfile) string {
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
