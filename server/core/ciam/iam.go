package ciam

import (
	"context"
	"time"
)

type User struct {
	ID       string
	APIToken string
	Role     Role
}

type Quotas struct {
	PromptLengthMax   uint16 `json:"prompt_length_max"`
	RequestsPerMinute uint16 `json:"rpm"`
	RequestsPerDay    uint16 `json:"rpd"`
}

type Role uint8

func (r Role) IsRegisteredUser() bool {
	return r == RoleRegisteredUser
}

func (r Role) IsValid() bool {
	switch r {
	case RoleAnonymUser, RoleRegisteredUser:
		return true
	default:
		return false
	}
}

func (r Role) Quotas() Quotas {
	switch r {
	case RoleAnonymUser:
		return Quotas{
			PromptLengthMax:   100,
			RequestsPerMinute: 1,
			RequestsPerDay:    5,
		}
	case RoleRegisteredUser:
		return Quotas{
			PromptLengthMax:   300,
			RequestsPerMinute: 3,
			RequestsPerDay:    20,
		}
	default:
		return Quotas{}
	}
}

const (
	RoleAnonymUser Role = iota
	RoleRegisteredUser
)

type QuotaRequestsConsumption struct {
	Limit uint16 `json:"limit"`
	Used  uint16 `json:"used"`
	Reset int64  `json:"reset"`
}

func (v quotaIssuer) quotaRPM(user *User) QuotaRequestsConsumption {
	return QuotaRequestsConsumption{
		Limit: user.Role.Quotas().RequestsPerMinute,
		Reset: v.minuteNext.Unix(),
	}
}

func (v quotaIssuer) quotaRPD(user *User) QuotaRequestsConsumption {
	return QuotaRequestsConsumption{
		Limit: user.Role.Quotas().RequestsPerDay,
		Reset: v.dayNext.Unix(),
	}
}

func (v quotaIssuer) quotaUsage(user *User) QuotasUsage {
	return QuotasUsage{
		PromptLengthMax: user.Role.Quotas().PromptLengthMax,
		RateMinute:      v.quotaRPM(user),
		RateDay:         v.quotaRPD(user),
	}
}

type quotaIssuer struct {
	dayNow     time.Time
	dayNext    time.Time
	minuteNow  time.Time
	minuteNext time.Time
}

func genNowDate() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}

func genNowMinute() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, time.UTC)
}

func newQuotaIssuer() quotaIssuer {
	const (
		day    = 24 * time.Hour
		minute = 1 * time.Minute
	)
	return quotaIssuer{
		dayNow:     genNowDate(),
		dayNext:    genNowDate().Add(day),
		minuteNow:  genNowMinute(),
		minuteNext: genNowMinute().Add(minute),
	}
}

type QuotasUsage struct {
	PromptLengthMax uint16                   `json:"prompt_length_max"`
	RateMinute      QuotaRequestsConsumption `json:"rate_minute"`
	RateDay         QuotaRequestsConsumption `json:"rate_day"`
}

func sliceWithinWindow(ts []time.Time, tsMin, tsMax time.Time) []time.Time {
	var o []time.Time
	for _, t := range ts {
		if t.After(tsMin) && t.Before(tsMax) || t == tsMin || t == tsMax {
			o = append(o, t)
		}
	}
	return o
}

// getQuotaUsage read current usage of the quota.
func getQuotaUsage(ctx context.Context, clientRepository RepositoryCIAM, user *User) (
	QuotasUsage, error,
) {
	requestsTimestamps, err := clientRepository.GetDailySuccessfulResultsTimestampsByUserID(ctx, user.ID)
	if err != nil {
		return QuotasUsage{}, err
	}

	quotasController := newQuotaIssuer()

	quotas := quotasController.quotaUsage(user)

	if len(requestsTimestamps) == 0 {
		return quotas, nil
	}

	requestsDaily := sliceWithinWindow(requestsTimestamps, quotasController.dayNow, quotasController.dayNext)
	quotas.RateDay.Used = uint16(len(requestsDaily))

	requestsMinute := sliceWithinWindow(requestsTimestamps, quotasController.minuteNow, quotasController.minuteNext)
	quotas.RateMinute.Used = uint16(len(requestsMinute))

	// by transitivity, the RPM/throttling quota is exceeded if the daily quota is exceeded
	if quotas.RateDay.Used >= quotas.RateDay.Limit {
		quotas.RateMinute.Used = quotas.RateMinute.Limit
		quotas.RateMinute.Reset = quotas.RateDay.Reset
	}

	return quotas, nil
}

var userKey = struct{}{}

func NewContext(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func FromContext(ctx context.Context) (*User, bool) {
	u, ok := ctx.Value(userKey).(*User)
	return u, ok
}
