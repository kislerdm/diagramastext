package ciam

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/kislerdm/diagramastext/server/core/diagram"
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

func TestValidateRequestsQuotaUsage(t *testing.T) {
	type args struct {
		ctx              context.Context
		clientRepository diagram.RepositoryPrediction
		user             *User
	}

	const (
		quotaRPMMax = 2
		quotaRPDMax = 10
	)

	tests := []struct {
		name              string
		args              args
		wantThrottling    bool
		wantQuotaExceeded bool
		wantErr           bool
	}{
		{
			name: "no request made so far",
			args: args{
				ctx:              context.TODO(),
				clientRepository: diagram.MockRepositoryPrediction{},
				user:             &User{},
			},
			wantThrottling:    false,
			wantQuotaExceeded: false,
			wantErr:           false,
		},
		{
			name: "throttling quota exceeded",
			args: args{
				ctx: context.TODO(),
				clientRepository: diagram.MockRepositoryPrediction{
					Timestamps: repeatTimestamp(genNowMinute(), RoleRegisteredUser.Quotas().RequestsPerMinute+1),
				},
				user: &User{},
			},
			wantThrottling:    true,
			wantQuotaExceeded: false,
			wantErr:           false,
		},
		{
			name: "daily quota exceeded",
			args: args{
				ctx: context.TODO(),
				clientRepository: diagram.MockRepositoryPrediction{
					Timestamps: repeatTimestamp(genNowDate(), RoleRegisteredUser.Quotas().RequestsPerDay+1),
				},
				user: &User{},
			},
			wantThrottling:    true,
			wantQuotaExceeded: true,
			wantErr:           false,
		},
		{
			name: "unhappy path",
			args: args{
				ctx: context.TODO(),
				clientRepository: diagram.MockRepositoryPrediction{
					Err: errors.New("foo"),
				},
				user: &User{},
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

func TestGetQuotaUsage(t *testing.T) {
	type args struct {
		ctx              context.Context
		clientRepository diagram.RepositoryPrediction
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
				clientRepository: diagram.MockRepositoryPrediction{},
			},
			want:    quotasController.quotaUsage(user),
			wantErr: false,
		},
		{
			name: "a single requests",
			args: args{
				ctx: context.TODO(),
				clientRepository: diagram.MockRepositoryPrediction{
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
				clientRepository: diagram.MockRepositoryPrediction{
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
				clientRepository: diagram.MockRepositoryPrediction{
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
