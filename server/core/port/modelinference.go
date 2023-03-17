package port

import "context"

type ModelInferenceConfig struct {
	Prompt string
	Model  string
	BestOf uint8
}

// ModelInference interface to communicate with the model.
type ModelInference interface {
	Do(ctx context.Context, cfg ModelInferenceConfig) ([]byte, error)
}

type MockModelInference struct {
	V   []byte
	Err error
}

func (m MockModelInference) Do(_ context.Context, _ ModelInferenceConfig) ([]byte, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.V, nil
}
