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
)

// NewHTTPHandler initialises HTTP handler.
func NewHTTPHandler(
	clientModel diagram.ModelInference, clientRepositoryPrediction diagram.RepositoryPrediction,
	httpClientDiagramRendering diagram.HTTPClient, corsHeaders map[string]string,
) (http.Handler, error) {
	var l = log.New(os.Stderr, "", log.Lmicroseconds|log.LUTC|log.Lshortfile)

	c4DiagramHandler, err := c4container.NewC4ContainersHandler(
		clientModel, clientRepositoryPrediction, httpClientDiagramRendering,
	)
	if err != nil {
		return nil, err
	}

	return &httpHandler{
		diagramRenderingHandler: map[string]diagram.DiagramHandler{
			"/c4": c4DiagramHandler,
		},
		corsHeaders:   corsHeaders,
		reportErrorFn: func(err error) { l.Println(err) },
	}, nil
}

type httpHandler struct {
	diagramRenderingHandler map[string]diagram.DiagramHandler
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
	_, _ = w.Write(body)
}

func (h httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch p := r.URL.Path; p {
	case "/status":
		switch r.Method {
		case http.MethodGet, http.MethodOptions:
			h.response(w, nil, nil)
			return
		default:
			h.response(
				w, nil, newInvalidMethodError(
					errors.New("method "+r.Method+" not allowed for path: "+p),
				),
			)
			return
		}

	default:
		renderingHandler, ok := h.diagramRenderingHandler[p]
		if !ok {
			h.response(w, []byte("not exists"), newHandlerNotExistsError(errors.New("handler not exists for path "+p)))
			return
		}

		switch r.Method {
		case http.MethodOptions:
			h.response(w, nil, nil)
			return

		case http.MethodPost:
			defer func() { _ = r.Body.Close() }()
			input, err := diagram.NewInputDriverHTTP(r.Body, r.Header)
			if err != nil {
				h.response(w, []byte("wrong request content"), newValidationError(err))
				return
			}

			generatedDiagram, err := renderingHandler(r.Context(), input)
			if err != nil {
				h.response(w, []byte("internal error"), newCoreLogicError(err))
				return
			}

			bytes, err := generatedDiagram.Serialize()
			if err != nil {
				h.response(w, []byte("internal error"), newDiagramSerialisationError(err))
				return
			}

			h.response(w, bytes, nil)
			return

		default:
			h.response(
				w, nil, newInvalidMethodError(
					errors.New("method "+r.Method+" not allowed for path: "+p),
				),
			)
			return
		}

	}
}

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
	errorInvalidMethod           = "Request:InvalidMethod"
	errorNotExists               = "Request:HandlerNotExists"
	errorInvalidJSON             = "InputValidation:InvalidJSON"
	errorValidJSONInvalidContent = "InputValidation:InvalidContent"
	errorCoreLogic               = "Core:DiagramRendering"
	errorResponseSerialisation   = "Response:DiagramSerialisation"
)

func newDiagramSerialisationError(err error) error {
	return httpHandlerError{
		Msg:      err.Error(),
		Type:     errorResponseSerialisation,
		HTTPCode: http.StatusInternalServerError,
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

func newValidationError(err error) error {
	switch err.(type) {
	case *json.SyntaxError:
		return httpHandlerError{
			Msg:      "faulty JSON",
			Type:     errorInvalidJSON,
			HTTPCode: http.StatusUnprocessableEntity,
		}
	default:
		return httpHandlerError{
			Msg:      err.Error(),
			Type:     errorValidJSONInvalidContent,
			HTTPCode: http.StatusBadRequest,
		}
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
