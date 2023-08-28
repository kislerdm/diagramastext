package diagram

import (
	"errors"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func randomString(length uint16) string {
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
		Prompt          string
		PromptLengthMax uint16
	}

	const promptLengthMax = 100

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				Prompt:          randomString(promptLengthMax - 1),
				PromptLengthMax: promptLengthMax,
			},
			wantErr: false,
		},
		{
			name: "unhappy path - too long",
			fields: fields{
				Prompt:          randomString(promptLengthMax + 1),
				PromptLengthMax: promptLengthMax,
			},
			wantErr: true,
		},
		{
			name: "unhappy path - too short",
			fields: fields{
				Prompt:          randomString(3 - 1),
				PromptLengthMax: promptLengthMax,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				v := inquiry{
					Prompt:          tt.fields.Prompt,
					PromptLengthMax: tt.fields.PromptLengthMax,
				}
				if err := v.Validate(); (err != nil) != tt.wantErr {
					t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				}
			},
		)
	}
}

func TestNewInput(t *testing.T) {
	type args struct {
		prompt          string
		userID          string
		apiToken        string
		promptLengthMax uint16
	}

	const promptLengthMax = 100

	validPrompt := randomString(promptLengthMax - 1)

	tests := []struct {
		name    string
		args    args
		want    Input
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				prompt:          validPrompt,
				userID:          "00000000-0000-0000-0000-000000000000",
				promptLengthMax: promptLengthMax,
				apiToken:        "foobar",
			},
			want: &inquiry{
				Prompt:   validPrompt,
				UserID:   "00000000-0000-0000-0000-000000000000",
				APIToken: "foobar",
			},
			wantErr: false,
		},
		{
			name: "unhappy path: invalid prompt",
			args: args{
				prompt:          randomString(promptLengthMin - 1),
				userID:          "00000000-0000-0000-0000-000000000000",
				promptLengthMax: promptLengthMax,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := NewInput(tt.args.prompt, tt.args.userID, tt.args.apiToken, tt.args.promptLengthMax)
				if (err != nil) != tt.wantErr {
					t.Errorf("NewInputDriverHTTP() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				if err == nil {
					if !reflect.DeepEqual(got.GetUserID(), tt.want.GetUserID()) {
						t.Errorf("NewInputDriverHTTP() unexpected userID: got = %v, want %v", got, tt.want)
					}

					if !reflect.DeepEqual(got.GetPrompt(), tt.want.GetPrompt()) {
						t.Errorf("NewInputDriverHTTP() unexpected prompt: got = %v, want %v", got, tt.want)
					}

					if !reflect.DeepEqual(got.GetUserAPIToken(), tt.want.GetUserAPIToken()) {
						t.Errorf("NewInputDriverHTTP() unexpected userAPIToken: got = %v, want %v", got, tt.want)
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
			const (
				wantUserID    = "id"
				wantPrompt    = "foobarbaz"
				wantRequestID = "bar"
			)

			input := MockInput{
				Prompt:    wantPrompt,
				RequestID: wantRequestID,
				UserID:    wantUserID,
			}

			// WHEN
			err := input.Validate()
			gotUserID := input.GetUserID()
			gotRequestID := input.GetRequestID()
			gotPrompt := input.GetPrompt()

			// THEN
			if err != nil {
				t.Fatalf("unexpected error")
			}

			if !reflect.DeepEqual(wantUserID, gotUserID) {
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
