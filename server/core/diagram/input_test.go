package diagram

import (
	"errors"
	"io"
	"math/rand"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"
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
		User   *User
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
				User:   &User{},
			},
			wantErr: false,
		},
		{
			name: "base user: unhappy path - too long",
			fields: fields{
				Prompt: randomString(promptLengthMaxBaseUser + 1),
				User:   &User{},
			},
			wantErr: true,
		},
		{
			name: "base user: unhappy path - too short",
			fields: fields{
				Prompt: randomString(promptLengthMin - 1),
				User:   &User{},
			},
			wantErr: true,
		},
		{
			name: "registered user: happy path",
			fields: fields{
				Prompt: randomString(promptLengthMaxRegisteredUser - 1),
				User:   &User{IsRegistered: true},
			},
			wantErr: false,
		},
		{
			name: "registered user: unhappy path -  too long",
			fields: fields{
				Prompt: randomString(promptLengthMaxRegisteredUser + 1),
				User:   &User{IsRegistered: true},
			},
			wantErr: true,
		},
		{
			name: "registered user: unhappy path -  too short",
			fields: fields{
				Prompt: randomString(promptLengthMin - 1),
				User:   &User{IsRegistered: true},
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
		want    Input
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				body: strings.NewReader(`{"prompt":"` + validPrompt + `"}`),
			},
			want: &inquiry{
				Prompt: validPrompt,
				User:   &User{ID: "00000000-0000-0000-0000-000000000000"},
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

func TestMockInput(t *testing.T) {
	t.Parallel()

	t.Run(
		"invalid input", func(t *testing.T) {
			// GIVEN
			input := MockInput{Err: errors.New("foobar")}

			// WHEN
			err := input.Validate()

			// THEN
			if !reflect.DeepEqual(err, errors.New("foobar")) {
				t.Fatalf("unexpected error")
			}
		},
	)

	t.Run(
		"getters", func(t *testing.T) {
			// GIVEN
			wantUser := &User{
				ID: "id",
			}

			const (
				wantPrompt    = "foobarbaz"
				wantRequestID = "bar"
			)

			input := MockInput{
				Prompt:    wantPrompt,
				RequestID: wantRequestID,
				User:      wantUser,
			}

			// WHEN
			err := input.Validate()
			gotUser := input.GetUser()
			gotRequestID := input.GetRequestID()
			gotPrompt := input.GetPrompt()

			// THEN
			if err != nil {
				t.Fatalf("unexpected error")
			}
			if !reflect.DeepEqual(wantUser, gotUser) {
				t.Fatalf("unexpected user attr")
			}
			if wantPrompt != gotPrompt {
				t.Fatalf("unexpected prompt")
			}
			if wantRequestID != gotRequestID {
				t.Fatalf("unexpected requestID")
			}
		},
	)
}
