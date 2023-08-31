package ciam

import (
	"bytes"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/kislerdm/diagramastext/server/core/internal/utils"
)

func TestServeHTTP(t *testing.T) {
	t.Parallel()

	t.Run(
		"CIAM: authN operations", func(t *testing.T) {
			t.Parallel()

			var init = func(t *testing.T) (
				http.Handler, *MockWriter, ed25519.PrivateKey,
			) {
				clientRepo := &MockRepositoryCIAM{}
				smtpClient := &MockSMTPClient{}
				key := GenerateCertificate()

				h, err := HTTPHandler(clientRepo, smtpClient, key)
				if err != nil {
					t.Fatal(err)
				}

				w := &MockWriter{}
				return h(nil), w, key
			}

			t.Run(
				"shall fail on GET request method for the token issue handlers", func(t *testing.T) {
					// GIVEN
					handler, writer, _ := init(t)

					request := &http.Request{
						Method: http.MethodGet,
						URL: &url.URL{
							Path: "/auth/foo",
						},
					}

					// WHEN
					handler.ServeHTTP(writer, request)

					// THEN
					wantStatus := http.StatusMethodNotAllowed
					wantBody := []byte(`{"error":"GET is not allowed"}`)

					if writer.StatusCode != wantStatus {
						t.Errorf("wrong status code. want: %d, got: %d", wantStatus, writer.StatusCode)
					}
					if !reflect.DeepEqual(writer.V, wantBody) {
						t.Errorf("wrong response content. want: %v, got: %v", wantBody, writer.V)
					}
				},
			)

			t.Run(
				"shall successfully process signin anonym request", func(t *testing.T) {
					// GIVEN
					handler, writer, key := init(t)
					iss, err := NewIssuer(key)
					if err != nil {
						t.Fatal(err)
					}

					request := &http.Request{
						Method: http.MethodPost,
						URL: &url.URL{
							Path: "/auth/anonym",
						},
						Body: io.NopCloser(
							bytes.NewReader(
								[]byte(
									`{"fingerprint":"9468a4a53a2f2fd9ea96db22dc9dd9bb6ce38b71"}`,
								),
							),
						),
					}

					// WHEN
					handler.ServeHTTP(writer, request)

					// THEN
					wantStatus := http.StatusOK
					var wantBodyValid = func(v []byte) {
						var o struct {
							ID  string `json:"id"`
							Acc string `json:"access"`
							Ref string `json:"refresh"`
						}
						if err := json.Unmarshal(v, &o); err != nil {
							t.Fatal(err)
						}

						if _, _, _, err := iss.ParseIDToken(o.ID); err != nil {
							t.Errorf("faulty ID token: %v", err)
						}
						if _, err := iss.ParseAccessToken(o.Acc); err != nil {
							t.Errorf("faulty Access token: %v", err)
						}
						if _, err := iss.ParseRefreshToken(o.Ref); err != nil {
							t.Errorf("faulty Refresh token: %v", err)
						}
					}

					if writer.StatusCode != wantStatus {
						t.Errorf("wrong status code. want: %d, got: %d", wantStatus, writer.StatusCode)
					}

					wantBodyValid(writer.V)
				},
			)

			t.Run(
				"shall successfully perform user signin init request", func(t *testing.T) {
					// GIVEN
					handler, writer, key := init(t)
					iss, err := NewIssuer(key)
					if err != nil {
						t.Fatal(err)
					}

					request := &http.Request{
						Method: http.MethodPost,
						URL: &url.URL{
							Path: "/auth/init",
						},
						Body: io.NopCloser(
							bytes.NewReader(
								[]byte(
									`{"email":"foo@bar.baz"}`,
								),
							),
						),
					}

					// WHEN
					handler.ServeHTTP(writer, request)

					// THEN
					wantStatus := http.StatusOK
					var wantBodyValid = func(v []byte) {
						if _, _, _, err := iss.ParseIDToken(string(v)); err != nil {
							t.Errorf("faulty ID token: %v", err)
						}
					}

					if writer.StatusCode != wantStatus {
						t.Errorf("wrong status code. want: %d, got: %d", wantStatus, writer.StatusCode)
					}

					wantBodyValid(writer.V)
				},
			)

			t.Run(
				"shall successfully perform user signin confirmation request", func(t *testing.T) {
					// GIVEN
					wantUserID := utils.NewUUID()
					const (
						wantSecret = "foobar"
						wantEmail  = "foo@bar.baz"
					)

					smtpClient := &MockSMTPClient{}
					key := GenerateCertificate()

					clientRepo := &MockRepositoryCIAM{
						UserID: map[string]*userContainer{
							wantUserID: {
								ID:     wantUserID,
								Email:  wantEmail,
								RoleID: uint8(RoleAnonymUser),
							},
						},
						Secret: map[string]Secret{
							wantUserID: {
								Secret:   wantSecret,
								IssuedAt: time.Now(),
							},
						},
					}

					handlerFn, err := HTTPHandler(clientRepo, smtpClient, key)
					if err != nil {
						t.Fatal(err)
					}
					handler := handlerFn(nil)

					iss, err := NewIssuer(key)
					if err != nil {
						t.Fatal(err)
					}

					wantIDToken, err := iss.NewIDToken(wantUserID, wantEmail, "")
					if err != nil {
						t.Fatal(err)
					}

					request := &http.Request{
						Method: http.MethodPost,
						URL: &url.URL{
							Path: "/auth/confirm",
						},
						Body: io.NopCloser(
							bytes.NewReader(
								[]byte(
									`{
"secret":"` + wantSecret + `",
"id_token": "` + wantIDToken + `"
}`,
								),
							),
						),
					}

					writer := &MockWriter{}

					// WHEN
					handler.ServeHTTP(writer, request)

					// THEN
					wantStatus := http.StatusOK
					var wantBodyValid = func(v []byte) {
						var o struct {
							ID  string `json:"id"`
							Acc string `json:"access"`
							Ref string `json:"refresh"`
						}
						if err := json.Unmarshal(v, &o); err != nil {
							t.Fatal(err)
						}

						if _, _, _, err := iss.ParseIDToken(o.ID); err != nil {
							t.Errorf("faulty ID token: %v", err)
						}
						user, err := iss.ParseAccessToken(o.Acc)
						if err != nil {
							t.Errorf("faulty Access token: %v", err)
						}

						if user.ID != wantUserID {
							t.Errorf(
								"faulty Access token, userID does not match. want: %s, got: %s",
								wantUserID, user.ID,
							)
						}

						if user.Role != RoleRegisteredUser {
							t.Errorf(
								"faulty Access token, user's role does not match. want: %d, got: %d",
								RoleRegisteredUser, user.Role,
							)
						}

						if _, err := iss.ParseRefreshToken(o.Ref); err != nil {
							t.Errorf("faulty Refresh token: %v", err)
						}

					}

					if writer.StatusCode != wantStatus {
						t.Errorf("wrong status code. want: %d, got: %d", wantStatus, writer.StatusCode)
					}

					wantBodyValid(writer.V)
				},
			)

			t.Run(
				"shall successfully refresh id and acc tokens given refresh token", func(t *testing.T) {
					// GIVEN
					wantUserID := utils.NewUUID()
					const wantEmail = "foo@bar.baz"

					clientRepo := &MockRepositoryCIAM{
						UserID: map[string]*userContainer{
							wantUserID: {
								ID:       wantUserID,
								Email:    wantEmail,
								IsActive: true,
								RoleID:   uint8(RoleRegisteredUser),
							},
						},
					}

					smtpClient := &MockSMTPClient{}
					key := GenerateCertificate()

					handlerFn, err := HTTPHandler(clientRepo, smtpClient, key)
					if err != nil {
						t.Fatal(err)
					}
					handler := handlerFn(nil)

					iss, err := NewIssuer(key)
					if err != nil {
						t.Fatal(err)
					}

					refToken, err := iss.NewRefreshToken(wantUserID)
					if err != nil {
						t.Fatal(err)
					}

					request := &http.Request{
						Method: http.MethodPost,
						URL: &url.URL{
							Path: "/auth/refresh",
						},
						Body: io.NopCloser(
							bytes.NewReader(
								[]byte(
									`{"refresh_token":"` + refToken + `"}`,
								),
							),
						),
					}

					writer := &MockWriter{}

					// WHEN
					handler.ServeHTTP(writer, request)

					// THEN
					wantStatus := http.StatusOK
					var wantBodyValid = func(v []byte) {
						var o struct {
							ID  string `json:"id"`
							Acc string `json:"access"`
						}
						if err := json.Unmarshal(v, &o); err != nil {
							t.Fatal(err)
						}

						if _, _, _, err := iss.ParseIDToken(o.ID); err != nil {
							t.Errorf("faulty ID token: %v", err)
						}
						user, err := iss.ParseAccessToken(o.Acc)
						if err != nil {
							t.Errorf("faulty Access token: %v", err)
						}

						if user.ID != wantUserID {
							t.Errorf(
								"faulty Access token, userID does not match. want: %s, got: %s",
								wantUserID, user.ID,
							)
						}

						if user.Role != RoleRegisteredUser {
							t.Errorf(
								"faulty Access token, user's role does not match. want: %d, got: %d",
								RoleRegisteredUser, user.Role,
							)
						}
					}

					if writer.StatusCode != wantStatus {
						t.Errorf("wrong status code. want: %d, got: %d", wantStatus, writer.StatusCode)
					}

					wantBodyValid(writer.V)
				},
			)
		},
	)

	t.Run(
		"API calls", func(t *testing.T) {
			t.Parallel()

			var initApiCallByRegisteredUser = func() (RepositoryCIAM, http.Header, string) {
				userID := utils.NewUUID()
				const (
					email  = "foo@bar.baz"
					apiKey = "foo"
				)

				header := http.Header{}
				header.Add("X-API-KEY", apiKey)

				clientRepo := &MockRepositoryCIAM{
					UserToken: map[string]string{
						apiKey: userID,
					},
					UserID: map[string]*userContainer{
						userID: {
							ID:       userID,
							Email:    email,
							IsActive: true,
							RoleID:   uint8(RoleRegisteredUser),
						},
					},
				}

				return clientRepo, header, userID
			}

			t.Run(
				"shall return quotas upon an API call given a valid API-KEY, the used has not used the service yet",
				func(t *testing.T) {
					// GIVEN
					clientRepo, header, _ := initApiCallByRegisteredUser()

					handlerFn, err := HTTPHandler(clientRepo, &MockSMTPClient{}, GenerateCertificate())
					if err != nil {
						t.Fatal(err)
					}

					handler := handlerFn(nil)

					request := &http.Request{
						Method: http.MethodGet,
						URL: &url.URL{
							Path: "/quotas",
						},
						Header: header,
					}

					writer := &MockWriter{}

					// WHEN
					handler.ServeHTTP(writer, request)

					// THEN
					wantStatusCode := http.StatusOK
					if writer.StatusCode != wantStatusCode {
						t.Errorf("unexpected status code. want: %d, got: %d", wantStatusCode, writer.StatusCode)
					}

					var gotBody QuotasUsage
					if err := json.Unmarshal(writer.V, &gotBody); err != nil {
						t.Fatal(err)
					}

					if gotBody.PromptLengthMax != RoleRegisteredUser.Quotas().PromptLengthMax {
						t.Error("unexpected PromptLengthMax")
					}

					if gotBody.RateDay.Limit != RoleRegisteredUser.Quotas().RequestsPerDay {
						t.Error("unexpected RequestsPerDay limit")
					}

					if gotBody.RateDay.Used != 0 {
						t.Error("unexpected RequestsPerDay used")
					}

					if gotBody.RateMinute.Limit != RoleRegisteredUser.Quotas().RequestsPerMinute {
						t.Error("unexpected RateMinute limit")
					}

					if gotBody.RateMinute.Used != 0 {
						t.Error("unexpected RateMinute used")
					}
				},
			)

			t.Run(
				"shall process API call by the next handler given a valid API-KEY, the used has not used the service yet",
				func(t *testing.T) {
					// GIVEN
					clientRepo, header, userID := initApiCallByRegisteredUser()

					handlerFn, err := HTTPHandler(clientRepo, &MockSMTPClient{}, GenerateCertificate())
					if err != nil {
						t.Fatal(err)
					}

					handler := handlerFn(mockHandlerAPIcall{userID: userID})

					request := &http.Request{
						Method: http.MethodPost,
						URL: &url.URL{
							Path: "/foo",
						},
						Header: header,
					}

					writer := &MockWriter{}

					// WHEN
					handler.ServeHTTP(writer, request)

					// THEN
					wantStatusCode := http.StatusOK
					if writer.StatusCode != wantStatusCode {
						t.Errorf("unexpected status code. want: %d, got: %d", wantStatusCode, writer.StatusCode)
					}
				},
			)

			t.Run(
				"shall return throttling error on API calls given a valid API-KEY", func(t *testing.T) {
					// GIVEN
					clientRepo, header, _ := initApiCallByRegisteredUser()
					clientRepo.(*MockRepositoryCIAM).Timestamps = repeatTimestamp(
						time.Now(), RoleRegisteredUser.Quotas().RequestsPerMinute+1,
					)

					handlerFn, err := HTTPHandler(clientRepo, &MockSMTPClient{}, GenerateCertificate())
					if err != nil {
						t.Fatal(err)
					}

					handler := handlerFn(nil)

					request := &http.Request{
						Method: http.MethodPost,
						URL: &url.URL{
							Path: "/foo",
						},
						Header: header,
					}

					writer := &MockWriter{}

					// WHEN
					handler.ServeHTTP(writer, request)

					// THEN
					wantStatusCode := http.StatusTooManyRequests
					if writer.StatusCode != wantStatusCode {
						t.Errorf("unexpected status code. want: %d, got: %d", wantStatusCode, writer.StatusCode)
					}
				},
			)

			t.Run(
				"shall shall return access forbidden on no token", func(t *testing.T) {
					// GIVEN
					clientRepo, _, _ := initApiCallByRegisteredUser()
					handlerFn, err := HTTPHandler(clientRepo, &MockSMTPClient{}, GenerateCertificate())
					if err != nil {
						t.Fatal(err)
					}

					handler := handlerFn(nil)

					request := &http.Request{
						Method: http.MethodPost,
						URL: &url.URL{
							Path: "/foo",
						},
					}

					writer := &MockWriter{}

					// WHEN
					handler.ServeHTTP(writer, request)

					// THEN
					wantStatusCode := http.StatusForbidden
					if writer.StatusCode != wantStatusCode {
						t.Errorf("unexpected status code. want: %d, got: %d", wantStatusCode, writer.StatusCode)
					}
					if string(writer.V) != `{"error":"no authentication token provided"}` {
						t.Errorf("unexpected response body")
					}
				},
			)

			t.Run(
				"shall shall return internal error upon checking API token", func(t *testing.T) {
					// GIVEN
					clientRepo, header, _ := initApiCallByRegisteredUser()
					clientRepo.(*MockRepositoryCIAM).Err = errors.New("foo")

					handlerFn, err := HTTPHandler(clientRepo, &MockSMTPClient{}, GenerateCertificate())
					if err != nil {
						t.Fatal(err)
					}

					handler := handlerFn(nil)

					request := &http.Request{
						Method: http.MethodPost,
						URL: &url.URL{
							Path: "/foo",
						},
						Header: header,
					}

					writer := &MockWriter{}

					// WHEN
					handler.ServeHTTP(writer, request)

					// THEN
					wantStatusCode := http.StatusInternalServerError
					if writer.StatusCode != wantStatusCode {
						t.Errorf("unexpected status code. want: %d, got: %d", wantStatusCode, writer.StatusCode)
					}
					if string(writer.V) != `{"error":"internal error"}` {
						t.Errorf("unexpected response body")
					}
				},
			)
		},
	)
}

type mockHandlerAPIcall struct {
	userID string
}

func (m mockHandlerAPIcall) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user, ok := FromContext(r.Context())
	if !ok || user.ID != m.userID {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}
