package httphandler

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/kislerdm/diagramastext/server/core/ciam"
	"github.com/kislerdm/diagramastext/server/core/diagram"
)

func NewHandler(
	ciamHandler ciam.HTTPHandlerFn, corsHeaders map[string]string, diagramHandlers map[string]diagram.HTTPHandler,
) http.Handler {
	return handlerCORS{
		headersMap: corsHeaders,
		next: handlerResponseType{
			mimeType: "application/json",
			next: ciamHandler(
				handlerDiagrams{
					diagramHandlers: diagramHandlers,
					log:             log.New(os.Stderr, "diagram-generator", log.Lmicroseconds|log.LUTC|log.Lshortfile),
				},
			),
		},
	}
}

type handlerDiagrams struct {
	diagramHandlers map[string]diagram.HTTPHandler
	log             *log.Logger
}

func (h handlerDiagrams) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = w.Write([]byte(`{"error":"` + r.Method + ` is not allowed"}`))
		return
	}

	const prefix = "/generate"

	t := strings.TrimPrefix(r.URL.Path, prefix)

	handler, ok := h.diagramHandlers[t]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"` + r.URL.Path + ` not found"}`))
		return
	}

	var requestContract struct {
		Prompt string `json:"prompt"`
	}

	defer func() { _ = r.Body.Close() }()
	if err := json.NewDecoder(r.Body).Decode(&requestContract); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"wrong request format"}`))
		h.log.Println(err)
		return
	}

	user, ok := ciam.FromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"user was not extracted from authorisation token"}`))
		return
	}

	input, err := diagram.NewInput(requestContract.Prompt, user.ID, user.APIToken, user.Role.Quotas().PromptLengthMax)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"error":"wrong request format"}`))
		h.log.Println(err)
		return
	}

	o, err := handler(r.Context(), input)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal error"}`))
		h.log.Println(err)
		return
	}

	oBytes, err := o.Serialize()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal error"}`))
		h.log.Println(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(oBytes)
	return
}

type handlerCORS struct {
	headersMap map[string]string
	next       http.Handler
}

func (c handlerCORS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for k, v := range c.headersMap {
		w.Header().Set(k, v)
		if k == "Access-Control-Allow-Origin" && (v == "" || v == "'*'") {
			w.Header().Set(k, "*")
		}
	}

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if c.next != nil {
		c.next.ServeHTTP(w, r)
	}
}

type handlerResponseType struct {
	mimeType string
	next     http.Handler
}

func (h handlerResponseType) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", h.mimeType)
	if h.next != nil {
		h.next.ServeHTTP(w, r)
	}
}
