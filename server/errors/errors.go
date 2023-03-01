package errors

import "strconv"

// Error `modelinference` client's error.
type Error struct {
	Service                   string
	Stage                     string
	Message                   string
	ServiceResponseStatusCode int
}

func (e Error) Error() string {
	status := ""
	if e.ServiceResponseStatusCode != 0 {
		status = "[http code:" + strconv.Itoa(e.ServiceResponseStatusCode) + "]"
	}
	return "[stage:" + e.Stage + "]" + status + " " + e.Message
}

const (
	ServiceOpenAI            = "OpenAI"
	ServiePlantUML           = "PlantUML"
	ServiceStorage           = "Storage"
	ServiceAWSConfig         = "AWSConfig"
	ServiceAWSSecretsManager = "AWSSecretsmanager"
)

const (
	StageInit            = "init"
	StageConnection      = "connection"
	StageRequest         = "request"
	StageResponse        = "response"
	StageSerialization   = "serialization"
	StageDeserialization = "deserialization"
)
