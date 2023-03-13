//go:build !unittest
// +build !unittest

package main

//var entrypoint contract.Entrypoint
//
//func init() {
//	awsConfig, err := cloudConfig.LoadDefaultConfig(context.Background())
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	clientSecretsmanager := secretsmanager.NewAWSSecretManagerFromConfig(awsConfig)
//
//	entrypoint, err = core.InitEntrypoint(context.Background(), clientSecretsmanager)
//	if err != nil {
//		log.Fatal(err)
//	}
//}
//
//func main() {
//	c4handler, err := c4container.NewHandler(nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	handler := httpHandler{
//		diagramRenderingHandler: map[string]contract.DiagramHandler{
//			"c4": c4handler,
//		},
//		reportErrorFn: func(err error) { log.Println(err) },
//	}
//
//	if v := os.Getenv("CORS_HEADERS"); v != "" {
//		if err := json.Unmarshal([]byte(v), &handler.corsHeaders); err != nil {
//			log.Fatal(err)
//		}
//	}
//
//	port := 9000
//	if v := utils.MustParseInt(os.Getenv("PORT")); v > 0 {
//		port = v
//	}
//
//	if err := http.ListenAndServe(":"+strconv.Itoa(port), handler); err != nil {
//		log.Println(err)
//	}
//}
//
//type httpHandler struct {
//	diagramRenderingHandler map[string]contract.DiagramHandler
//	reportErrorFn           func(err error)
//	corsHeaders             corsHeaders
//}
//
//func (h httpHandler) response(w http.ResponseWriter, body []byte, status int, err error) {
//	if err != nil {
//		h.reportErrorFn(err)
//	}
//	h.corsHeaders.setHeaders(w.Header())
//	w.WriteHeader(status)
//	_, _ = w.Write(body)
//}
//
//func (h httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
//	switch p := r.URL.Path; p {
//	case "status":
//		if r.Method == http.MethodGet {
//			h.response(w, nil, http.StatusOK, nil)
//			return
//		}
//		h.response(w, nil, http.StatusMethodNotAllowed, nil)
//		return
//	default:
//		renderingHandler, ok := h.diagramRenderingHandler[p]
//		if !ok {
//			h.response(w, nil, http.StatusNotFound, nil)
//			return
//		}
//
//		switch r.Method {
//		case http.MethodOptions:
//			h.response(w, nil, http.StatusOK, nil)
//			return
//		case http.MethodPost:
//			h.serve(context.Background(), w, r, renderingHandler)
//		}
//	}
//}
//
//func (h httpHandler) serve(
//	ctx context.Context, w http.ResponseWriter, r *http.Request, renderingHandler contract.DiagramHandler,
//) {
//	defer func() { _ = r.Body.Close() }()
//	body, err := io.ReadAll(r.Body)
//	if err != nil {
//		h.response(w, []byte("could not recognise the prompt format"), http.StatusUnprocessableEntity, err)
//		return
//	}
//	var input port.Request
//	if err := json.Unmarshal(body, &input); err != nil {
//		h.response(w, []byte("could not recognise the prompt format"), http.StatusUnprocessableEntity, err)
//		return
//	}
//
//	inquiry := contract.Inquiry{
//		Request: &input,
//		UserProfile: &contract.UserProfile{
//			UserID:                 readUserID(r.Header),
//			IsRegistered:           isRegisteredUser(r.Header),
//			OptOutFromSavingPrompt: isOptOutFromSavingPrompt(r.Header),
//		},
//	}
//
//	diagram, err := entrypoint(ctx, inquiry, renderingHandler)
//
//	o, err := json.Marshal(port.Response{SVG: string(diagram)})
//	if err != nil {
//		h.response(w, []byte(err.Error()), http.StatusInternalServerError, err)
//		return
//	}
//
//	h.response(w, o, http.StatusOK, nil)
//}
//
//func isOptOutFromSavingPrompt(headers http.Header) bool {
//	// FIXME: extract registration from JWT when authN is implemented
//	return false
//}
//
//func isRegisteredUser(headers http.Header) bool {
//	// FIXME: extract registration from JWT when authN is implemented
//	return false
//}
//
//func readUserID(headers http.Header) string {
//	// FIXME: extract UserID from the headers when authN is implemented
//	return "NA"
//}
//
//type corsHeaders map[string]string
//
//func (h corsHeaders) setHeaders(header http.Header) {
//	for k, v := range h {
//		header.Set(k, v)
//
//		if k == "Access-Control-Allow-Origin" && (v == "" || v == "'*'") {
//			header.Set(k, "*")
//		}
//	}
//}
