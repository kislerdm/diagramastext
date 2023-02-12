package handler

import (
	"errors"
	"math/rand"
	"net/http"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/kislerdm/diagramastext/core"
)

func TestReadPrompt(t *testing.T) {
	type args struct {
		requestBody []byte
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				requestBody: []byte(`{"prompt":"foobar"}`),
			},
			want:    "foobar",
			wantErr: false,
		},
		{
			name: "unhappy path: base64 encoded",
			args: args{
				requestBody: []byte(`eyJwcm9tcHQiOiJmb29iYXIifQ==`),
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "unhappy path: corrupt data",
			args: args{
				requestBody: []byte(`{"prompt":"foobar"`),
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := ReadPrompt(tt.args.requestBody)
				if (err != nil) != tt.wantErr {
					t.Errorf("ReadPrompt() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("ReadPrompt() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func randomString(length int) string {
	const charset = "abcdef"
	var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	var b = make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func TestValidatePrompt(t *testing.T) {
	type args struct {
		prompt string
	}
	tests := []struct {
		name string
		args args
		want error
	}{
		{
			name: "valid",
			args: args{
				prompt: "c4 diagram with Go backend reading postgres",
			},
			want: nil,
		},
		{
			name: "invalid: short",
			args: args{
				prompt: randomString(PromptLengthMin - 1),
			},
			want: errors.New(
				"prompt length must be between " + strconv.Itoa(PromptLengthMin) + " and " +
					strconv.Itoa(PromptLengthMax) + " characters",
			),
		},
		{
			name: "invalid: long",
			args: args{
				prompt: randomString(PromptLengthMax + 1),
			},
			want: errors.New(
				"prompt length must be between " + strconv.Itoa(PromptLengthMin) + " and " +
					strconv.Itoa(PromptLengthMax) + " characters",
			),
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if err := ValidatePrompt(tt.args.prompt); (err != nil) && errors.Is(err, tt.want) {
					t.Errorf("ValidatePrompt() error = %v, wantErr %v", err, tt.want)
				}
			},
		)
	}
}

func TestParseClientError(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want ResponseError
	}{
		{
			name: "too many requests",
			args: args{
				err: core.Error{
					ServiceResponseStatusCode: http.StatusTooManyRequests,
				},
			},
			want: ResponseError{
				StatusCode: http.StatusTooManyRequests,
				Body:       []byte("service experiences high load, please try later"),
			},
		},
		{
			name: "opanAI failed to predict",
			args: args{
				err: core.Error{
					Service:                   core.ServiceOpenAI,
					ServiceResponseStatusCode: http.StatusInternalServerError,
				},
			},
			want: ResponseError{
				StatusCode: http.StatusInternalServerError,
				Body:       []byte("could not recognise diagram description"),
			},
		},
		{
			name: "plantUML failed to predict",
			args: args{
				err: core.Error{
					Service:                   core.ServiePlantUML,
					ServiceResponseStatusCode: http.StatusInternalServerError,
				},
			},
			want: ResponseError{
				StatusCode: http.StatusInternalServerError,
				Body:       []byte("could not generate diagram using provided description"),
			},
		},
		{
			name: "unknown",
			args: args{
				err: core.Error{},
			},
			want: ResponseError{
				StatusCode: http.StatusInternalServerError,
				Body:       []byte("unknown"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := ParseClientError(tt.args.err); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ParseClientError() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}
