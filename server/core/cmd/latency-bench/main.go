//go:build !unittest
// +build !unittest

package main

import (
	"context"
	"log"

	"github.com/kislerdm/diagramastext/server/core/configuration"
	"github.com/kislerdm/diagramastext/server/core/openai"
)

const prompt = `Given prompts and corresponding graphs as json define new graph based on new prompt. ` +
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
	`{"from":"0","to":"2","label":"publishes to kafka","technology":"TCP/AVRO"}]` + "\n" +
	"anna call bob" + "\n"

func main() {
	cfg, err := configuration.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalln(err)
	}

	c, err := openai.NewClient(cfg.ModelInferenceConfig)
	if err != nil {
		log.Fatalln(err)
	}

	call(context.Background(), c)
}

func call(ctx context.Context, c openai.Client) {
	resp, err := c.Do(
		ctx, openai.Request{
			BestOf: 2,
			Prompt: prompt,
			Model:  "code-davinci-002",
		},
	)
	if err != nil {
		log.Printf("error: %+v", err)
	}
	log.Printf("resp: %+v", string(resp))
}
