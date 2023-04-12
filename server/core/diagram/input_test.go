package diagram

import (
	"context"
	"errors"
	"math/rand"
	"reflect"
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
				Prompt: randomString(quotaBaseUserPromptLengthMax - 1),
				User:   &User{},
			},
			wantErr: false,
		},
		{
			name: "base user: unhappy path - too long",
			fields: fields{
				Prompt: randomString(quotaBaseUserPromptLengthMax + 1),
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
				Prompt: randomString(quotaRegisteredUserPromptLengthMax - 1),
				User:   &User{IsRegistered: true},
			},
			wantErr: false,
		},
		{
			name: "registered user: unhappy path -  too long",
			fields: fields{
				Prompt: randomString(quotaRegisteredUserPromptLengthMax + 1),
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

func TestNewInput(t *testing.T) {
	type args struct {
		prompt string
		user   *User
	}

	validPrompt := randomString(quotaBaseUserPromptLengthMax - 1)

	tests := []struct {
		name    string
		args    args
		want    Input
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				prompt: validPrompt,
				user:   &User{ID: "00000000-0000-0000-0000-000000000000", IsRegistered: false},
			},
			want: &inquiry{
				Prompt: validPrompt,
				User:   &User{ID: "00000000-0000-0000-0000-000000000000", IsRegistered: false},
			},
			wantErr: false,
		},
		{
			name: "unhappy path: invalid prompt",
			args: args{
				prompt: randomString(promptLengthMin - 1),
				user:   &User{ID: "00000000-0000-0000-0000-000000000000", IsRegistered: false},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := NewInput(tt.args.prompt, tt.args.user)
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

func mustGenerateTimestamps(tsStr ...string) []time.Time {
	o := make([]time.Time, len(tsStr))

	for i, ts := range tsStr {
		t, err := time.Parse(time.RFC3339, ts)
		if err != nil {
			panic(err)
		}
		o[i] = t
	}

	return o
}

func TestValidateRequestsQuotaUsage(t *testing.T) {
	type args struct {
		ctx              context.Context
		clientRepository RepositoryPrediction
		user             *User
	}
	tests := []struct {
		name              string
		args              args
		wantThrottling    bool
		wantQuotaExceeded bool
		wantErr           bool
	}{
		{
			name: "no request made so far: non registered",
			args: args{
				ctx:              context.TODO(),
				clientRepository: MockRepositoryPrediction{},
				user:             &User{},
			},
			wantThrottling:    false,
			wantQuotaExceeded: false,
			wantErr:           false,
		},
		{
			name: "no request made so far: registered user",
			args: args{
				ctx:              context.TODO(),
				clientRepository: MockRepositoryPrediction{},
				user:             &User{IsRegistered: true},
			},
			wantThrottling:    false,
			wantQuotaExceeded: false,
			wantErr:           false,
		},
		{
			name: "throttling quota exceeded",
			args: args{
				ctx: context.TODO(),
				clientRepository: MockRepositoryPrediction{
					Timestamps: mustGenerateTimestamps(
						repeatStr("2023-01-01T00:00:00Z", quotaRegisteredUserRPM+1)...,
					),
				},
				user: &User{IsRegistered: true},
			},
			wantThrottling:    false,
			wantQuotaExceeded: false,
			wantErr:           false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				gotThrottling, gotQuotaExceeded, err := ValidateRequestsQuotaUsage(
					tt.args.ctx, tt.args.clientRepository, tt.args.user,
				)
				if (err != nil) != tt.wantErr {
					t.Errorf("ValidateRequestsQuotaUsage() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if gotThrottling != tt.wantThrottling {
					t.Errorf(
						"ValidateRequestsQuotaUsage() gotThrottling = %v, want %v", gotThrottling, tt.wantThrottling,
					)
				}
				if gotQuotaExceeded != tt.wantQuotaExceeded {
					t.Errorf(
						"ValidateRequestsQuotaUsage() gotQuotaExceeded = %v, want %v", gotQuotaExceeded,
						tt.wantQuotaExceeded,
					)
				}
			},
		)
	}
}

func repeatStr(s string, nElements int) []string {
	o := make([]string, nElements)
	var i int
	for i < nElements {
		o[i] = s
		i++
	}
	return o
}
