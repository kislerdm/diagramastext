package port

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
