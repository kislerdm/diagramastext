package errors

import (
	"log"
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

type Err struct {
	w []byte
}

func (e *Err) Write(p []byte) (n int, err error) {
	e.w = p
	return len(p), nil
}

func (e *Err) Error() string {
	return string(e.w)
}

func NewErrorHandler(msg string) *log.Logger {
	return log.New(&Err{}, "", log.LstdFlags|log.LUTC|log.Llongfile)
}
