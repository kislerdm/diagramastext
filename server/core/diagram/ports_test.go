package diagram

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestMockModelInference(t *testing.T) {
	t.Parallel()

	t.Run(
		"happy path", func(t *testing.T) {
			// GIVEN
			want := []byte(`{"foo":"bar"}`)
			c := MockModelInference{
				V: want,
			}

			// WHEN
			got, err := c.Do(context.TODO(), "foobarbaz", "model-foo", 2)

			// THEN
			if err != nil {
				t.Error("unexpected error")
				return
			}
			if !reflect.DeepEqual(got, want) {
				t.Error("unexpected output")
				return
			}
		},
	)

	t.Run(
		"unhappy path", func(t *testing.T) {
			// GIVEN
			wantErr := errors.New("foobar")
			c := MockModelInference{Err: wantErr}

			// WHEN
			got, err := c.Do(context.TODO(), "foobarbaz", "model-foo", 2)

			// THEN
			if !reflect.DeepEqual(err, wantErr) {
				t.Error("unexpected error")
				return
			}
			if got != nil {
				t.Error("unexpected output")
				return
			}
		},
	)
}

func TestMockRepositoryPrediction(t *testing.T) {
	t.Parallel()

	t.Run(
		"happy path", func(t *testing.T) {
			c := MockRepositoryPrediction{}
			const (
				requestID  = "foobar"
				userID     = "BA"
				prompt     = "foobar"
				prediction = "bazqux"
			)

			if err := c.WriteInputPrompt(context.TODO(), requestID, userID, prompt); err != nil {
				t.Error("unexpected error when execute WriteInputPrompt")
				return
			}
			if err := c.WriteModelResult(context.TODO(), requestID, userID, prediction); err != nil {
				t.Error("unexpected error when execute WriteModelResult")
				return
			}
			if err := c.Close(context.TODO()); err != nil {
				t.Error("unexpected error when execute Close")
				return
			}
		},
	)

	t.Run(
		"unhappy path", func(t *testing.T) {
			wantErr := errors.New("foobar")

			c := MockRepositoryPrediction{
				Err: wantErr,
			}

			const (
				requestID  = "foobar"
				userID     = "BA"
				prompt     = "foobar"
				prediction = "bazqux"
			)

			if err := c.WriteInputPrompt(context.TODO(), requestID, userID, prompt); !reflect.DeepEqual(err, wantErr) {
				t.Error("unexpected error when execute WriteInputPrompt")
				return
			}
			if err := c.WriteModelResult(context.TODO(), requestID, userID, prediction); !reflect.DeepEqual(
				err, wantErr,
			) {
				t.Error("unexpected error when execute WriteModelResult")
				return
			}
			if err := c.Close(context.TODO()); !reflect.DeepEqual(err, wantErr) {
				t.Error("unexpected error when execute Close")
				return
			}
		},
	)
}

func TestMockRepositorySecretsVault(t *testing.T) {
	t.Parallel()

	t.Run(
		"happy path", func(t *testing.T) {
			// GIVEN
			type secret struct {
				Foo string `json:"foo"`
			}

			c := MockRepositorySecretsVault{
				V: []byte(`{"foo":"bar"}`),
			}

			// WHEN
			var s secret
			err := c.ReadLastVersion(context.TODO(), randomString(10), &s)

			// THEN
			if err != nil {
				t.Error("unexpected error")
				return
			}

			if !reflect.DeepEqual(s, secret{Foo: "bar"}) {
				t.Error("unexpected result")
				return
			}
		},
	)

	t.Run(
		"unhappy path", func(t *testing.T) {
			// GIVEN
			c := MockRepositorySecretsVault{
				Err: errors.New("foobar"),
			}

			// WHEN
			var s map[string]string
			err := c.ReadLastVersion(context.TODO(), randomString(10), &s)

			// THEN
			if err == nil {
				t.Error("unexpected error")
				return
			}

			if s != nil {
				t.Error("unexpected result")
				return
			}
		},
	)
}

func TestMockHTTPClient(t *testing.T) {
	t.Parallel()

	t.Run(
		"happy path", func(t *testing.T) {
			// GIVEN
			want := &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"foo":"bar"}`)),
			}

			c := MockHTTPClient{
				V: want,
			}

			// WHEN
			got, err := c.Do(
				&http.Request{
					Method: http.MethodGet,
					URL: &url.URL{
						Scheme: "https",
						Host:   "localhost",
						Path:   "foo",
					},
				},
			)

			// THEN
			if err != nil {
				t.Error("unexpected error")
				return
			}

			if !reflect.DeepEqual(got, want) {
				t.Error("unexpected result")
				return
			}
		},
	)

	t.Run(
		"unhappy path", func(t *testing.T) {
			// GIVEN
			wantErr := errors.New("foobar")
			c := MockHTTPClient{
				Err: wantErr,
			}

			// WHEN
			got, err := c.Do(
				&http.Request{
					Method: http.MethodGet,
					URL: &url.URL{
						Scheme: "https",
						Host:   "localhost",
						Path:   "foo",
					},
				},
			)

			// THEN
			if !reflect.DeepEqual(err, wantErr) {
				t.Error("unexpected error")
				return
			}

			if got != nil {
				t.Error("unexpected result")
				return
			}
		},
	)

}
