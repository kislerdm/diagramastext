package httphandler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

const (
	errorInvalidMethod             = "Request:InvalidMethod"
	errorNotExists                 = "Request:HandlerNotExists"
	errorNotAuthorizedNoToken      = "Request:AccessDenied:NoAPIToken"
	errorNotAuthorizedInvalidToken = "Request:AccessDenied:InvalidToken"
	errorInvalidRequest            = "InputValidation:InvalidContent"
	errorInvalidPrompt             = "InputValidation:InvalidPrompt"
	errorCoreLogic                 = "Core:DiagramRendering"
	errorResponseSerialisation     = "Response:DiagramSerialisation"
	errorRepositoryToken           = "DrivenInterface:RepositoryToken"
	errorQuotaValidation           = "Quota:ValidationError"
	errorQuotaExceeded             = "Quota:Excess"
	errorQuotaFetching             = "Quota:ReadingError"
	errorQuotaDataSerialization    = "Quota:SerializationError"
	errorCIAMSigninAnonym          = "CIAM:Signin:Anonym"
)

func newResponseSerialisationError(err error) error {
	return httpHandlerError{
		Msg:      err.Error(),
		Type:     errorResponseSerialisation,
		HTTPCode: http.StatusInternalServerError,
	}
}

func newModelPredictionError(err error) error {
	return httpHandlerError{
		Msg:      err.Error(),
		Type:     errorCoreLogic,
		HTTPCode: http.StatusBadRequest,
	}
}

func newCoreLogicError(err error) error {
	return httpHandlerError{
		Msg:      err.Error(),
		Type:     errorCoreLogic,
		HTTPCode: http.StatusInternalServerError,
	}
}

func newHandlerNotExistsError(err error) error {
	return httpHandlerError{
		Msg:      err.Error(),
		Type:     errorNotExists,
		HTTPCode: http.StatusNotFound,
	}
}

func newInvalidMethodError(err error) error {
	return httpHandlerError{
		Msg:      err.Error(),
		Type:     errorInvalidMethod,
		HTTPCode: http.StatusMethodNotAllowed,
	}
}

func newInputFormatValidationError(err error) error {
	msg := err.Error()

	switch err.(type) {
	case *json.SyntaxError:
		msg = "faulty JSON"
	}

	return httpHandlerError{
		Msg:      msg,
		Type:     errorInvalidRequest,
		HTTPCode: http.StatusBadRequest,
	}
}

func newInputContentValidationError(err error) error {
	return httpHandlerError{
		Msg:      err.Error(),
		Type:     errorInvalidPrompt,
		HTTPCode: http.StatusUnprocessableEntity,
	}
}

type httpHandlerError struct {
	Msg      string
	Type     string
	HTTPCode int
}

func (e httpHandlerError) Error() string {
	var o strings.Builder
	writeStrings(&o, "[type:", e.Type, "][code:", strconv.Itoa(e.HTTPCode), "] ", e.Msg)
	return o.String()
}

func writeStrings(o *strings.Builder, text ...string) {
	for _, s := range text {
		_, _ = o.WriteString(s)
	}
}
