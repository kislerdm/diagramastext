package diagram

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/kislerdm/diagramastext/server/core/ciam"
)

func HTTPHandlers(handlers map[string]HTTPHandler) http.Handler {
	return httphandler{
		diagramHandler: handlers,
		log:            log.New(os.Stderr, "diagram-generator", log.Lmicroseconds|log.LUTC|log.Lshortfile),
	}
}

type httphandler struct {
	diagramHandler map[string]HTTPHandler
	log            *log.Logger
}

func (h httphandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = w.Write([]byte(`{"error":"` + r.Method + ` is not allowed"}`))
		return
	}

	const prefix = "/generate"

	t := strings.TrimLeft(r.URL.Path, prefix)

	handler, ok := h.diagramHandler[t]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"handler for path ` + r.URL.Path + ` is not allowed"}`))
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

	input, err := NewInput(requestContract.Prompt, user.ID, user.APIToken, user.Role.Quotas().PromptLengthMax)
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
}
