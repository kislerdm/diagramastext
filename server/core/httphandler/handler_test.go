package httphandler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/kislerdm/diagramastext/server/core/ciam"
	"github.com/kislerdm/diagramastext/server/core/diagram"
	"github.com/kislerdm/diagramastext/server/core/diagram/c4container"
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

const mockDiagram = `<?xml version="1.0" encoding="us-ascii" standalone="no"?>
<svg xmlns="http://www.w3.org/2000/svg" contentstyletype="text/css" height="179px" preserveAspectRatio="none" version="1.1" viewBox="0 0 375 179" width="375px" zoomAndPan="magnify">
<defs></defs>
<g>
	<g id="elem_n0">
		<rect fill="#438DD5" height="52.5938" rx="2.5" ry="2.5" style="stroke:#3C7FC0;stroke-width:0.5;" width="125" x="7" y="11.8301"></rect>
		<text fill="#FFFFFF" font-family="sans-serif" font-size="16" font-weight="bold" lengthAdjust="spacing" textLength="40" x="17" y="36.6816">Web
		</text>
		<text fill="#FFFFFF" font-family="sans-serif" font-size="16" font-weight="bold" lengthAdjust="spacing" textLength="6" x="57" y="36.6816"></text>
		<text fill="#FFFFFF" font-family="sans-serif" font-size="16" font-weight="bold" lengthAdjust="spacing" textLength="59" x="63" y="36.6816">Server
		</text>
		<text fill="#FFFFFF" font-family="sans-serif" font-size="12" font-style="italic" lengthAdjust="spacing" textLength="26" x="56.5" y="51.5938">[Go]
		</text>
	</g>
	<g id="elem_n1">
		<path d="M251,17.3301 C251,7.3301 304.5,7.3301 304.5,7.3301 C304.5,7.3301 358,7.3301 358,17.3301 L358,58.9238 C358,68.9238 304.5,68.9238 304.5,68.9238 C304.5,68.9238 251,68.9238 251,58.9238 L251,17.3301 " fill="#B3B3B3" style="stroke:#A6A6A6;stroke-width:0.5;"></path>
		<path d="M251,17.3301 C251,27.3301 304.5,27.3301 304.5,27.3301 C304.5,27.3301 358,27.3301 358,17.3301 " fill="none" style="stroke:#A6A6A6;stroke-width:0.5;"></path>
		<text fill="#FFFFFF" font-family="sans-serif" font-size="16" font-weight="bold" lengthAdjust="spacing" textLength="87" x="261" y="46.1816">Database
		</text>
		<text fill="#FFFFFF" font-family="sans-serif" font-size="12" font-style="italic" lengthAdjust="spacing" textLength="61" x="274" y="61.0938">[Postgres]
		</text>
	</g>
</g>
</svg>`

func TestE2e(t *testing.T) {
	t.Run(
		"anonym user behaviour", func(t *testing.T) {
			t.Parallel()

			t.Run(
				"shall issue tokens, and generate diagram using acc token", func(t *testing.T) {
					// GIVEN

					// CIAM http handler
					key := ciam.GenerateCertificate()
					mockCIAMRepo := &ciam.MockRepositoryCIAM{}
					mockSMTP := &ciam.MockSMTPClient{}

					handlerCIAM, err := ciam.HTTPHandler(mockCIAMRepo, mockSMTP, key)
					if err != nil {
						t.Fatal(err)
					}

					// diagram's http handler
					diagramHandler, err := c4container.NewC4ContainersHTTPHandler(
						&diagram.MockModelInference{V: []byte(`{"nodes":[{"id":"0"}]}`)},
						&diagram.MockRepositoryPrediction{},
						diagram.MockHTTPClient{
							V: &http.Response{
								StatusCode:    http.StatusOK,
								Body:          io.NopCloser(bytes.NewReader([]byte(mockDiagram))),
								ContentLength: 100,
							},
						},
					)
					if err != nil {
						t.Fatal(err)
					}

					// CORS headers
					corsHeadersMap := map[string]string{
						"Access-Control-Allow-Origin":  "https://diagramastext.dev",
						"Access-Control-Allow-Headers": "Content-Type,X-Amz-Date,x-api-key,Authorization,X-Api-Key,X-Amz-Security-Token",
						"Access-Control-Allow-Methods": "POST,OPTIONS,GET",
					}
					corsHeaders := http.Header{}
					for k, v := range corsHeadersMap {
						corsHeaders.Add(k, v)
					}

					handler := NewHandler(
						handlerCIAM, corsHeadersMap,
						map[string]diagram.HTTPHandler{
							"/c4": diagramHandler,
						},
					)

					// WHEN

					// authenticate anonym user
					w := &mockWriter{
						Headers: http.Header{},
					}
					r := &http.Request{
						Method: http.MethodPost,
						URL:    &url.URL{Path: "/auth/anonym"},
						Body:   io.NopCloser(bytes.NewReader([]byte(`{"fingerprint":"9468a4a53a2f2fd9ea96db22dc9dd9bb6ce38b71"}`))),
					}

					handler.ServeHTTP(w, r)

					// THEN

					// JWT is returned
					if w.StatusCode != http.StatusOK {
						t.Error("unexpected status code, 200 is expected")
					}
					const wantResponseMimeType = "application/json"
					if w.Headers.Get("Content-Type") != wantResponseMimeType {
						t.Errorf("content type is expected to be %s", wantResponseMimeType)
					}

					var accTkn struct {
						Acc string `json:"access"`
					}
					if err := json.Unmarshal(w.V, &accTkn); err != nil {
						t.Fatal(err)
					}

					// WHEN

					// diagram is generated

					w = &mockWriter{
						Headers: http.Header{},
					}

					header := http.Header{}
					header.Add("Authorization", "Bearer "+accTkn.Acc)

					r = &http.Request{
						Method: http.MethodPost,
						URL:    &url.URL{Path: "/generate/c4"},
						Header: header,
						Body:   io.NopCloser(bytes.NewReader([]byte(`{"prompt":"foo bar qux"}`))),
					}

					handler.ServeHTTP(w, r)
					if w.StatusCode != http.StatusOK {
						t.Error("unexpected status code, 200 is expected")
					}

					if w.Headers.Get("Content-Type") != wantResponseMimeType {
						t.Errorf("content type is expected to be %s", wantResponseMimeType)
					}

					var o struct {
						SVG string `json:"svg"`
					}
					if err := json.Unmarshal(w.V, &o); err != nil {
						t.Fatal(err)
					}

					if o.SVG == "" {
						t.Error("empty SVG returned")
					}

				},
			)
		},
	)
}
