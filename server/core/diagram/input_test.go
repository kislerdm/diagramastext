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
					Timestamps: repeatTimestamp(genNowMinute(), quotaRegisteredUserRPM),
				},
				user: &User{IsRegistered: true},
			},
			wantThrottling:    true,
			wantQuotaExceeded: false,
			wantErr:           false,
		},
		{
			name: "daily quota exceeded",
			args: args{
				ctx: context.TODO(),
				clientRepository: MockRepositoryPrediction{
					Timestamps: repeatTimestamp(genNowDate(), quotaRegisteredUserRPD),
				},
				user: &User{IsRegistered: true},
			},
			wantThrottling:    true,
			wantQuotaExceeded: true,
			wantErr:           false,
		},
		{
			name: "unhappy path",
			args: args{
				ctx: context.TODO(),
				clientRepository: MockRepositoryPrediction{
					Err: errors.New("foo"),
				},
				user: &User{IsRegistered: true},
			},
			wantThrottling:    false,
			wantQuotaExceeded: false,
			wantErr:           true,
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

func repeatTimestamp(ts time.Time, nElements int) []time.Time {
	o := make([]time.Time, nElements)
	var i int
	for i < nElements {
		o[i] = ts
		i++
	}
	return o
}

func Test_sliceWithinWindow(t *testing.T) {
	type args struct {
		ts    []time.Time
		tsMin time.Time
		tsMax time.Time
	}
	tests := []struct {
		name string
		args args
		want []time.Time
	}{
		{
			name: "non-empty slice",
			args: args{
				ts: mustGenerateTimestamps(
					"2023-01-01T00:00:00Z", "2023-01-01T10:00:00Z", "2023-01-01T11:00:00Z", "2023-01-02T00:00:00Z",
					"2023-01-04T00:00:00Z",
				),
				tsMin: mustGenerateTimestamps("2023-01-01T00:00:00Z")[0],
				tsMax: mustGenerateTimestamps("2023-01-02T00:00:00Z")[0],
			},
			want: mustGenerateTimestamps(
				"2023-01-01T00:00:00Z", "2023-01-01T10:00:00Z", "2023-01-01T11:00:00Z", "2023-01-02T00:00:00Z",
			),
		},
		{
			name: "input empty slice",
			args: args{
				ts:    nil,
				tsMin: mustGenerateTimestamps("2023-01-01T00:00:00Z")[0],
				tsMax: mustGenerateTimestamps("2023-01-02T00:00:00Z")[0],
			},
			want: nil,
		},
		{
			name: "non-empty input, empty output",
			args: args{
				ts: mustGenerateTimestamps(
					"2023-01-01T00:00:00Z", "2023-01-01T01:00:00Z", "2023-01-01T02:00:00Z",
				),
				tsMin: mustGenerateTimestamps("2023-01-02T00:00:00Z")[0],
				tsMax: mustGenerateTimestamps("2023-01-03T00:00:00Z")[0],
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := sliceWithinWindow(tt.args.ts, tt.args.tsMin, tt.args.tsMax); !reflect.DeepEqual(
					got, tt.want,
				) {
					t.Errorf("sliceWithinWindow() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

var quotasController = newQuotaIssuer()

func TestGetQuotaUsageBaseUser(t *testing.T) {
	type args struct {
		ctx              context.Context
		clientRepository RepositoryPrediction
	}
	user := &User{}
	tests := []struct {
		name    string
		args    args
		want    QuotasUsage
		wantErr bool
	}{
		{
			name: "no previous requests",
			args: args{
				ctx:              context.TODO(),
				clientRepository: MockRepositoryPrediction{},
			},
			want:    quotasController.quotaUsage(user),
			wantErr: false,
		},
		{
			name: "a single requests",
			args: args{
				ctx: context.TODO(),
				clientRepository: MockRepositoryPrediction{
					Timestamps: repeatTimestamp(quotasController.minuteNow, 1),
				},
			},
			want: QuotasUsage{
				PromptLengthMax: quotaPromptLengthMax(user),
				RateMinute: QuotaRequestsConsumption{
					Limit: quotaBaseUserRPM,
					Used:  1,
					Reset: quotasController.minuteNext.Unix(),
				},
				RateDay: QuotaRequestsConsumption{
					Limit: quotaBaseUserRPD,
					Used:  1,
					Reset: quotasController.dayNext.Unix(),
				},
			},
			wantErr: false,
		},
		{
			name: "daily quota exceeded",
			args: args{
				ctx: context.TODO(),
				clientRepository: MockRepositoryPrediction{
					Timestamps: repeatTimestamp(quotasController.minuteNow, quotaBaseUserRPD),
				},
			},
			want: QuotasUsage{
				PromptLengthMax: quotaPromptLengthMax(user),
				RateMinute: QuotaRequestsConsumption{
					Limit: quotaBaseUserRPM,
					Used:  quotaBaseUserRPM,
					Reset: quotasController.dayNext.Unix(),
				},
				RateDay: QuotaRequestsConsumption{
					Limit: quotaBaseUserRPD,
					Used:  quotaBaseUserRPD,
					Reset: quotasController.dayNext.Unix(),
				},
			},
			wantErr: false,
		},
		{
			name: "throttling quota exceeded",
			args: args{
				ctx: context.TODO(),
				clientRepository: MockRepositoryPrediction{
					Timestamps: repeatTimestamp(quotasController.minuteNow, quotaBaseUserRPM),
				},
			},
			want: QuotasUsage{
				PromptLengthMax: quotaPromptLengthMax(user),
				RateMinute: QuotaRequestsConsumption{
					Limit: quotaBaseUserRPM,
					Used:  quotaBaseUserRPM,
					Reset: quotasController.minuteNext.Unix(),
				},
				RateDay: QuotaRequestsConsumption{
					Limit: quotaBaseUserRPD,
					Used:  quotaBaseUserRPM,
					Reset: quotasController.dayNext.Unix(),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := GetQuotaUsage(tt.args.ctx, tt.args.clientRepository, user)
				if (err != nil) != tt.wantErr {
					t.Errorf("GetQuotaUsage() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetQuotaUsage() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestGetQuotaUsageRegisteredUser(t *testing.T) {
	type args struct {
		ctx              context.Context
		clientRepository RepositoryPrediction
	}
	user := &User{IsRegistered: true}
	tests := []struct {
		name    string
		args    args
		want    QuotasUsage
		wantErr bool
	}{
		{
			name: "no previous requests",
			args: args{
				ctx:              context.TODO(),
				clientRepository: MockRepositoryPrediction{},
			},
			want:    quotasController.quotaUsage(user),
			wantErr: false,
		},
		{
			name: "a single requests",
			args: args{
				ctx: context.TODO(),
				clientRepository: MockRepositoryPrediction{
					Timestamps: repeatTimestamp(quotasController.minuteNow, 1),
				},
			},
			want: QuotasUsage{
				PromptLengthMax: quotaPromptLengthMax(user),
				RateMinute: QuotaRequestsConsumption{
					Limit: quotaRegisteredUserRPM,
					Used:  1,
					Reset: quotasController.minuteNext.Unix(),
				},
				RateDay: QuotaRequestsConsumption{
					Limit: quotaRegisteredUserRPD,
					Used:  1,
					Reset: quotasController.dayNext.Unix(),
				},
			},
			wantErr: false,
		},
		{
			name: "daily quota exceeded",
			args: args{
				ctx: context.TODO(),
				clientRepository: MockRepositoryPrediction{
					Timestamps: repeatTimestamp(quotasController.minuteNow, quotaRegisteredUserRPD),
				},
			},
			want: QuotasUsage{
				PromptLengthMax: quotaPromptLengthMax(user),
				RateMinute: QuotaRequestsConsumption{
					Limit: quotaRegisteredUserRPM,
					Used:  quotaRegisteredUserRPM,
					Reset: quotasController.dayNext.Unix(),
				},
				RateDay: QuotaRequestsConsumption{
					Limit: quotaRegisteredUserRPD,
					Used:  quotaRegisteredUserRPD,
					Reset: quotasController.dayNext.Unix(),
				},
			},
			wantErr: false,
		},
		{
			name: "throttling quota exceeded",
			args: args{
				ctx: context.TODO(),
				clientRepository: MockRepositoryPrediction{
					Timestamps: repeatTimestamp(quotasController.minuteNow, quotaRegisteredUserRPM),
				},
			},
			want: QuotasUsage{
				PromptLengthMax: quotaPromptLengthMax(user),
				RateMinute: QuotaRequestsConsumption{
					Limit: quotaRegisteredUserRPM,
					Used:  quotaRegisteredUserRPM,
					Reset: quotasController.minuteNext.Unix(),
				},
				RateDay: QuotaRequestsConsumption{
					Limit: quotaRegisteredUserRPD,
					Used:  quotaRegisteredUserRPM,
					Reset: quotasController.dayNext.Unix(),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := GetQuotaUsage(tt.args.ctx, tt.args.clientRepository, user)
				if (err != nil) != tt.wantErr {
					t.Errorf("GetQuotaUsage() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetQuotaUsage() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}
