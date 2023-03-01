package errors

import (
	"strconv"
	"strings"
)

// Error client's error.
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
	service := ""
	if e.Service != "" {
		service = "[service:" + e.Service + "]"
	}
	return "[stage:" + e.Stage + "]" + service + status + " " + e.Message
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
	StageValidation      = "validation"
)

func CombineStages(stages ...string) string {
	return strings.Join(stages, ":")
}
