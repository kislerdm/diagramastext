package c4container

import (
	"bytes"
	"context"
	"encoding/json"
	"log"

	"github.com/kislerdm/diagramastext/server/core/diagram"
	"github.com/kislerdm/diagramastext/server/core/errors"
)

// c4ContainersGraph defines the containers and relations for C4 container diagram's graph.
type c4ContainersGraph struct {
	Containers []*container `json:"nodes"`
	Rels       []*rel       `json:"links"`
	Title      string       `json:"title,omitempty"`
	Footer     string       `json:"footer,omitempty"`
	WithLegend bool         `json:"legend,omitempty"`
}

func (l *c4ContainersGraph) UnmarshalJSON(data []byte) error {
	type tmp c4ContainersGraph
	var v tmp
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	if !bytes.Contains(data, []byte(`"legend"`)) {
		v.WithLegend = true
	}
	*l = c4ContainersGraph(v)
	return nil
}

// container C4 container definition.
type container struct {
	ID          string `json:"id"`
	Label       string `json:"label,omitempty"`
	Technology  string `json:"technology,omitempty"`
	Description string `json:"description,omitempty"`
	System      string `json:"group,omitempty"`
	IsExternal  bool   `json:"external,omitempty"`
	IsQueue     bool   `json:"queue,omitempty"`
	IsDatabase  bool   `json:"database,omitempty"`
	IsUser      bool   `json:"user,omitempty"`
}

// rel containers relations.
type rel struct {
	From       string `json:"from"`
	To         string `json:"to"`
	Label      string `json:"label,omitempty"`
	Direction  string `json:"direction,omitempty"`
	Technology string `json:"technology,omitempty"`
}

// NewC4ContainersHTTPHandler initialises the handler to generate C4 containers diagram.
func NewC4ContainersHTTPHandler(
	clientModelInference diagram.ModelInference, clientRepositoryPrediction diagram.RepositoryPrediction,
	httpClient diagram.HTTPClient,
) (diagram.HTTPHandler, error) {
	if clientModelInference == nil {
		return nil, errors.New("model inference client must be provided")
	}
	if httpClient == nil {
		return nil, errors.New("http client must be provided")
	}
	return func(ctx context.Context, input diagram.Input) (diagram.Output, error) {
		if err := input.Validate(); err != nil {
			return nil, err
		}

		if clientRepositoryPrediction != nil {
			if err := clientRepositoryPrediction.WriteInputPrompt(
				ctx, input.GetRequestID(), input.GetUserID(), input.GetPrompt(),
			); err != nil {
				// FIXME: add proper logging
				log.Printf("clientRepositoryPrediction.WriteInputPrompt err: %+v", err)
			}
		}

		predictionRaw, diagramPrediction, usageTokensPrompt, usageTokensCompletions, err := clientModelInference.Do(
			ctx, input.GetPrompt(), contentSystem, model,
		)
		if err != nil {
			return nil, errors.New(err.Error())
		}

		if clientRepositoryPrediction != nil {
			if err := clientRepositoryPrediction.WriteModelResult(
				ctx, input.GetRequestID(), input.GetUserID(), predictionRaw, string(diagramPrediction), model,
				usageTokensPrompt, usageTokensCompletions,
			); err != nil {
				// FIXME: add proper logging
				log.Printf("clientRepositoryPrediction.WriteModelResult err: %+v", err)
			}
		}

		if err := errors.NewPredictionError(diagramPrediction); err != nil {
			return nil, err
		}

		var diagramGraph c4ContainersGraph
		if err := json.Unmarshal(diagramPrediction, &diagramGraph); err != nil {
			return nil, err
		}

		diagramPostRendering, err := renderDiagram(ctx, httpClient, &diagramGraph)
		if err != nil {
			return nil, err
		}

		if clientRepositoryPrediction != nil {
			if err := clientRepositoryPrediction.WriteSuccessFlag(
				ctx, input.GetRequestID(), input.GetUserID(), input.GetUserAPIToken(),
			); err != nil {
				// FIXME: add proper logging
				log.Printf("clientRepositoryPrediction.WriteSuccessFlag err: %+v", err)
			}
		}

		return diagram.NewResultSVG(diagramPostRendering)

	}, nil
}

const model = "gpt-3.5-turbo"

const contentSystem =
// instruction
`Given prompts and corresponding graphs as json define new graph based on new prompt.` +
	`Every node has id,label,group,technology as strings, and external,queue,database,user as bool.` +
	`Every link connects nodes using their id:from,to. It also has label,technology and direction as strings.` +
	`Every json has title and footer as string.` +
	`Output JSON. If error, return {"error": {{detailed decision explanation}} }` + "\n" +

	// example
	`Draw c4 container diagram with four containers,thee of which are external and belong to the system X.
	{"nodes":[{"id":"0"},{"id":"1","group":"X","external":true},{"id":"2","group":"X","external":true},` +
	`{"id":"3","group":"X","external":true}]}` + "\n" +

	// example
	`three connected boxes
	{"nodes":[{"id":"0"},{"id":"1"},{"id":"2"}],` +
	`"links":[{"from":"0","to":"1"},{"from":"1","to":"2"},{"from":"2","to":"0"}]}` + "\n" +

	// example
	`three boxes without legend
	{"nodes":[{"id":"0"},{"id":"1"},{"id":"2"}],` +
	`"links":[{"from":"0","to":"1"},{"from":"1","to":"2"},{"from":"2","to":"0"}],"legend":false}` + "\n" +

	// example
	`three boxes, remove legend
	{"nodes":[{"id":"0"},{"id":"1"},{"id":"2"}],` +
	`"links":[{"from":"0","to":"1"},{"from":"1","to":"2"},{"from":"2","to":"0"}],"legend":false}` + "\n" +

	// example
	`c4 containers:golang web server authenticating users read from external mysql database
	{"nodes":[{"id":"0","label":"Web Server","technology":"Go","description":"Authenticates users"},` + "\n" +
	`{"id":"1","label":"Database","technology":"MySQL","external":true,"database":true}]` + "\n" +
	`"links":[{"from":"0","to":"1","direction":"LR"}]}` + "\n" +

	// example
	`draw c4 diagram with python backend reading from postgres over tcp` + "\n" +
	`{"nodes":[{"id":"0","label":"Postgres","technology":"Postgres","database":true},` +
	`{"id":"1","label":"Backend","technology":"Python"}],` +
	`"links":[{"from":"1","to":"0","label":"reads from postgres","technology":"TCP","direction":"LR"}]}` + "\n" +

	// example
	`draw c4 diagram with java backend reading from dynamoDB over tcp` + "\n" +
	`{"nodes":[{"id":"0","label":"DynamoDB","technology":"DynamoDB","database":true},` +
	`{"id":"1","label":"Backend","technology":"Java"}],` +
	`"links":[{"from":"1","to":"0","label":"reads from dynamoDB","technology":"TCP","direction":"LR"}]}` + "\n" +

	// example
	`c4 diagram with kotlin backend reading from mysql and publishing to kafka avro encoded events` + "\n" +
	`{"nodes":[{"id":"0","label":"Backend","technology":"Kotlin"},` +
	`{"id":"1","label":"Kafka","technology":"Kafka","queue":true},` +
	`{"id":"2","label":"Database","technology":"MySQL","database":true}],` +
	`"links":[{"from":"0","to":"2","label":"reads from database","technology":"TCP","direction":"RL"},` +
	`{"from":"0","to":"2","label":"publishes to kafka","technology":"AVRO/TCP","direction":"LR"}]` +

	// example
	`user interacts with webclient which uses go backend` + "\n" +
	`{"nodes":[{"id":"0","label":"User","user":true},` +
	`{"id":"1","label":"WebClient","technology":"JavaScript"},` +
	`{"id":"2","label":"Backend","technology":"Go"}],` +
	`"links":[{"from":"0","to":"1","label":"Uses","technology":"HTTP","direction":"LR"},` +
	`{"from":"1","to":"2","label":"Uses","technology":"HTTP","direction":"LR"}]}` +

	// example
	`anna calls bob` + "\n" +
	`{"nodes":[{"id":"0","label":"Anna","user":true},{"id":"1","label":"Bob","user":true}],` +
	`"links":[{"from":"0","to":"1","label":"Calls","technology":"Phone","direction":"LR"}]`
