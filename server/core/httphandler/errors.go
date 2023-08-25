package httphandler

import (
	"net/http"

	"github.com/kislerdm/diagramastext/server/core/errors"
)

const (
	errorInvalidMethod             = "Request:InvalidMethod"
	errorNotAuthorizedNoToken      = "Request:AccessDenied:NoAPIToken"
	errorNotAuthorizedInvalidToken = "Request:AccessDenied:InvalidToken"
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
	return errors.HTTPHandlerError{
		Msg:      err.Error(),
		Type:     errorResponseSerialisation,
		HTTPCode: http.StatusInternalServerError,
	}
}

func newModelPredictionError(err error) error {
	return errors.HTTPHandlerError{
		Msg:      err.Error(),
		Type:     errorCoreLogic,
		HTTPCode: http.StatusBadRequest,
	}
}

func newCoreLogicError(err error) error {
	return errors.HTTPHandlerError{
		Msg:      err.Error(),
		Type:     errorCoreLogic,
		HTTPCode: http.StatusInternalServerError,
	}
}

func newInvalidMethodError(err error) error {
	return errors.HTTPHandlerError{
		Msg:      err.Error(),
		Type:     errorInvalidMethod,
		HTTPCode: http.StatusMethodNotAllowed,
	}
}
