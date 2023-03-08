package port

// Input defines the entrypoint's interface.
type Input interface {
	Validate() error
}

type MockInput struct {
	Err error
}

func (v MockInput) Validate() error {
	return v.Err
}

// Output defines the exit point's interface.
type Output interface {
	Serialize() ([]byte, error)
}

type MockOutput struct {
	V   []byte
	Err error
}

func (m MockOutput) Serialize() ([]byte, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.V, nil
}
