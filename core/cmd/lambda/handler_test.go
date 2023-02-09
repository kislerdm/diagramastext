package main

import (
	"math/rand"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

func Test_readPrompt(t *testing.T) {
	type args struct {
		req events.APIGatewayProxyRequest
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				req: events.APIGatewayProxyRequest{
					Body: `{"prompt":"foobar"}`,
				},
			},
			want:    "foobar",
			wantErr: false,
		},
		{
			name: "unhappy path: base64 encoded",
			args: args{
				req: events.APIGatewayProxyRequest{
					Body:            `eyJwcm9tcHQiOiJmb29iYXIifQ==`,
					IsBase64Encoded: true,
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "unhappy path: corrupt data",
			args: args{
				req: events.APIGatewayProxyRequest{
					Body: `{"prompt":"foobar"`,
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "unhappy path: corrupt data",
			args: args{
				req: events.APIGatewayProxyRequest{
					Body: `{"prompt":"foobar"`,
				},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := readPrompt(tt.args.req)
				if (err != nil) != tt.wantErr {
					t.Errorf("readPrompt() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("readPrompt() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_validatePrompt(t *testing.T) {
	type args struct {
		prompt string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid",
			args: args{
				prompt: "c4 diagram with Go backend reading postgres",
			},
			wantErr: false,
		},
		{
			name: "invalid: short",
			args: args{
				prompt: randomString(promptLengthMin - 1),
			},
			wantErr: true,
		},
		{
			name: "invalid: long",
			args: args{
				prompt: randomString(promptLengthMax + 1),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if err := validatePrompt(tt.args.prompt); (err != nil) != tt.wantErr {
					t.Errorf("validatePrompt() error = %v, wantErr %v", err, tt.wantErr)
				}
			},
		)
	}
}

func randomString(length int) string {
	const charset = "abcdef"
	var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	var b = make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
