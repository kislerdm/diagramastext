package adapter

import (
	"io"
	"math/rand"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/kislerdm/diagramastext/server/core/port"
)

func randomString(length int) string {
	const charset = "abcdef"
	var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	var b = make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func Test_inquiry_Validate(t *testing.T) {
	type fields struct {
		Prompt string
		User   *port.User
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "base user: happy path",
			fields: fields{
				Prompt: randomString(promptLengthMaxBaseUser - 1),
				User:   &port.User{},
			},
			wantErr: false,
		},
		{
			name: "base user: unhappy path - too long",
			fields: fields{
				Prompt: randomString(promptLengthMaxBaseUser + 1),
				User:   &port.User{},
			},
			wantErr: true,
		},
		{
			name: "base user: unhappy path - too short",
			fields: fields{
				Prompt: randomString(promptLengthMin - 1),
				User:   &port.User{},
			},
			wantErr: true,
		},
		{
			name: "registered user: happy path",
			fields: fields{
				Prompt: randomString(promptLengthMaxRegisteredUser - 1),
				User:   &port.User{IsRegistered: true},
			},
			wantErr: false,
		},
		{
			name: "registered user: unhappy path -  too long",
			fields: fields{
				Prompt: randomString(promptLengthMaxRegisteredUser + 1),
				User:   &port.User{IsRegistered: true},
			},
			wantErr: true,
		},
		{
			name: "registered user: unhappy path -  too short",
			fields: fields{
				Prompt: randomString(promptLengthMin - 1),
				User:   &port.User{IsRegistered: true},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				v := inquiry{
					Prompt: tt.fields.Prompt,
					User:   tt.fields.User,
				}
				if err := v.Validate(); (err != nil) != tt.wantErr {
					t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				}
			},
		)
	}
}

func TestNewInquiryDriverHTTP(t *testing.T) {
	type args struct {
		body    io.Reader
		headers http.Header
	}

	validPrompt := randomString(promptLengthMaxBaseUser - 1)

	tests := []struct {
		name    string
		args    args
		want    port.Input
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				body: strings.NewReader(`{"prompt":"` + validPrompt + `"}`),
			},
			want: &inquiry{
				Prompt: validPrompt,
				User:   &port.User{ID: "NA"},
			},
			wantErr: false,
		},
		{
			name: "unhappy path: invalid json",
			args: args{
				body: io.NopCloser(
					strings.NewReader(`{"prompt":`),
				),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "unhappy path: invalid prompt",
			args: args{
				body: io.NopCloser(
					strings.NewReader(`{"prompt":"` + randomString(promptLengthMin-1) + `"}`),
				),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := NewInputDriverHTTP(tt.args.body, tt.args.headers)
				if (err != nil) != tt.wantErr {
					t.Errorf("NewInputDriverHTTP() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				if err == nil {
					if !reflect.DeepEqual(got.GetUser(), tt.want.GetUser()) {
						t.Errorf("NewInputDriverHTTP() unexpected user: got = %v, want %v", got, tt.want)
					}

					if !reflect.DeepEqual(got.GetPrompt(), tt.want.GetPrompt()) {
						t.Errorf("NewInputDriverHTTP() unexpected prompt: got = %v, want %v", got, tt.want)
					}
				}
			},
		)
	}
}
