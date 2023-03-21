package utils

import (
	"github.com/google/uuid"
)

// NewUUID generated UUID as a string.
func NewUUID() string {
	return uuid.New().String()
}
