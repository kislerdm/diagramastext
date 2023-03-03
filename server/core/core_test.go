package core

import (
	"errors"
	"math/rand"
	"strconv"
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

func Test_validatePrompt(t *testing.T) {
	type args struct {
		prompt string
		max    int
	}
	tests := []struct {
		name string
		args args
		want error
	}{
		{
			name: "valid",
			args: args{
				prompt: "c4 diagram with Go backend reading postgres",
				max:    promptLengthMaxRegistered,
			},
			want: nil,
		},
		{
			name: "invalid: short",
			args: args{
				prompt: randomString(promptLengthMin - 1),
				max:    promptLengthMaxRegistered,
			},
			want: errors.New(
				"prompt length must be between " + strconv.Itoa(promptLengthMin) + " and " +
					strconv.Itoa(promptLengthMaxRegistered) + " characters",
			),
		},
		{
			name: "invalid: long",
			args: args{
				prompt: randomString(promptLengthMaxRegistered + 1),
				max:    promptLengthMaxRegistered,
			},
			want: errors.New(
				"prompt length must be between " + strconv.Itoa(promptLengthMin) + " and " +
					strconv.Itoa(promptLengthMaxRegistered) + " characters",
			),
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if err := validatePromptLength(tt.args.prompt, tt.args.max); (err != nil) && errors.Is(err, tt.want) {
					t.Errorf("validatePromptLength() error = %v, wantErr %v", err, tt.want)
				}
			},
		)
	}
}
