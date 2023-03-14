package contract

//
//import (
//	"context"
//	"net/http"
//)
//
//// DiagramHandler handler the model inference request's attributes
//// and communicates with the diagram rendering backend.
//type DiagramHandler interface {
//	// GetModelRequest returns the model inference request's attributes to generate the diagram graph.
//	//GetModelRequest(inquery Inquiry) DiagramGraphPredictionRequest
//
//	// RenderPredictionResultsDiagramSVG renders the diagram using the prediction results.
//	RenderPredictionResultsDiagramSVG(ctx context.Context, prediction []byte) ([]byte, error)
//}
//
//// Entrypoint functional entrypoint to render the text prompt to a diagram.
//type Entrypoint func(
//	//ctx context.Context, inquery Inquiry, diagramImplementationHandler DiagramHandler,
//) (
//	[]byte, error,
//)
//
//// ClientModelInference the logic to infer the model and predict a diagram graph.
//type ClientModelInference interface {
//	Do(ctx context.Context, prompt, model string) ([]byte, error)
//}
//
//// MockClientModelInference mock of the model inference client.
//type MockClientModelInference struct {
//	Prediction []byte
//	Err        error
//}
//
//func (m MockClientModelInference) Do(_ context.Context, _, _ string) ([]byte, error) {
//	if m.Err != nil {
//		return nil, m.Err
//	}
//	return m.Prediction, nil
//}
//
//// ClientHTTP http client.
//type ClientHTTP interface {
//	Do(req *http.Request) (resp ClientHTTPResp, err error)
//}
//
//type ClientHTTPResp struct {
//	Body       []byte
//	StatusCode int
//}
//
//// ClientSecretsmanager client to communicate to the secretsmanager.
//type ClientSecretsmanager interface {
//	// ReadLatestSecret reads and deserializes the latest version of JSON-encoded secret.
//	ReadLatestSecret(ctx context.Context, uri string, output interface{}) error
//}
