package port

import "context"

// RepositoryPrediction defines the interface to store prediction input (prompt) and model result.
type RepositoryPrediction interface {
	WriteInputPrompt(ctx context.Context, input Input) error
	WriteModelResult(ctx context.Context, input Input, prediction string) error
}

type MockRepositoryPrediction struct {
	Err error
}

func (m MockRepositoryPrediction) WriteInputPrompt(_ context.Context, _ Input) error {
	return m.Err
}

func (m MockRepositoryPrediction) WriteModelResult(_ context.Context, _ Input, _ string) error {
	return m.Err
}
