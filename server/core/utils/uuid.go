package utils

import "github.com/aws/smithy-go/rand"

// NewUUID generated UUID as a string.
func NewUUID() string {
	o, _ := rand.NewUUID(rand.Reader).GetUUID()
	return o
}
