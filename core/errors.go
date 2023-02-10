package core

import "strconv"

// Error defines the core logic error to facilitate errors handling across the product elements.
type Error struct {
	Service                   string
	Stage                     string
	Message                   string
	ServiceResponseStatusCode int
}

func (e Error) Error() string {
	return "[" + e.Service + "][stage:" + e.Stage + "][http code: " + strconv.Itoa(e.ServiceResponseStatusCode) + "] " +
		e.Message
}

const (
	ServiceOpenAI  = "OpenAI"
	ServiePlantUML = "PlantUML"
)

const (
	StageInit            = "init"
	StageRequest         = "request"
	StageResponse        = "response"
	StageSerialization   = "serialization"
	StageDeserialization = "deserialization"
)
