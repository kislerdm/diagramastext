package port

type User struct {
	ID                     string
	IsRegistered           bool
	OptOutFromSavingPrompt bool
}

// Input defines the entrypoint interface.
type Input interface {
	Validate() error
	GetUser() *User
	GetPrompt() string
	GetRequestID() string
}

type MockInput struct {
	Err       error
	Prompt    string
	RequestID string
	User      *User
}

func (v MockInput) Validate() error {
	return v.Err
}

func (v MockInput) GetUser() *User {
	return v.User
}

func (v MockInput) GetPrompt() string {
	return v.Prompt
}

func (v MockInput) GetRequestID() string {
	return v.RequestID
}
