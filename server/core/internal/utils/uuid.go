package utils

import (
	"github.com/google/uuid"
)

// NewUUID generated UUID as a string.
func NewUUID() string {
	return uuid.New().String()
}

// ValidateUUID validates UUID represented as the string.
func ValidateUUID(s string) error {
	_, err := uuid.Parse(s)
	return err
}
