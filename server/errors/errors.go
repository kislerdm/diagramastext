package errors

import (
	"strconv"
)

// Error client's error.
type Error struct {
	Service                   string
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
	return service + status + " " + e.Message
}

const (
	ServiceOpenAI         = "OpenAI"
	ServiePlantUML        = "PlantUML"
	ServiceStorage        = "Storage"
	ServiceAWSConfig      = "AWSConfig"
	ServiceSecretsManager = "AWSSecretsmanager"
)
