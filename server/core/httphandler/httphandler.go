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
	apiTokensRepository diagram.RepositoryToken,
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
		// FIXME: add caching layer
		repositoryAPITokens:       apiTokensRepository,
		repositoryRequestsHistory: clientRepositoryPrediction,
	}, nil
}

type httpHandler struct {
	diagramRenderingHandler   map[string]diagram.HTTPHandler
	reportErrorFn             func(err error)
	corsHeaders               corsHeaders
	repositoryAPITokens       diagram.RepositoryToken
	repositoryRequestsHistory diagram.RepositoryPrediction
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
		pathQuotas = "/quotas"
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

	var user diagram.User
	switch strings.HasPrefix(r.URL.Path, pathPrefixInternal) {
	case false:
		if err := h.authorizationAPI(r, &user); err != nil {
			h.response(w, []byte(`{"error":"`+err.(httpHandlerError).Msg+`"}`), err)
			return
		}
	default:
		if err := h.authorizationWebclient(r, &user); err != nil {
			h.response(w, []byte(`{"error":"`+err.(httpHandlerError).Msg+`"}`), err)
			return
		}
		r.URL.Path = strings.TrimPrefix(r.URL.Path, pathPrefixInternal)
	}

	if strings.HasPrefix(r.URL.Path, pathPrefixDiagramGeneration) {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, pathPrefixDiagramGeneration)
		h.diagramRendering(w, r, &user)
		return
	}

	if r.URL.Path == pathQuotas {
		h.getQuotasUsage(w, r, &user)
		return
	}
}

func (h httpHandler) authorizationWebclient(_ *http.Request, user *diagram.User) error {
	user.ID = "00000000-0000-0000-0000-000000000000"
	// TODO: add JWT authN
	return nil
}

func (h httpHandler) authorizationAPI(r *http.Request, user *diagram.User) error {
	authToken := readAuthHeaderValue(r.Header)
	if authToken == "" {
		return httpHandlerError{
			Msg:      "no authorizationWebclient token provided",
			Type:     errorNotAuthorizedNoToken,
			HTTPCode: http.StatusUnauthorized,
		}
	}
	userID, err := h.repositoryAPITokens.GetActiveUserIDByActiveTokenID(r.Context(), authToken)
	if err != nil {
		return httpHandlerError{
			Msg:      "internal error",
			Type:     errorRepositoryToken,
			HTTPCode: http.StatusInternalServerError,
		}
	}
	if userID == "" {
		return httpHandlerError{
			Msg:      "the authorizationWebclient token does not exist, or not active, or account is suspended",
			Type:     errorNotAuthorizedNoToken,
			HTTPCode: http.StatusUnauthorized,
		}
	}

	user.ID = userID
	user.IsRegistered = true
	user.APIToken = authToken

	return h.checkQuota(r, user)
}

func (h httpHandler) checkQuota(r *http.Request, user *diagram.User) error {
	throttling, quotaExceeded, err := diagram.ValidateRequestsQuotaUsage(r.Context(), h.repositoryRequestsHistory, user)
	if err != nil {
		return httpHandlerError{
			Msg:      "internal error",
			Type:     errorQuotaValidation,
			HTTPCode: http.StatusInternalServerError,
		}
	}
	if throttling {
		return httpHandlerError{
			Msg:      "throttling quota exceeded",
			Type:     errorQuotaExceeded,
			HTTPCode: http.StatusTooManyRequests,
		}
	}
	if quotaExceeded {
		return httpHandlerError{
			Msg:      "quota exceeded",
			Type:     errorQuotaExceeded,
			HTTPCode: http.StatusForbidden,
		}
	}
	return nil
}

func (h httpHandler) diagramRendering(w http.ResponseWriter, r *http.Request, user *diagram.User) {
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

		input, err := diagram.NewInput(requestContract.Prompt, user)
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

func (h httpHandler) getQuotasUsage(w http.ResponseWriter, r *http.Request, user *diagram.User) {
	switch r.Method {
	case http.MethodOptions:
		h.response(w, nil, nil)
		return
	case http.MethodGet:
		quotasUsage, err := diagram.GetQuotaUsage(r.Context(), h.repositoryRequestsHistory, user)
		if err != nil {
			h.response(
				w, []byte(`{"error":"internal error"}`), httpHandlerError{
					Msg:      err.Error(),
					Type:     errorQuotaFetching,
					HTTPCode: http.StatusInternalServerError,
				},
			)
			return
		}

		oBytes, err := json.Marshal(quotasUsage)
		if err != nil {
			h.response(
				w, []byte(`{"error":"internal error"}`), httpHandlerError{
					Msg:      err.Error(),
					Type:     errorQuotaDataSerialization,
					HTTPCode: http.StatusInternalServerError,
				},
			)
			return
		}

		h.response(w, oBytes, nil)
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

func readAuthHeaderValue(header http.Header) string {
	const authorizationHeaderName = "Authorization"
	authHeader := header.Get(authorizationHeaderName)
	if authHeader == "" {
		authHeader = header.Get(strings.ToLower(authorizationHeaderName))
	}
	_, v, _ := strings.Cut(authHeader, "Bearer ")
	return v
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
	errorInvalidMethod          = "Request:InvalidMethod"
	errorNotExists              = "Request:HandlerNotExists"
	errorNotAuthorizedNoToken   = "Request:AccessDenied:NoAPIToken"
	errorInvalidRequest         = "InputValidation:InvalidContent"
	errorInvalidPrompt          = "InputValidation:InvalidPrompt"
	errorCoreLogic              = "Core:DiagramRendering"
	errorResponseSerialisation  = "Response:DiagramSerialisation"
	errorRepositoryToken        = "DrivenInterface:RepositoryToken"
	errorQuotaValidation        = "Quota:ValidationError"
	errorQuotaExceeded          = "Quota:Excess"
	errorQuotaFetching          = "Quota:ReadingError"
	errorQuotaDataSerialization = "Quota:SerializationError"
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
