package httphandler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/kislerdm/diagramastext/server/core/diagram"
	"github.com/kislerdm/diagramastext/server/core/diagram/c4container"
	diagramErrors "github.com/kislerdm/diagramastext/server/core/errors"
)

// NewHTTPHandler initialises HTTP handler.
func NewHTTPHandler(
	clientModel diagram.ModelInference, clientRepositoryPrediction diagram.RepositoryPrediction,
	httpClientDiagramRendering diagram.HTTPClient, corsHeaders map[string]string,
) (http.Handler, error) {
	var l = log.New(os.Stderr, "", log.Lmicroseconds|log.LUTC|log.Lshortfile)

	c4DiagramHandler, err := c4container.NewC4ContainersHTTPHandler(
		clientModel, clientRepositoryPrediction, httpClientDiagramRendering,
	)
	if err != nil {
		return nil, err
	}

	return &httpHandler{
		diagramRenderingHandler: map[string]diagram.HTTPHandler{
			"/c4": c4DiagramHandler,
		},
		corsHeaders:   corsHeaders,
		reportErrorFn: func(err error) { l.Println(err) },
	}, nil
}

type httpHandler struct {
	diagramRenderingHandler map[string]diagram.HTTPHandler
	reportErrorFn           func(err error)
	corsHeaders             corsHeaders
}

func (h httpHandler) response(w http.ResponseWriter, body []byte, err error) {
	status := http.StatusOK

	if err != nil {
		h.reportErrorFn(err)
		status = err.(httpHandlerError).HTTPCode
	}

	h.corsHeaders.setHeaders(w.Header())
	w.WriteHeader(status)
	w.Header().Add("Content-Type", "application/json")
	_, _ = w.Write(body)
}

func (h httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const (
		pathPrefixInternal          = "/internal"
		pathPrefixDiagramGeneration = "/generate"

		pathStatus = "/status"
	)
	if r.URL.Path == pathStatus {
		switch r.Method {
		case http.MethodGet, http.MethodOptions:
			h.response(w, nil, nil)
			return
		default:
			h.response(
				w, nil, newInvalidMethodError(
					errors.New("method "+r.Method+" not allowed for path: "+r.URL.Path),
				),
			)
			return
		}
	}

	if strings.HasPrefix(r.URL.Path, pathPrefixInternal) {
		h.authorization(w, r)
		r.URL.Path = strings.TrimPrefix(r.URL.Path, pathPrefixInternal)
	} else {
		h.response(
			w, []byte(`{"error":"The Rest API will be supported in the upcoming release v0.0.6"}`), httpHandlerError{
				Msg:      "Rest API call attempt",
				Type:     "NotImplemented",
				HTTPCode: http.StatusNotImplemented,
			},
		)
		return
		//h.authorizationAPI(w, r)
	}

	switch strings.HasPrefix(r.URL.Path, pathPrefixDiagramGeneration) {
	case true:
		r.URL.Path = strings.TrimPrefix(r.URL.Path, pathPrefixDiagramGeneration)
		h.diagramRendering(w, r)
	}
}

func (h httpHandler) diagramRendering(w http.ResponseWriter, r *http.Request) {
	routePath := r.URL.Path

	renderingHandler, ok := h.diagramRenderingHandler[routePath]
	if !ok {
		h.response(
			w, []byte(`{"error":"not exists"}`),
			newHandlerNotExistsError(errors.New("handler not exists for path "+routePath)),
		)
		return
	}

	switch r.Method {
	case http.MethodOptions:
		h.response(w, nil, nil)
		return

	case http.MethodPost:
		var requestContract struct {
			Prompt string `json:"prompt"`
		}

		defer func() { _ = r.Body.Close() }()
		if err := json.NewDecoder(r.Body).Decode(&requestContract); err != nil {
			h.response(w, []byte(`{"error":"wrong request format"}`), newInputFormatValidationError(err))
			return
		}

		input, err := diagram.NewInput(requestContract.Prompt, userProfileFromHTTPHeaders(r.Header))
		if err != nil {
			h.response(w, []byte(`{"error":"wrong request content"}`), newInputContentValidationError(err))
			return
		}

		result, err := renderingHandler(r.Context(), input)
		switch err.(type) {
		case nil:
		case diagramErrors.ModelPredictionError:
			h.response(w, err.(diagramErrors.ModelPredictionError).RawJSON, newModelPredictionError(err))
			return
		default:
			h.response(w, []byte(`{"error":"internal error"}`), newCoreLogicError(err))
			return
		}

		bytes, err := result.Serialize()
		if err != nil {
			h.response(w, []byte(`{"error":"internal error"}`), newResponseSerialisationError(err))
			return
		}

		h.response(w, bytes, nil)
		return

	default:
		h.response(
			w, nil, newInvalidMethodError(
				errors.New("method "+r.Method+" not allowed for path: "+routePath),
			),
		)
		return
	}
}

func (h httpHandler) authorization(_ http.ResponseWriter, _ *http.Request) {}

func (h httpHandler) authorizationAPI(_ http.ResponseWriter, _ *http.Request) {}

type corsHeaders map[string]string

func (h corsHeaders) setHeaders(header http.Header) {
	for k, v := range h {
		header.Set(k, v)

		if k == "Access-Control-Allow-Origin" && (v == "" || v == "'*'") {
			header.Set(k, "*")
		}
	}
}

const (
	errorInvalidMethod         = "Request:InvalidMethod"
	errorNotExists             = "Request:HandlerNotExists"
	errorInvalidRequest        = "InputValidation:InvalidContent"
	errorInvalidPrompt         = "InputValidation:InvalidPrompt"
	errorCoreLogic             = "Core:DiagramRendering"
	errorResponseSerialisation = "Response:DiagramSerialisation"
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

func userProfileFromHTTPHeaders(_ http.Header) *diagram.User {
	// FIXME: change when the auth layer is implemented
	return &diagram.User{ID: "00000000-0000-0000-0000-000000000000"}
}
