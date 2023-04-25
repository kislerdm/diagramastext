package httphandler

import (
	"regexp"
)

type CIAMRequest interface {
	IsValid() bool
}

type ciamRequestAnonym struct {
	Fingerprint string `json:"fingerprint"`
}

func (c ciamRequestAnonym) IsValid() bool {
	f, _ := regexp.MatchString(`^[a-f0-9]{40}$`, c.Fingerprint)
	return f
}
