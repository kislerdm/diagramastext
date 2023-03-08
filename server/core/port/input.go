package port

type User struct {
	ID                     string
	IsRegistered           bool
	OptOutFromSavingPrompt bool
}

// Input defines the entrypoint interface.
type Input interface {
	Validate() error
	GetUser() User
}

type MockInput struct {
	Err  error
	User User
}

func (v MockInput) Validate() error {
	return v.Err
}

func (v MockInput) GetUser() User {
	return v.User
}
