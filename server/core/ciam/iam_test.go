package ciam

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/kislerdm/diagramastext/server/core/internal/utils"
)

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

func repeatTimestamp(ts time.Time, nElements uint16) []time.Time {
	o := make([]time.Time, nElements)
	var i uint16
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

func Test_getQuotaUsage(t *testing.T) {
	type args struct {
		ctx              context.Context
		clientRepository RepositoryCIAM
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
				clientRepository: &MockRepositoryCIAM{},
			},
			want:    quotasController.quotaUsage(user),
			wantErr: false,
		},
		{
			name: "a single requests",
			args: args{
				ctx: context.TODO(),
				clientRepository: &MockRepositoryCIAM{
					Timestamps: repeatTimestamp(quotasController.minuteNow, 1),
				},
			},
			want: QuotasUsage{
				PromptLengthMax: user.Role.Quotas().PromptLengthMax,
				RateMinute: QuotaRequestsConsumption{
					Limit: user.Role.Quotas().RequestsPerMinute,
					Used:  1,
					Reset: quotasController.minuteNext.Unix(),
				},
				RateDay: QuotaRequestsConsumption{
					Limit: user.Role.Quotas().RequestsPerDay,
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
				clientRepository: &MockRepositoryCIAM{
					Timestamps: repeatTimestamp(quotasController.minuteNow, user.Role.Quotas().RequestsPerDay),
				},
			},
			want: QuotasUsage{
				PromptLengthMax: user.Role.Quotas().PromptLengthMax,
				RateMinute: QuotaRequestsConsumption{
					Limit: user.Role.Quotas().RequestsPerMinute,
					Used:  user.Role.Quotas().RequestsPerMinute,
					Reset: quotasController.dayNext.Unix(),
				},
				RateDay: QuotaRequestsConsumption{
					Limit: user.Role.Quotas().RequestsPerDay,
					Used:  user.Role.Quotas().RequestsPerDay,
					Reset: quotasController.dayNext.Unix(),
				},
			},
			wantErr: false,
		},
		{
			name: "throttling quota exceeded",
			args: args{
				ctx: context.TODO(),
				clientRepository: &MockRepositoryCIAM{
					Timestamps: repeatTimestamp(quotasController.minuteNow, user.Role.Quotas().RequestsPerMinute),
				},
			},
			want: QuotasUsage{
				PromptLengthMax: user.Role.Quotas().PromptLengthMax,
				RateMinute: QuotaRequestsConsumption{
					Limit: user.Role.Quotas().RequestsPerMinute,
					Used:  user.Role.Quotas().RequestsPerMinute,
					Reset: quotasController.minuteNext.Unix(),
				},
				RateDay: QuotaRequestsConsumption{
					Limit: user.Role.Quotas().RequestsPerDay,
					Used:  user.Role.Quotas().RequestsPerMinute,
					Reset: quotasController.dayNext.Unix(),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := getQuotaUsage(tt.args.ctx, tt.args.clientRepository, user)
				if (err != nil) != tt.wantErr {
					t.Errorf("getQuotaUsage() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("getQuotaUsage() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestRole_IsRegisteredUser(t *testing.T) {
	tests := []struct {
		name string
		r    Role
		want bool
	}{
		{
			name: "registered",
			r:    RoleRegisteredUser,
			want: true,
		},
		{
			name: "not registered",
			r:    RoleAnonymUser,
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := tt.r.IsRegisteredUser(); got != tt.want {
					t.Errorf("IsRegisteredUser() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_client_validateRequestsQuotaUsage(t *testing.T) {
	type args struct {
		clientRepository RepositoryCIAM
		user             *User
		writer           http.ResponseWriter
	}

	certificate := GenerateCertificate()

	tests := []struct {
		name          string
		args          args
		wantFn        func(w http.ResponseWriter) error
		want          bool
		wantBody      []byte
		wantStatuCode int
	}{
		{
			name: "no request made so far",
			args: args{
				clientRepository: &MockRepositoryCIAM{},
				user:             &User{},
				writer:           &utils.MockWriter{},
			},
			wantStatuCode: 0,
			wantBody:      nil,
			want:          true,
		},
		{
			name: "throttling quota exceeded",
			args: args{
				clientRepository: &MockRepositoryCIAM{
					Timestamps: repeatTimestamp(genNowMinute(), RoleRegisteredUser.Quotas().RequestsPerMinute+1),
				},
				user:   &User{},
				writer: &utils.MockWriter{},
			},
			wantStatuCode: http.StatusTooManyRequests,
			wantBody:      []byte(`{"error":"throttling quota exceeded"}`),
			want:          false,
		},
		{
			name: "daily quota exceeded",
			args: args{
				clientRepository: &MockRepositoryCIAM{
					Timestamps: repeatTimestamp(genNowDate(), RoleRegisteredUser.Quotas().RequestsPerDay+1),
				},
				user:   &User{},
				writer: &utils.MockWriter{},
			},
			wantStatuCode: http.StatusTooManyRequests,
			wantBody:      []byte(`{"error":"daily quota exceeded"}`),
			want:          false,
		},
		{
			name: "unhappy path",
			args: args{
				clientRepository: &MockRepositoryCIAM{
					Err: errors.New("foo"),
				},
				user:   &User{},
				writer: &utils.MockWriter{},
			},
			wantStatuCode: http.StatusInternalServerError,
			wantBody:      []byte(`{"error":"internal error"}`),
			want:          false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c, err := HTTPHandler(tt.args.clientRepository, &MockSMTPClient{}, certificate)
				if err != nil {
					t.Fatal(err)
				}

				got := c(nil).(client).validateRequestsQuotaUsage(tt.args.writer, &http.Request{}, tt.args.user)

				if got != tt.want {
					t.Errorf("unexpected return value. want: %v, got: %v", tt.want, got)
				}

				gotStatusCode := tt.args.writer.(*utils.MockWriter).StatusCode
				if tt.wantStatuCode != gotStatusCode {
					t.Errorf("wrong status code. want: %d, got: %d", tt.wantStatuCode, gotStatusCode)
				}

				gotBody := tt.args.writer.(*utils.MockWriter).V
				if !reflect.DeepEqual(gotBody, tt.wantBody) {
					t.Errorf("wrong body. want: %v, got: %v", tt.wantBody, gotBody)
				}
			},
		)
	}
}
