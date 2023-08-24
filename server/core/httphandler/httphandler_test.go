package httphandler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/kislerdm/diagramastext/server/core/ciam"
	"github.com/kislerdm/diagramastext/server/core/diagram"
	diagramErrors "github.com/kislerdm/diagramastext/server/core/errors"
)

type mockWriter struct {
	Headers    http.Header
	StatusCode int
	V          []byte
}

func (m *mockWriter) Header() http.Header {
	return m.Headers
}

func (m *mockWriter) Write(bytes []byte) (int, error) {
	m.V = bytes
	return len(bytes), nil
}

func (m *mockWriter) WriteHeader(statusCode int) {
	m.StatusCode = statusCode
}

type errCollector struct {
	V error
}

func (c *errCollector) Err(err error) {
	c.V = err
}

func Test_httpHandler_ServeHTTPStatus(t *testing.T) {
	type fields struct {
		diagramRenderingHandler map[string]diagram.HTTPHandler
		corsHeaders             corsHeaders
	}

	corsHeaders := map[string]string{
		"Access-Control-Allow-Origin": "https://diagramastext.dev",
	}

	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		errorsCollector *errCollector
		wantW           http.ResponseWriter
		wantErr         error
	}{
		{
			name:            "happy path: GET /status",
			errorsCollector: &errCollector{},
			fields: fields{
				corsHeaders: map[string]string{
					"Access-Control-Allow-Origin": "'*'",
				},
			},
			args: args{
				w: &mockWriter{
					Headers: http.Header{},
				},
				r: &http.Request{
					Method: http.MethodGet,
					URL: &url.URL{
						Path: "/status",
					},
				},
			},
			wantW: &mockWriter{
				Headers: httpHeaders(
					map[string]string{
						"Access-Control-Allow-Origin": "*",
					},
				),
				StatusCode: http.StatusOK,
				V:          nil,
			},
			wantErr: nil,
		},
		{
			name:            "happy path: OPTIONS /status",
			errorsCollector: &errCollector{},
			fields: fields{
				corsHeaders: map[string]string{
					"Access-Control-Allow-Origin": "'*'",
				},
			},
			args: args{
				w: &mockWriter{
					Headers: http.Header{},
				},
				r: &http.Request{
					Method: http.MethodOptions,
					URL: &url.URL{
						Path: "/status",
					},
				},
			},
			wantW: &mockWriter{
				Headers: httpHeaders(
					map[string]string{
						"Access-Control-Allow-Origin": "*",
					},
				),
				StatusCode: http.StatusOK,
				V:          nil,
			},
			wantErr: nil,
		},
		{
			name:            "unhappy path: POST /status",
			errorsCollector: &errCollector{},
			fields: fields{
				corsHeaders: corsHeaders,
			},
			args: args{
				w: &mockWriter{
					Headers: http.Header{},
				},
				r: &http.Request{
					Method: http.MethodPost,
					URL: &url.URL{
						Path: "/status",
					},
					Body: io.NopCloser(strings.NewReader(`foo`)),
				},
			},
			wantW: &mockWriter{
				Headers:    httpHeaders(corsHeaders),
				StatusCode: http.StatusMethodNotAllowed,
				V:          nil,
			},
			wantErr: newInvalidMethodError(
				errors.New("method " + http.MethodPost + " not allowed for path: /status"),
			),
		},
	}

	t.Parallel()

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				h := httpHandler{
					diagramRenderingHandler: tt.fields.diagramRenderingHandler,
					reportErrorFn:           tt.errorsCollector.Err,
					corsHeaders:             tt.fields.corsHeaders,
				}
				h.ServeHTTP(tt.args.w, tt.args.r)

				if tt.args.w.(*mockWriter).StatusCode != tt.wantW.(*mockWriter).StatusCode {
					t.Errorf("unexpected response status code")
					return
				}

				if !reflect.DeepEqual(tt.args.w.Header(), tt.wantW.Header()) {
					t.Errorf("unexpected response header")
					return
				}

				if !reflect.DeepEqual(tt.args.w.(*mockWriter).V, tt.wantW.(*mockWriter).V) {
					t.Errorf("unexpected response content")
					return
				}

				if !reflect.DeepEqual(tt.errorsCollector.V, tt.wantErr) {
					t.Errorf("unexpected error message collected")
					return
				}
			},
		)
	}
}

func httpHeaders(h map[string]string) http.Header {
	o := http.Header{}
	o.Add("Content-Type", "application/json")
	for k, v := range h {
		o.Add(k, v)
	}
	return o
}

// func Test_httpHandler_ServeHTTPDiagramRenderingHappyPath(t *testing.T) {
// 	corsHeaders := map[string]string{
// 		"Access-Control-Allow-Origin": "https://diagramastext.dev",
// 	}
//
// 	diagramRenderingHandler := map[string]diagram.HTTPHandler{
// 		"/c4": func(_ context.Context, _ diagram.Input) (diagram.Output, error) {
// 			return diagram.MockOutput{
// 				V: []byte(`{"svg":"foo"}`),
// 			}, nil
// 		},
// 	}
//
// 	wantW := mockWriter{
// 		Headers:    httpHeaders(corsHeaders),
// 		StatusCode: http.StatusOK,
// 		V:          []byte(`{"svg":"foo"}`),
// 	}
//
// 	ciamClient, _, token := mustCIAMClientAndJWTandTokenStr("foo@bar.baz")
//
// 	type args struct {
// 		w          http.ResponseWriter
// 		path       string
// 		authHeader string
// 	}
// 	tests := []struct {
// 		name            string
// 		args            args
// 		errorsCollector *errCollector
// 	}{
// 		{
//
// 			name: "webclient",
// 			args: args{
// 				w: &mockWriter{
// 					Headers: http.Header{},
// 				},
// 				path:       "/internal/generate/c4",
// 				authHeader: "Bearer " + token,
// 			},
// 			errorsCollector: &errCollector{},
// 		},
// 		{
//
// 			name: "api",
// 			args: args{
// 				w: &mockWriter{
// 					Headers: http.Header{},
// 				},
// 				path:       "/generate/c4",
// 				authHeader: "Bearer 1410904f-f646-488f-ae08-cc341dfb321c",
// 			},
// 			errorsCollector: &errCollector{},
// 		},
// 	}
//
// 	t.Parallel()
//
// 	for _, tt := range tests {
// 		t.Run(
// 			tt.name, func(t *testing.T) {
// 				h := httpHandler{
// 					diagramRenderingHandler: diagramRenderingHandler,
// 					reportErrorFn:           tt.errorsCollector.Err,
// 					corsHeaders:             corsHeaders,
// 					repositoryAPITokens: &diagram.MockRepositoryToken{
// 						V: "c40bad11-0822-4d84-9f61-44b9a97b0432",
// 					},
// 					repositoryRequestsHistory: &diagram.MockRepositoryPrediction{},
// 					ciam:                      ciamClient,
// 				}
// 				h.ServeHTTP(
// 					tt.args.w, &http.Request{
// 						Method: http.MethodPost,
// 						URL: &url.URL{
// 							Path: tt.args.path,
// 						},
// 						Header: httpHeaders(
// 							map[string]string{
// 								"Authorization": tt.args.authHeader,
// 							},
// 						),
// 						Body: io.NopCloser(strings.NewReader(`{"prompt":"foobar"}`)),
// 					},
// 				)
//
// 				if tt.args.w.(*mockWriter).StatusCode != wantW.StatusCode {
// 					t.Errorf("unexpected response status code")
// 					return
// 				}
//
// 				if !reflect.DeepEqual(tt.args.w.Header(), wantW.Header()) {
// 					t.Errorf("unexpected response header")
// 					return
// 				}
//
// 				if !reflect.DeepEqual(tt.args.w.(*mockWriter).V, wantW.V) {
// 					t.Errorf("unexpected response content")
// 					return
// 				}
//
// 				if tt.errorsCollector.V != nil {
// 					t.Errorf("unexpected error message collected")
// 					return
// 				}
// 			},
// 		)
// 	}
// }

func Test_httpHandler_diagramRendering(t *testing.T) {
	type fields struct {
		diagramRenderingHandler map[string]diagram.HTTPHandler
		corsHeaders             corsHeaders
	}

	corsHeaders := map[string]string{
		"Access-Control-Allow-Origin": "https://diagramastext.dev",
	}

	var httpHeaders = func(h map[string]string) http.Header {
		o := http.Header{}
		o.Add("Content-Type", "application/json")
		for k, v := range h {
			o.Add(k, v)
		}
		return o
	}

	type args struct {
		w    http.ResponseWriter
		r    *http.Request
		user *diagram.User
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		errorsCollector *errCollector
		wantW           http.ResponseWriter
		wantErr         error
	}{
		{

			name:            "happy path: POST /c4",
			errorsCollector: &errCollector{},
			fields: fields{
				diagramRenderingHandler: map[string]diagram.HTTPHandler{
					"/c4": func(_ context.Context, _ diagram.Input) (diagram.Output, error) {
						return diagram.MockOutput{
							V: []byte(`{"svg":"foo"}`),
						}, nil
					},
				},
				corsHeaders: corsHeaders,
			},
			args: args{
				w: &mockWriter{
					Headers: http.Header{},
				},
				r: &http.Request{
					Method: http.MethodPost,
					URL: &url.URL{
						Path: "/c4",
					},
					Body: io.NopCloser(strings.NewReader(`{"prompt":"foobar"}`)),
				},
				user: &diagram.User{},
			},
			wantW: &mockWriter{
				Headers:    httpHeaders(corsHeaders),
				StatusCode: http.StatusOK,
				V:          []byte(`{"svg":"foo"}`),
			},
		},
		{

			name:            "happy path: OPTIONS /c4",
			errorsCollector: &errCollector{},
			fields: fields{
				diagramRenderingHandler: map[string]diagram.HTTPHandler{
					"/c4": func(_ context.Context, _ diagram.Input) (diagram.Output, error) {
						return nil, nil
					},
				},
				corsHeaders: corsHeaders,
			},
			args: args{
				w: &mockWriter{
					Headers: http.Header{},
				},
				r: &http.Request{
					Method: http.MethodOptions,
					URL: &url.URL{
						Path: "/c4",
					},
				},
			},
			wantW: &mockWriter{
				Headers:    httpHeaders(corsHeaders),
				StatusCode: http.StatusOK,
			},
		},
		{
			name: "unhappy path: GET /c4, unsupported method",
			fields: fields{
				diagramRenderingHandler: map[string]diagram.HTTPHandler{
					"/c4": func(_ context.Context, _ diagram.Input) (diagram.Output, error) {
						return nil, nil
					},
				},
				corsHeaders: corsHeaders,
			},
			args: args{
				w: &mockWriter{
					Headers: http.Header{},
				},
				r: &http.Request{
					Method: http.MethodGet,
					URL: &url.URL{
						Path: "/c4",
					},
				},
			},
			errorsCollector: &errCollector{},
			wantW: &mockWriter{
				Headers:    httpHeaders(corsHeaders),
				StatusCode: http.StatusMethodNotAllowed,
			},
			wantErr: newInvalidMethodError(errors.New("method GET not allowed for path: /c4")),
		},
		{
			name:            "unhappy path: POST /c4, faulty input for non registered user",
			errorsCollector: &errCollector{},
			fields: fields{
				diagramRenderingHandler: map[string]diagram.HTTPHandler{
					"/c4": func(_ context.Context, _ diagram.Input) (diagram.Output, error) {
						return nil, nil
					},
				},
				corsHeaders: corsHeaders,
			},
			args: args{
				w: &mockWriter{
					Headers: http.Header{},
				},
				r: &http.Request{
					Method: http.MethodPost,
					URL: &url.URL{
						Path: "/c4",
					},
					Body: io.NopCloser(strings.NewReader(`{"prompt":"a"}`)),
				},
				user: &diagram.User{},
			},
			wantW: &mockWriter{
				Headers:    httpHeaders(corsHeaders),
				StatusCode: http.StatusUnprocessableEntity,
				V:          []byte(`{"error":"wrong request content"}`),
			},
			wantErr: newInputContentValidationError(
				errors.New("prompt length must be between 3 and 100 characters"),
			),
		},
		{
			name:            "unhappy path: POST /c4, faulty input JSON deserialization error",
			errorsCollector: &errCollector{},
			fields: fields{
				diagramRenderingHandler: map[string]diagram.HTTPHandler{
					"/c4": func(_ context.Context, _ diagram.Input) (diagram.Output, error) {
						return nil, nil
					},
				},
				corsHeaders: corsHeaders,
			},
			args: args{
				w: &mockWriter{
					Headers: http.Header{},
				},
				r: &http.Request{
					Method: http.MethodPost,
					URL: &url.URL{
						Path: "/c4",
					},
					Body: io.NopCloser(strings.NewReader(`foo`)),
				},
			},
			wantW: &mockWriter{
				Headers:    httpHeaders(corsHeaders),
				StatusCode: http.StatusBadRequest,
				V:          []byte(`{"error":"wrong request format"}`),
			},
			wantErr: httpHandlerError{
				Msg:      "faulty JSON",
				Type:     errorInvalidRequest,
				HTTPCode: http.StatusBadRequest,
			},
		},
		{
			name:            "unhappy path: POST /c4, model prediction error",
			errorsCollector: &errCollector{},
			fields: fields{
				diagramRenderingHandler: map[string]diagram.HTTPHandler{
					"/c4": func(_ context.Context, _ diagram.Input) (diagram.Output, error) {
						return nil, diagramErrors.NewPredictionError([]byte(`{"error":"qux"}`))
					},
				},
				corsHeaders: corsHeaders,
			},
			args: args{
				w: &mockWriter{
					Headers: http.Header{},
				},
				r: &http.Request{
					Method: http.MethodPost,
					URL: &url.URL{
						Path: "/c4",
					},
					Body: io.NopCloser(strings.NewReader(`{"prompt":"foobar"}`)),
				},
				user: &diagram.User{},
			},
			wantW: &mockWriter{
				Headers:    httpHeaders(corsHeaders),
				StatusCode: http.StatusBadRequest,
				V:          []byte(`{"error":"qux"}`),
			},
			wantErr: newModelPredictionError(diagramErrors.NewPredictionError([]byte(`{"error":"qux"}`))),
		},
		{
			name:            "unhappy path: POST /c4, diagram rendering error",
			errorsCollector: &errCollector{},
			fields: fields{
				diagramRenderingHandler: map[string]diagram.HTTPHandler{
					"/c4": func(_ context.Context, _ diagram.Input) (diagram.Output, error) {
						return nil, errors.New("foobar")
					},
				},
				corsHeaders: corsHeaders,
			},
			args: args{
				w: &mockWriter{
					Headers: http.Header{},
				},
				r: &http.Request{
					Method: http.MethodPost,
					URL: &url.URL{
						Path: "/c4",
					},
					Body: io.NopCloser(strings.NewReader(`{"prompt": "foobar"}`)),
				},
				user: &diagram.User{Role: diagram.RoleAnonymUser},
			},
			wantW: &mockWriter{
				Headers:    httpHeaders(corsHeaders),
				StatusCode: http.StatusInternalServerError,
				V:          []byte(`{"error":"internal error"}`),
			},
			wantErr: newCoreLogicError(errors.New("foobar")),
		},
		{
			name:            "unhappy path: POST /c4, diagram response serialisation error",
			errorsCollector: &errCollector{},
			fields: fields{
				diagramRenderingHandler: map[string]diagram.HTTPHandler{
					"/c4": func(_ context.Context, _ diagram.Input) (diagram.Output, error) {
						return diagram.MockOutput{Err: errors.New("foobar")}, nil
					},
				},
				corsHeaders: corsHeaders,
			},
			args: args{
				w: &mockWriter{
					Headers: http.Header{},
				},
				r: &http.Request{
					Method: http.MethodPost,
					URL: &url.URL{
						Path: "/c4",
					},
					Body: io.NopCloser(strings.NewReader(`{"prompt": "foobar"}`)),
				},
				user: &diagram.User{},
			},
			wantW: &mockWriter{
				Headers:    httpHeaders(corsHeaders),
				StatusCode: http.StatusInternalServerError,
				V:          []byte(`{"error":"internal error"}`),
			},
			wantErr: newResponseSerialisationError(errors.New("foobar")),
		},
		{
			name:            "unhappy path: path not found",
			errorsCollector: &errCollector{},
			fields: fields{
				diagramRenderingHandler: map[string]diagram.HTTPHandler{
					"/c4": func(_ context.Context, _ diagram.Input) (diagram.Output, error) {
						return nil, nil
					},
				},
				corsHeaders: corsHeaders,
			},
			args: args{
				w: &mockWriter{
					Headers: http.Header{},
				},
				r: &http.Request{
					Method: http.MethodOptions,
					URL: &url.URL{
						Path: "/notFound",
					},
				},
			},
			wantW: &mockWriter{
				Headers:    httpHeaders(corsHeaders),
				StatusCode: http.StatusNotFound,
				V:          []byte(`{"error":"not exists"}`),
			},
			wantErr: newHandlerNotExistsError(errors.New("handler not exists for path /notFound")),
		},
	}

	t.Parallel()
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				h := httpHandler{
					diagramRenderingHandler: tt.fields.diagramRenderingHandler,
					reportErrorFn:           tt.errorsCollector.Err,
					corsHeaders:             tt.fields.corsHeaders,
				}
				h.diagramRendering(tt.args.w, tt.args.r, tt.args.user)

				if tt.args.w.(*mockWriter).StatusCode != tt.wantW.(*mockWriter).StatusCode {
					t.Errorf("unexpected response status code")
					return
				}

				if !reflect.DeepEqual(tt.args.w.Header(), tt.wantW.Header()) {
					t.Errorf("unexpected response header")
					return
				}

				if !reflect.DeepEqual(tt.args.w.(*mockWriter).V, tt.wantW.(*mockWriter).V) {
					t.Errorf("unexpected response content")
					return
				}

				if !reflect.DeepEqual(tt.errorsCollector.V, tt.wantErr) {
					t.Errorf("unexpected error message collected")
					return
				}
			},
		)
	}
}

func Test_httpHandlerError_Error(t *testing.T) {
	t.Run(
		"error message test", func(t *testing.T) {
			// GIVEN
			err := httpHandlerError{
				Msg:      "foobar",
				Type:     errorCoreLogic,
				HTTPCode: http.StatusInternalServerError,
			}
			want := `[type:Core:DiagramRendering][code:500] foobar`

			// WHEN
			got := err.Error()

			// THEN
			if got != want {
				t.Errorf("unexpected error print")
			}
		},
	)
}

func Test_httpHandler_authorizationAPI(t *testing.T) {
	type fields struct {
		repositoryAPITokens       diagram.RepositoryToken
		repositoryRequestsHistory diagram.RepositoryPrediction
	}
	type args struct {
		r    *http.Request
		user *diagram.User
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		want     error
		wantUser *diagram.User
	}{
		{
			name: "happy path",
			fields: fields{
				repositoryAPITokens: &diagram.MockRepositoryToken{
					V: "bar",
				},
				repositoryRequestsHistory: &diagram.MockRepositoryPrediction{},
			},
			args: args{
				r: &http.Request{
					Header: httpHeaders(
						map[string]string{
							"Authorization": "Bearer foo",
						},
					),
				},
				user: &diagram.User{},
			},
			want: nil,
			wantUser: &diagram.User{
				ID:       "bar",
				APIToken: "foo",
				Role:     diagram.RoleRegisteredUser,
			},
		},
		{
			name: "unhappy path: no auth token found",
			fields: fields{
				repositoryAPITokens: &diagram.MockRepositoryToken{
					V: "bar",
				},
			},
			args: args{
				r: &http.Request{
					Header: http.Header{},
				},
			},
			want: httpHandlerError{
				Msg:      "no authorization token provided",
				Type:     errorNotAuthorizedNoToken,
				HTTPCode: http.StatusUnauthorized,
			},
		},
		{
			name: "unhappy path: failed to interact with the repository",
			fields: fields{
				repositoryAPITokens: &diagram.MockRepositoryToken{
					Err: errors.New("foobar"),
				},
			},
			args: args{
				r: &http.Request{
					Header: httpHeaders(
						map[string]string{
							"Authorization": "Bearer foo",
						},
					),
				},
			},
			want: httpHandlerError{
				Msg:      "internal error",
				Type:     errorRepositoryToken,
				HTTPCode: http.StatusInternalServerError,
			},
		},
		{
			name: "unhappy path: failed to find token",
			fields: fields{
				repositoryAPITokens: &diagram.MockRepositoryToken{},
			},
			args: args{
				r: &http.Request{
					Header: httpHeaders(
						map[string]string{
							"Authorization": "Bearer foo",
						},
					),
				},
			},
			want: httpHandlerError{
				Msg:      "the authorization token does not exist, or not active, or account is suspended",
				Type:     errorNotAuthorizedNoToken,
				HTTPCode: http.StatusUnauthorized,
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				h := httpHandler{
					repositoryAPITokens:       tt.fields.repositoryAPITokens,
					repositoryRequestsHistory: tt.fields.repositoryRequestsHistory,
				}

				if got := h.authorizationAPI(tt.args.r, tt.args.user); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("unexpected error message collected")
					return
				}

				if !reflect.DeepEqual(tt.args.user, tt.wantUser) {
					t.Errorf("unexpected user data fetched")
				}
			},
		)
	}
}

func Test_httpHandler_authorizationWebclient(t *testing.T) {
	type fields struct {
		ciam ciam.Client
	}
	type args struct {
		r    *http.Request
		user *diagram.User
	}

	ciamClient, _, token := mustCIAMClientAndJWTandTokenStr("foo@bar.baz")
	userID := ciamClient.(*ciam.MockCIAMClient).UserID

	tests := []struct {
		name     string
		fields   fields
		args     args
		want     error
		wantUser *diagram.User
	}{
		{
			name: "happy path",
			fields: fields{
				ciam: ciamClient,
			},
			args: args{
				r: &http.Request{
					Header: httpHeaders(
						map[string]string{
							"Authorization": "Bearer " + token,
						},
					),
				},
				user: &diagram.User{},
			},
			want: nil,
			wantUser: &diagram.User{
				ID:   userID,
				Role: diagram.RoleRegisteredUser,
			},
		},
		{
			name: "unhappy path: no token",
			args: args{
				r:    &http.Request{},
				user: &diagram.User{},
			},
			want: httpHandlerError{
				Msg:      "no authorization token provided",
				Type:     errorNotAuthorizedNoToken,
				HTTPCode: http.StatusUnauthorized,
			},
			wantUser: &diagram.User{},
		},
		{
			name: "unhappy path: invalid token",
			fields: fields{
				ciam: &ciam.MockCIAMClient{
					Err: errors.New("foobar"),
				},
			},
			args: args{
				r: &http.Request{
					Header: httpHeaders(
						map[string]string{
							"Authorization": "Bearer foobar",
						},
					),
				},
				user: &diagram.User{},
			},
			want: httpHandlerError{
				Msg:      "invalid access token",
				Type:     errorNotAuthorizedInvalidToken,
				HTTPCode: http.StatusUnauthorized,
			},
			wantUser: &diagram.User{},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				h := httpHandler{
					ciam: tt.fields.ciam,
				}

				if got := h.authorizationWebclient(tt.args.r, tt.args.user); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("unexpected error message collected")
					return
				}

				if !reflect.DeepEqual(tt.args.user, tt.wantUser) {
					t.Errorf("unexpected user data fetched")
				}
			},
		)
	}
}

func Test_readAuthHeaderValue(t *testing.T) {
	type args struct {
		header http.Header
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "extract Bearer token",
			args: args{
				header: httpHeaders(
					map[string]string{
						"Authorization": "Bearer foo",
					},
				),
			},
			want: "foo",
		},
		{
			name: "extract Bearer token - lower case",
			args: args{
				header: httpHeaders(
					map[string]string{
						"authorization": "Bearer foo",
					},
				),
			},
			want: "foo",
		},
		{
			name: "no token header found",
			args: args{
				header: httpHeaders(nil),
			},
			want: "",
		},
		{
			name: "token is present in headers, but wrong format",
			args: args{
				header: httpHeaders(
					map[string]string{
						"Authorization": "foo",
					},
				),
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := readAuthHeaderValue(tt.args.header); got != tt.want {
					t.Errorf("readAuthHeaderValue() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_httpHandler_HandlerNotFound(t *testing.T) {
	t.Parallel()
	t.Run(
		"api", func(t *testing.T) {
			// GIVEN
			errs := &errCollector{}

			handler := httpHandler{
				reportErrorFn:             errs.Err,
				repositoryAPITokens:       diagram.MockRepositoryToken{V: "bar"},
				repositoryRequestsHistory: diagram.MockRepositoryPrediction{},
				corsHeaders:               corsHeaders{},
			}

			r := &http.Request{
				Method: http.MethodGet,
				URL:    &url.URL{Path: "/foo/bar/qux/unknown"},
				Header: httpHeaders(
					map[string]string{
						"Authorization": "Bearer foo",
					},
				),
			}

			w := &mockWriter{
				Headers: httpHeaders(nil),
			}

			// WHEN
			handler.ServeHTTP(w, r)

			// THEN
			if !reflect.DeepEqual(w.V, []byte(`{"error":"resource `+r.URL.Path+` not found"}`)) {
				t.Errorf("unexpected response body")
			}

			if !reflect.DeepEqual(errs.V, newHandlerNotExistsError(errors.New(r.URL.Path+" not found"))) {
				t.Errorf("unexpected error reported")
			}
		},
	)

	t.Run(
		"webclient", func(t *testing.T) {
			// GIVEN
			errs := &errCollector{}

			ciamClient, _, token := mustCIAMClientAndJWTandTokenStr("foo@bar.baz")

			handler := httpHandler{
				reportErrorFn:             errs.Err,
				repositoryAPITokens:       diagram.MockRepositoryToken{V: "bar"},
				repositoryRequestsHistory: diagram.MockRepositoryPrediction{},
				corsHeaders:               corsHeaders{},
				ciam:                      ciamClient,
			}

			r := &http.Request{
				Method: http.MethodGet,
				URL:    &url.URL{Path: "/internal/foo/bar/qux/unknown"},
				Header: httpHeaders(
					map[string]string{
						"Authorization": "Bearer " + token,
					},
				),
			}

			w := &mockWriter{
				Headers: httpHeaders(nil),
			}

			// WHEN
			handler.ServeHTTP(w, r)

			// THEN
			if !reflect.DeepEqual(w.V, []byte(`{"error":"resource `+r.URL.Path+` not found"}`)) {
				t.Errorf("unexpected response body")
			}

			if !reflect.DeepEqual(errs.V, newHandlerNotExistsError(errors.New(r.URL.Path+" not found"))) {
				t.Errorf("unexpected error reported")
			}
		},
	)
}

func Test_httpHandler_GetQuotas(t *testing.T) {
	t.Run(
		"shall follow the happy path", func(t *testing.T) {
			// GIVEN
			handler := httpHandler{
				repositoryAPITokens:       diagram.MockRepositoryToken{V: "bar"},
				repositoryRequestsHistory: diagram.MockRepositoryPrediction{},
				corsHeaders:               corsHeaders{},
			}

			r := &http.Request{
				Method: http.MethodGet,
				URL:    &url.URL{Path: "/quotas"},
				Header: httpHeaders(
					map[string]string{
						"Authorization": "Bearer foo",
					},
				),
			}

			w := &mockWriter{
				Headers: httpHeaders(nil),
			}

			// WHEN
			handler.ServeHTTP(w, r)

			// THEN
			if w.StatusCode != http.StatusOK {
				t.Errorf("unexpected response")
			}
		},
	)
}

func Test_getQuotasUsage(t *testing.T) {
	t.Parallel()
	t.Run(
		"shall follow the happy path", func(t *testing.T) {
			// GIVEN
			handler := httpHandler{
				repositoryAPITokens:       diagram.MockRepositoryToken{V: "bar"},
				repositoryRequestsHistory: diagram.MockRepositoryPrediction{},
				corsHeaders:               corsHeaders{},
			}

			r := &http.Request{
				Method: http.MethodGet,
				URL:    &url.URL{Path: "/quotas"},
				Header: httpHeaders(
					map[string]string{
						"Authorization": "Bearer foo",
					},
				),
			}

			w := &mockWriter{
				Headers: httpHeaders(nil),
			}

			// WHEN
			handler.getQuotasUsage(w, r, &diagram.User{})

			// THEN
			if w.StatusCode != http.StatusOK {
				t.Errorf("unexpected response")
			}
		},
	)

	t.Run(
		"shall return method not allowed", func(t *testing.T) {
			// GIVEN
			errs := &errCollector{}
			handler := httpHandler{
				reportErrorFn:             errs.Err,
				corsHeaders:               corsHeaders{},
				repositoryAPITokens:       diagram.MockRepositoryToken{V: "bar"},
				repositoryRequestsHistory: diagram.MockRepositoryPrediction{},
			}

			r := &http.Request{
				Method: http.MethodDelete,
				URL:    &url.URL{Path: "/quotas"},
				Header: httpHeaders(
					map[string]string{
						"Authorization": "Bearer foo",
					},
				),
			}

			w := &mockWriter{
				Headers: httpHeaders(nil),
			}

			// WHEN
			handler.getQuotasUsage(w, r, &diagram.User{})

			// THEN
			if w.StatusCode != http.StatusMethodNotAllowed {
				t.Errorf("unexpected response")
			}

			if !reflect.DeepEqual(
				errs.V, newInvalidMethodError(
					errors.New("method "+r.Method+" not allowed for path: "+r.URL.Path),
				),
			) {
				t.Errorf("unexpected error reported")
			}
		},
	)

	t.Run(
		"shall return internal error caused by interaction error with repository", func(t *testing.T) {
			// GIVEN
			errs := &errCollector{}
			handler := httpHandler{
				reportErrorFn:       errs.Err,
				corsHeaders:         corsHeaders{},
				repositoryAPITokens: diagram.MockRepositoryToken{V: "bar"},
				repositoryRequestsHistory: diagram.MockRepositoryPrediction{
					Err: errors.New("foobar"),
				},
			}

			r := &http.Request{
				Method: http.MethodGet,
				URL:    &url.URL{Path: "/quotas"},
				Header: httpHeaders(
					map[string]string{
						"Authorization": "Bearer foo",
					},
				),
			}

			w := &mockWriter{
				Headers: httpHeaders(nil),
			}

			// WHEN
			handler.getQuotasUsage(w, r, &diagram.User{})

			// THEN
			if w.StatusCode != http.StatusInternalServerError {
				t.Errorf("unexpected response")
			}

			if !reflect.DeepEqual(w.V, []byte(`{"error":"internal error"}`)) {
				t.Errorf("unexpected response body")
			}

			if !reflect.DeepEqual(
				errs.V, httpHandlerError{
					Msg:      "foobar",
					Type:     errorQuotaFetching,
					HTTPCode: http.StatusInternalServerError,
				},
			) {
				t.Errorf("unexpected error reported")
			}
		},
	)
}

func repeatTimestamp(ts time.Time, nElements uint16) []time.Time {
	o := make([]time.Time, nElements)
	var i uint16
	for i < nElements {
		o[i] = ts
		i++
	}
	return o
}

func Test_httpHandler_checkQuota(t *testing.T) {
	type fields struct {
		diagramRenderingHandler   map[string]diagram.HTTPHandler
		reportErrorFn             func(err error)
		corsHeaders               corsHeaders
		repositoryAPITokens       diagram.RepositoryToken
		repositoryRequestsHistory diagram.RepositoryPrediction
	}
	type args struct {
		r    *http.Request
		user *diagram.User
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "shall yield no quota excess",
			fields: fields{
				repositoryRequestsHistory: diagram.MockRepositoryPrediction{},
			},
			args: args{
				r: &http.Request{},
				user: &diagram.User{
					Quotas: ciam.QuotasAnonymUser,
				},
			},
			wantErr: nil,
		},
		{
			name: "shall yield quota excess",
			fields: fields{
				repositoryRequestsHistory: diagram.MockRepositoryPrediction{
					Timestamps: repeatTimestamp(time.Now().UTC(), ciam.QuotasAnonymUser.RequestsPerDay),
				},
			},
			args: args{
				r:    &http.Request{},
				user: &diagram.User{Quotas: ciam.QuotasAnonymUser},
			},
			wantErr: httpHandlerError{
				Msg:      "quota exceeded",
				Type:     errorQuotaExceeded,
				HTTPCode: http.StatusForbidden,
			},
		},
		{
			name: "shall yield throttling error",
			fields: fields{
				repositoryRequestsHistory: diagram.MockRepositoryPrediction{
					Timestamps: repeatTimestamp(time.Now().UTC(), ciam.QuotasAnonymUser.RequestsPerMinute),
				},
			},
			args: args{
				r:    &http.Request{},
				user: &diagram.User{Quotas: ciam.QuotasAnonymUser},
			},
			wantErr: httpHandlerError{
				Msg:      "throttling quota exceeded",
				Type:     errorQuotaExceeded,
				HTTPCode: http.StatusTooManyRequests,
			},
		},
		{
			name: "shall return in internal error",
			fields: fields{
				repositoryRequestsHistory: diagram.MockRepositoryPrediction{
					Err: errors.New("foo"),
				},
			},
			args: args{
				r:    &http.Request{},
				user: &diagram.User{},
			},
			wantErr: httpHandlerError{
				Msg:      "internal error",
				Type:     errorQuotaValidation,
				HTTPCode: http.StatusInternalServerError,
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				h := httpHandler{
					repositoryRequestsHistory: tt.fields.repositoryRequestsHistory,
				}
				if err := h.checkQuota(tt.args.r, tt.args.user); !reflect.DeepEqual(err, tt.wantErr) {
					t.Errorf("checkQuota() error = %v, wantErr %v", err, tt.wantErr)
				}
			},
		)
	}
}

func Test_httpHandler_response(t *testing.T) {
	t.Run(
		"shall have the Content-Type header set to application/json", func(t *testing.T) {
			// GIVEN
			handler := httpHandler{
				corsHeaders: corsHeaders{},
			}
			w := &mockWriter{Headers: httpHeaders(nil)}

			// WHEN
			handler.response(w, nil, nil)

			// THEN
			if w.StatusCode != http.StatusOK {
				t.Errorf("unexpected response status code")
			}

			if w.Headers.Get("Content-Type") != "application/json" {
				t.Errorf("unexpected Content-Type set")
			}
		},
	)
}

func Test_httpHandler_ciamHandler(t *testing.T) {
	type fields struct {
		diagramRenderingHandler   map[string]diagram.HTTPHandler
		reportErrorFn             func(err error)
		corsHeaders               corsHeaders
		repositoryAPITokens       diagram.RepositoryToken
		repositoryRequestsHistory diagram.RepositoryPrediction
		ciam                      ciam.Client
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		wantStatusCode int
	}{
		{
			name:   "unhappy path: not allowed method",
			fields: fields{reportErrorFn: func(err error) {}},
			args: args{
				r: &http.Request{
					Method: http.MethodGet,
					URL:    &url.URL{Path: "/auth/foobar"},
				},
				w: &mockWriter{
					Headers: httpHeaders(nil),
				},
			},
			wantStatusCode: http.StatusMethodNotAllowed,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				h := httpHandler{
					diagramRenderingHandler:   tt.fields.diagramRenderingHandler,
					reportErrorFn:             tt.fields.reportErrorFn,
					corsHeaders:               tt.fields.corsHeaders,
					repositoryAPITokens:       tt.fields.repositoryAPITokens,
					repositoryRequestsHistory: tt.fields.repositoryRequestsHistory,
					ciam:                      tt.fields.ciam,
				}
				h.ciamHandler(tt.args.w, tt.args.r)
				if tt.args.w.(*mockWriter).StatusCode != tt.wantStatusCode {
					t.Errorf(
						"unexpected status code. want = %d, got = %d", tt.wantStatusCode,
						tt.args.w.(*mockWriter).StatusCode,
					)
				}
			},
		)
	}
}

func Test_httpHandler_ciamHandlerSigninAnonym(t *testing.T) {
	type fields struct {
		diagramRenderingHandler   map[string]diagram.HTTPHandler
		corsHeaders               corsHeaders
		repositoryAPITokens       diagram.RepositoryToken
		repositoryRequestsHistory diagram.RepositoryPrediction
		ciam                      ciam.Client
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		wantStatusCode  int
		wantErr         error
		wantBody        []byte
		errorsCollector *errCollector
	}{
		{
			name: "happy path",
			fields: fields{
				diagramRenderingHandler: nil,
				ciam:                    &ciam.MockCIAMClient{},
			},
			args: args{
				r: &http.Request{
					Body: io.NopCloser(
						strings.NewReader(`{"fingerprint": "9468a4a53a2f2fd9ea96db22dc9dd9bb6ce38b7c"}`),
					),
				},
				w: &mockWriter{
					Headers: httpHeaders(nil),
				},
			},
			wantStatusCode:  http.StatusOK,
			wantErr:         nil,
			errorsCollector: &errCollector{},
		},
		{
			name: "invalid input body: format",
			fields: fields{
				diagramRenderingHandler: nil,
				ciam:                    &ciam.MockCIAMClient{},
			},
			args: args{
				r: &http.Request{
					Body: io.NopCloser(strings.NewReader(`{""}`)),
				},
				w: &mockWriter{
					Headers: httpHeaders(nil),
				},
			},
			wantStatusCode: http.StatusBadRequest,
			wantBody:       []byte(`{"error":"request parsing error"}`),
			wantErr: httpHandlerError{
				Msg:      "faulty JSON",
				Type:     errorInvalidRequest,
				HTTPCode: http.StatusBadRequest,
			},
			errorsCollector: &errCollector{},
		},
		{
			name: "invalid input body: content",
			fields: fields{
				diagramRenderingHandler: nil,
				ciam:                    &ciam.MockCIAMClient{},
			},
			args: args{
				r: &http.Request{
					Body: io.NopCloser(strings.NewReader(`{"fingerprint":"foo"}`)),
				},
				w: &mockWriter{
					Headers: httpHeaders(nil),
				},
			},
			wantStatusCode:  http.StatusUnprocessableEntity,
			wantBody:        []byte(`{"error":"invalid request"}`),
			wantErr:         newInputContentValidationError(errors.New("invalid fingerprint")),
			errorsCollector: &errCollector{},
		},
		{
			name: "CIAM failure",
			fields: fields{
				diagramRenderingHandler: nil,
				ciam: &ciam.MockCIAMClient{
					Err: errors.New("foobar"),
				},
			},
			args: args{
				r: &http.Request{
					Body: io.NopCloser(strings.NewReader(`{"fingerprint":"9468a4a53a2f2fd9ea96db22dc9dd9bb6ce38b7c"}`)),
				},
				w: &mockWriter{
					Headers: httpHeaders(nil),
				},
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       []byte(`{"error":"internal error"}`),
			wantErr: httpHandlerError{
				Msg:      "foobar",
				Type:     errorCIAMSigninAnonym,
				HTTPCode: http.StatusInternalServerError,
			},
			errorsCollector: &errCollector{},
		},
	}
	t.Parallel()
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				h := httpHandler{
					diagramRenderingHandler:   tt.fields.diagramRenderingHandler,
					reportErrorFn:             tt.errorsCollector.Err,
					corsHeaders:               tt.fields.corsHeaders,
					repositoryAPITokens:       tt.fields.repositoryAPITokens,
					repositoryRequestsHistory: tt.fields.repositoryRequestsHistory,
					ciam:                      tt.fields.ciam,
				}
				h.ciamHandlerSigninAnonym(tt.args.w, tt.args.r)

				if tt.args.w.(*mockWriter).StatusCode != tt.wantStatusCode {
					t.Errorf(
						"unexpected status code. want = %d, got = %d", tt.wantStatusCode,
						tt.args.w.(*mockWriter).StatusCode,
					)
				}

				var wantBody []byte
				wantBody = tt.wantBody
				if wantBody == nil {
					var err error
					wantBody, err = tt.fields.ciam.(*ciam.MockCIAMClient).Tokens().Serialize()
					if err != nil {
						panic(err)
					}
				}

				if !reflect.DeepEqual(tt.args.w.(*mockWriter).V, wantBody) {
					t.Error("unexpected response body")
				}

				if !reflect.DeepEqual(tt.wantErr, tt.errorsCollector.V) {
					t.Error("unexpected error collected")
				}
			},
		)
	}
}

func Test_extractLeadingPath(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty string",
			args: args{
				path: "",
			},
			want: "",
		},
		{
			name: "leading slash",
			args: args{
				path: "/",
			},
			want: "/",
		},
		{
			name: "path starting with slash",
			args: args{
				path: "/foo/bar/baz",
			},
			want: "/foo",
		},
		{
			name: "path starting without slash",
			args: args{
				path: "foo/bar/baz",
			},
			want: "foo",
		},
		{
			name: "path with trailing slashes",
			args: args{
				path: "/foo///bar/baz",
			},
			want: "/foo",
		},
		{
			name: "path with trailing slashes in the beginning",
			args: args{
				path: "//foo///bar/baz",
			},
			want: "/",
		},
		{
			name: "path as a single 'word'",
			args: args{
				path: "foobarbaz",
			},
			want: "foobarbaz",
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := extractLeadingPath(tt.args.path); got != tt.want {
					t.Errorf("extractLeadingPath() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}
