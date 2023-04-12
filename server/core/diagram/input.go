package diagram

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/kislerdm/diagramastext/server/core/internal/utils"
)

type User struct {
	ID           string
	APIToken     string
	IsRegistered bool
}

// Input defines the entrypoint interface.
type Input interface {
	Validate() error
	GetUser() *User
	GetPrompt() string
	GetRequestID() string
}

type MockInput struct {
	Err       error
	Prompt    string
	RequestID string
	User      *User
}

func (v MockInput) Validate() error {
	return v.Err
}

func (v MockInput) GetUser() *User {
	return v.User
}

func (v MockInput) GetPrompt() string {
	return strings.ReplaceAll(v.Prompt, "\n", "")
}

func (v MockInput) GetRequestID() string {
	return v.RequestID
}

type inquiry struct {
	Prompt    string
	RequestID string
	User      *User
}

const (
	promptLengthMin = 3

	// QUOTAS
	// base user
	quotaBaseUserPromptLengthMax = 100
	quotaBaseUserRPM             = 2
	quotaBaseUserRPD             = 10

	// registered user
	quotaRegisteredUserPromptLengthMax = 300
	quotaRegisteredUserRPM             = 3
	quotaRegisteredUserRPD             = 20
)

func (v inquiry) GetPrompt() string {
	return v.Prompt
}

func (v inquiry) GetRequestID() string {
	return v.RequestID
}

func (v inquiry) GetUser() *User {
	return v.User
}

func (v inquiry) Validate() error {
	prompt := strings.ReplaceAll(v.Prompt, "\n", "")
	return validatePromptLength(prompt, quotaPromptLengthMax(v.GetUser()))
}

func quotaPromptLengthMax(user *User) int {
	if user.IsRegistered {
		return quotaRegisteredUserPromptLengthMax
	}
	return quotaBaseUserPromptLengthMax
}

func validatePromptLength(prompt string, max int) error {
	if len(prompt) < promptLengthMin || len(prompt) > max {
		return errors.New(
			"prompt length must be between " + strconv.Itoa(promptLengthMin) + " and " +
				strconv.Itoa(max) + " characters",
		)
	}
	return nil
}

// NewInput initialises the `Input` object.
func NewInput(prompt string, user *User) (Input, error) {
	o := &inquiry{
		Prompt:    prompt,
		User:      user,
		RequestID: utils.NewUUID(),
	}

	if err := o.Validate(); err != nil {
		return nil, err
	}

	return o, nil
}

type QuotaRequestsConsumption struct {
	Limit int   `json:"limit"`
	Used  int   `json:"used"`
	Reset int64 `json:"reset"`
}

func quotaRPM(user *User) QuotaRequestsConsumption {
	if user.IsRegistered {
		return QuotaRequestsConsumption{
			Limit: quotaRegisteredUserRPM,
		}
	}
	return QuotaRequestsConsumption{
		Limit: quotaBaseUserRPM,
	}
}

func quotaRPD(user *User) QuotaRequestsConsumption {
	if user.IsRegistered {
		return QuotaRequestsConsumption{
			Limit: quotaRegisteredUserRPD,
		}
	}
	return QuotaRequestsConsumption{
		Limit: quotaBaseUserRPD,
	}
}

// ValidateRequestsQuotaUsage checks if the requests' quota was exceeded.
func ValidateRequestsQuotaUsage(ctx context.Context, clientRepository RepositoryPrediction, user *User) (
	throttling bool, quotaExceeded bool, err error,
) {
	quotasUsage, err := GetQuotaUsage(ctx, clientRepository, user)
	if err != nil {
		return
	}

	if quotasUsage.RateMinute.Used >= quotasUsage.RateMinute.Limit {
		throttling = true
	}

	if quotasUsage.RateDay.Used >= quotasUsage.RateDay.Limit {
		quotaExceeded = true
	}

	return
}

type QuotasUsage struct {
	PromptLengthMax int                      `json:"prompt_length_max"`
	RateMinute      QuotaRequestsConsumption `json:"rate_minute"`
	RateDay         QuotaRequestsConsumption `json:"rate_day"`
}

func sliceWithinWindow(ts []time.Time, tsMin, tsMax time.Time) []time.Time {
	var o []time.Time
	for _, t := range ts {
		if t.After(tsMin) && t.Before(tsMax) {
			o = append(o, t)
		}
	}
	return o
}

const (
	day    = 24 * time.Hour
	minute = 1 * time.Minute
)

var (
	today     = time.Now().UTC().Round(day)
	nowMinute = time.Now().UTC().Round(minute)
)

// GetQuotaUsage read current usage of the quota.
func GetQuotaUsage(ctx context.Context, clientRepository RepositoryPrediction, user *User) (QuotasUsage, error) {
	requestsTimestamps, err := clientRepository.GetDailySuccessfulResultsTimestampsByUserID(ctx, user.ID)
	if err != nil {
		return QuotasUsage{}, err
	}

	quotas := QuotasUsage{
		PromptLengthMax: quotaPromptLengthMax(user),
		RateMinute:      quotaRPM(user),
		RateDay:         quotaRPD(user),
	}

	if len(requestsTimestamps) == 0 {
		return quotas, nil
	}

	tomorrow := today.Add(day)
	requestsDaily := sliceWithinWindow(requestsTimestamps, today, tomorrow)
	quotas.RateDay.Used = len(requestsDaily)
	quotas.RateDay.Reset = tomorrow.Unix()

	nextMinute := nowMinute.Add(minute)
	requestsMinute := sliceWithinWindow(requestsTimestamps, nowMinute, nextMinute)
	quotas.RateMinute.Used = len(requestsMinute)
	quotas.RateMinute.Reset = nextMinute.Unix()

	return quotas, nil
}
