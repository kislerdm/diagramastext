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

	var httpHeaders = func(h map[string]string) http.Header {
		o := http.Header{}
		o.Add("Content-Type", "application/json")
		for k, v := range h {
			o.Add(k, v)
		}
		return o
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

func Test_httpHandler_ServeHTTPDiagramRenderingHappyPath(t *testing.T) {
	corsHeaders := map[string]string{
		"Access-Control-Allow-Origin": "https://diagramastext.dev",
	}

	diagramRenderingHandler := map[string]diagram.HTTPHandler{
		"/c4": func(_ context.Context, _ diagram.Input) (diagram.Output, error) {
			return diagram.MockOutput{
				V: []byte(`{"svg":"foo"}`),
			}, nil
		},
	}

	type args struct {
		w          http.ResponseWriter
		path       string
		authHeader string
	}
	tests := []struct {
		name            string
		args            args
		errorsCollector *errCollector
		wantW           http.ResponseWriter
		wantErr         error
	}{
		{

			name: "webclient",
			args: args{
				w: &mockWriter{
					Headers: http.Header{},
				},
				path:       "/internal/generate/c4",
				authHeader: "foobar",
			},
			wantW: &mockWriter{
				Headers:    httpHeaders(corsHeaders),
				StatusCode: http.StatusOK,
				V:          []byte(`{"svg":"foo"}`),
			},
			errorsCollector: &errCollector{},
		},
		{

			name: "api",
			args: args{
				w: &mockWriter{
					Headers: http.Header{},
				},
				path:       "/generate/c4",
				authHeader: "Bearer 1410904f-f646-488f-ae08-cc341dfb321c",
			},
			wantW: &mockWriter{
				Headers:    httpHeaders(corsHeaders),
				StatusCode: http.StatusNotImplemented,
				V:          []byte(`{"error":"The Rest API will be supported in the upcoming release v0.0.6"}`),
			},
			errorsCollector: &errCollector{},
			wantErr: httpHandlerError{
				Msg:      "Rest API call attempt",
				Type:     "NotImplemented",
				HTTPCode: http.StatusNotImplemented,
			},
		},
	}

	t.Parallel()

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				h := httpHandler{
					diagramRenderingHandler: diagramRenderingHandler,
					reportErrorFn:           tt.errorsCollector.Err,
					corsHeaders:             corsHeaders,
				}
				h.ServeHTTP(
					tt.args.w, &http.Request{
						Method: http.MethodPost,
						URL: &url.URL{
							Path: tt.args.path,
						},
						Header: httpHeaders(
							map[string]string{
								"Authorization": tt.args.authHeader,
							},
						),
						Body: io.NopCloser(strings.NewReader(`{"prompt":"foobar"}`)),
					},
				)

				if tt.args.w.(*mockWriter).StatusCode != tt.wantW.(*mockWriter).StatusCode {
					t.Errorf("unexpected response status code")
					return
				}

				if !reflect.DeepEqual(tt.args.w.Header(), tt.wantW.(*mockWriter).Header()) {
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
				h.diagramRendering(tt.args.w, tt.args.r)

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
