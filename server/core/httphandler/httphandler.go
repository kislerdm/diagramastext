package httphandler

import (
	"context"
	"log"
	"net/http"
	"os"

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
			"c4": c4DiagramHandler,
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

func (h httpHandler) response(w http.ResponseWriter, body []byte, status int, err error) {
	if err != nil {
		h.reportErrorFn(err)
	}

	h.corsHeaders.setHeaders(w.Header())
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

func (h httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch p := r.URL.Path; p {
	case "status":
		if r.Method == http.MethodGet {
			h.response(w, nil, http.StatusOK, nil)
			return
		}
		h.response(w, nil, http.StatusMethodNotAllowed, nil)
		return
	default:
		renderingHandler, ok := h.diagramRenderingHandler[p]
		if !ok {
			h.response(w, nil, http.StatusNotFound, nil)
			return
		}

		switch r.Method {
		case http.MethodOptions:
			h.response(w, nil, http.StatusOK, nil)
			return
		case http.MethodPost:
			defer func() { _ = r.Body.Close() }()
			input, err := diagram.NewInputDriverHTTP(r.Body, r.Header)
			if err != nil {
				h.response(w, []byte("could not recognise the request"), http.StatusUnprocessableEntity, err)
				return
			}

			generatedDiagram, err := renderingHandler(context.Background(), input)
			if err != nil {
				h.response(w, []byte("internal error"), http.StatusInternalServerError, err)
				return
			}

			bytes, err := generatedDiagram.Serialize()
			if err != nil {
				h.response(w, []byte("could not serialise the resulting diagram"), http.StatusInternalServerError, err)
				return
			}
			h.response(w, bytes, http.StatusOK, nil)
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
