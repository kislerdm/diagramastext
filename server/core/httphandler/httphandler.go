package httphandler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/kislerdm/diagramastext/server/core/diagram"
	"github.com/kislerdm/diagramastext/server/core/diagram/c4container"
	diagramErrors "github.com/kislerdm/diagramastext/server/core/errors"
	"github.com/kislerdm/diagramastext/server/core/httphandler/ciam"
)

// NewHTTPHandler initialises HTTP handler.
func NewHTTPHandler(
	clientModel diagram.ModelInference, clientRepositoryPrediction diagram.RepositoryPrediction,
	httpClientDiagramRendering diagram.HTTPClient, corsHeaders map[string]string,
	apiTokensRepository diagram.RepositoryToken,
	// requirements for the CIAM handling logic
	ciamClientRepository ciam.RepositoryCIAM, ciamClientKMS ciam.TokenSigningClient, ciamClientSMTP ciam.SMTPClient,
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
		// TODO: move handling logic for API tokens to the CIAM package
		repositoryAPITokens:       apiTokensRepository,
		repositoryRequestsHistory: clientRepositoryPrediction,
		ciam:                      ciam.NewClient(ciamClientRepository, ciamClientKMS, ciamClientSMTP),
	}, nil
}

type httpHandler struct {
	diagramRenderingHandler   map[string]diagram.HTTPHandler
	reportErrorFn             func(err error)
	corsHeaders               corsHeaders
	repositoryAPITokens       diagram.RepositoryToken
	repositoryRequestsHistory diagram.RepositoryPrediction
	ciam                      ciam.Client
}

func (h httpHandler) response(w http.ResponseWriter, body []byte, err error) {
	status := http.StatusOK

	if err != nil {
		h.reportErrorFn(err)
		status = err.(httpHandlerError).HTTPCode
	}

	h.corsHeaders.setHeaders(w.Header())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

func (h httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const (
		pathPrefixInternal          = "/internal"
		pathPrefixDiagramGeneration = "/generate"

		pathStatus = "/status"
		pathQuotas = "/quotas"
		pathCIAM   = "/auth"
	)

	if r.Method == http.MethodOptions {
		h.response(w, nil, nil)
		return
	}

	var user diagram.User

	switch r.URL.Path {
	case pathStatus:
		switch r.Method {
		case http.MethodGet:
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
	case pathQuotas:
		h.getQuotasUsage(w, r, &user)
		return
	case pathCIAM:
		r.URL.Path = strings.TrimPrefix(r.URL.Path, pathCIAM)
		h.ciamHandler(w, r)
		return
	}

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

	h.response(
		w, []byte(`{"error":"resource `+r.URL.Path+` not found"}`),
		newHandlerNotExistsError(errors.New(r.URL.Path+" not found")),
	)
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
			Msg:      "no authorization token provided",
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
			Msg:      "the authorization token does not exist, or not active, or account is suspended",
			Type:     errorNotAuthorizedNoToken,
			HTTPCode: http.StatusUnauthorized,
		}
	}

	user.ID = userID
	user.IsRegistered = true
	user.APIToken = authToken

	return h.checkQuota(r, user)
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

func (h httpHandler) checkQuota(r *http.Request, user *diagram.User) error {
	throttling, quotaExceeded, err := diagram.ValidateRequestsQuotaUsage(r.Context(), h.repositoryRequestsHistory, user)
	if err != nil {
		return httpHandlerError{
			Msg:      "internal error",
			Type:     errorQuotaValidation,
			HTTPCode: http.StatusInternalServerError,
		}
	}
	if quotaExceeded {
		return httpHandlerError{
			Msg:      "quota exceeded",
			Type:     errorQuotaExceeded,
			HTTPCode: http.StatusForbidden,
		}
	}
	if throttling {
		return httpHandlerError{
			Msg:      "throttling quota exceeded",
			Type:     errorQuotaExceeded,
			HTTPCode: http.StatusTooManyRequests,
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
		if err != nil {
			if v, ok := err.(diagramErrors.ModelPredictionError); ok {
				h.response(w, v.RawJSON, newModelPredictionError(err))
				return
			}
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

func (h httpHandler) ciamHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.response(
			w, nil, newInvalidMethodError(
				errors.New("method "+r.Method+" not allowed for path: "+r.URL.Path),
			),
		)
		return
	}

	switch r.URL.Path {
	case "/anonym":
		h.ciamHandlerSigninAnonym(w, r)
	case "/signin/init":
		h.ciamHandlerSigninUserInit(w, r)
	case "/signin/confirm":
		h.ciamHandlerSigninUserConfirm(w, r)
	case "/refresh":
		h.ciamRefreshTokens(w, r)
	default:
		h.response(
			w, []byte(`{"error":"CIAM resource `+r.URL.Path+` not found"}`),
			newHandlerNotExistsError(errors.New("CIAM: "+r.URL.Path+" not found")),
		)
	}
	return
}

func (h httpHandler) ciamHandlerSigninAnonym(w http.ResponseWriter, r *http.Request) {
	defer func() { _ = r.Body.Close() }()
	var req ciamRequestAnonym
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.response(w, []byte(`{"error":"request parsing error"}`), newInputFormatValidationError(err))
		return
	}
	if !req.IsValid() {
		h.response(
			w, []byte(`{"error":"invalid request"}`),
			newInputContentValidationError(errors.New("invalid fingerprint")),
		)
		return
	}

	tokens, err := h.ciam.SigninAnonym(r.Context(), req.Fingerprint)
	if err != nil {
		h.response(
			w, []byte(`{"error":"internal error"}`),
			httpHandlerError{
				Msg:      err.Error(),
				Type:     errorCIAMSigninAnonym,
				HTTPCode: http.StatusInternalServerError,
			},
		)
		return
	}

	o, err := tokens.Serialize()
	// TODO(?): remove error handling at this level
	// motivation: if CIAM generated tokens, their integrity is guaranteed
	if err != nil {
		h.response(
			w, []byte(`{"error":"internal error"}`),
			httpHandlerError{
				Msg:      err.Error(),
				Type:     errorCIAMSigninAnonym,
				HTTPCode: http.StatusInternalServerError,
			},
		)
		return
	}

	h.response(w, o, nil)
	return
}

func (h httpHandler) ciamHandlerSigninUserInit(w http.ResponseWriter, r *http.Request) {
	panic("todo: signin/init logic")
}

func (h httpHandler) ciamHandlerSigninUserConfirm(w http.ResponseWriter, r *http.Request) {
	panic("todo: signin/confirm logic")
}

func (h httpHandler) ciamRefreshTokens(w http.ResponseWriter, r *http.Request) {
	panic("todo: refresh logic")
}
