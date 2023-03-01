package server

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

func TestValidatePrompt(t *testing.T) {
	type args struct {
		prompt string
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
			},
			want: nil,
		},
		{
			name: "invalid: short",
			args: args{
				prompt: randomString(PromptLengthMin - 1),
			},
			want: errors.New(
				"prompt length must be between " + strconv.Itoa(PromptLengthMin) + " and " +
					strconv.Itoa(PromptLengthMax) + " characters",
			),
		},
		{
			name: "invalid: long",
			args: args{
				prompt: randomString(PromptLengthMax + 1),
			},
			want: errors.New(
				"prompt length must be between " + strconv.Itoa(PromptLengthMin) + " and " +
					strconv.Itoa(PromptLengthMax) + " characters",
			),
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if err := validatePrompt(tt.args.prompt); (err != nil) && errors.Is(err, tt.want) {
					t.Errorf("validatePrompt() error = %v, wantErr %v", err, tt.want)
				}
			},
		)
	}
}
