package httphandler

import (
	"net/http"

	"github.com/kislerdm/diagramastext/server/core/ciam"
)

func NewHandler(
	ciamHandler ciam.HTTPHandlerFn, corsHeaders map[string]string, diagramHandlers http.Handler,
) http.Handler {
	return CORSHandler(
		corsHeaders,
		ResponseTypeHandler(
			"application/json",
			ciamHandler(diagramHandlers),
		),
	)
}
