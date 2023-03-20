package gcpsecretsmanager

import (
	"context"
	_ "embed"
	"errors"
	"hash/crc32"
	"os"
	"testing"

	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

//go:embed dummy-key.json
var key []byte

func TestNewSecretmanager(t *testing.T) {
	t.Parallel()

	t.Run(
		"happy path", func(t *testing.T) {
			// GIVEN
			// The credentials json key
			dir := t.TempDir()
			p := dir + "/google-key.json"
			if err := os.WriteFile(p, key, 0777); err != nil {
				t.Fatal(err)
			}
			t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", p)
			// WHEN
			_, err := NewSecretmanager(context.TODO())

			// THEN
			if err != nil {
				t.Errorf("NewSecretmanager(): no error expected")
				return
			}
		},
	)

	t.Run(
		"unhappy path", func(t *testing.T) {
			// GIVEN
			// No credentials
			// WHEN
			_, err := NewSecretmanager(context.TODO())

			// THEN
			if err == nil {
				t.Errorf("NewSecretmanager(): error expected")
				return
			}
		},
	)
}

type secret struct {
	Foo string `json:"foo"`
}

func probeChecksum(v []byte) *int64 {
	crc32c := crc32.MakeTable(crc32.Castagnoli)
	o := int64(crc32.Checksum(v, crc32c))
	return &o
}

func TestClient_ReadLastVersionHappyPath(t *testing.T) {
	t.Run(
		"happy path", func(t *testing.T) {
			// GIVEN
			client := Client{
				c: mockGCPSecretsmanagerClient{
					v: &secretmanagerpb.AccessSecretVersionResponse{
						Name: "foo",
						Payload: &secretmanagerpb.SecretPayload{
							Data:       []byte(`{"foo":"bar"}`),
							DataCrc32C: probeChecksum([]byte(`{"foo":"bar"}`)),
						},
					},
				},
			}

			// WHEN
			var s secret
			err := client.ReadLastVersion(context.TODO(), "foo", &s)

			// THEN
			if err != nil {
				t.Fatalf("unexpected error: %+v", err)
			}

			if s.Foo != "bar" {
				t.Fatalf("unexpected secret value")
			}
		},
	)
}

func TestClient_ReadLastVersionUnhappyPaths(t *testing.T) {
	type fields struct {
		c gcpSecretsmanagerClient
	}
	type args struct {
		ctx    context.Context
		uri    string
		output interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "missing secret",
			fields: fields{
				c: mockGCPSecretsmanagerClient{
					v:   nil,
					err: errors.New("secret 'foo' is missing"),
				},
			},
			args: args{
				ctx:    context.TODO(),
				uri:    "foo",
				output: nil,
			},
			wantErr: true,
		},
		{
			name: "wrong payload checksum",
			fields: fields{
				c: mockGCPSecretsmanagerClient{
					v: &secretmanagerpb.AccessSecretVersionResponse{
						Name: "foo",
						Payload: &secretmanagerpb.SecretPayload{
							Data: []byte(`{"foo":"bar"}`),
						},
					},
				},
			},
			args: args{
				ctx:    context.TODO(),
				uri:    "foo",
				output: nil,
			},
			wantErr: true,
		},
		{
			name: "wrong payload encoding format",
			fields: fields{
				c: mockGCPSecretsmanagerClient{
					v: &secretmanagerpb.AccessSecretVersionResponse{
						Name: "foo",
						Payload: &secretmanagerpb.SecretPayload{
							Data:       []byte(`"foo":"bar"`),
							DataCrc32C: probeChecksum([]byte(`"foo":"bar"`)),
						},
					},
				},
			},
			args: args{
				ctx:    context.TODO(),
				uri:    "foo",
				output: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c := Client{
					c: tt.fields.c,
				}
				if err := c.ReadLastVersion(tt.args.ctx, tt.args.uri, tt.args.output); (err != nil) != tt.wantErr {
					t.Errorf("ReadLastVersion() error = %v, wantErr %v", err, tt.wantErr)
				}
			},
		)
	}
}

func Test_latestVersionURI(t *testing.T) {
	type args struct {
		uri string
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
				uri: "projects/619195795350/secrets/core",
			},
			want:    "projects/619195795350/secrets/core/versions/latest",
			wantErr: false,
		},
		{
			name: "happy path: fixed version",
			args: args{
				uri: "projects/619195795350/secrets/core/versions/1",
			},
			want:    "projects/619195795350/secrets/core/versions/latest",
			wantErr: false,
		},
		{
			name: "happy path: corrupt version",
			args: args{
				uri: "projects/619195795350/secrets/core/version/1",
			},
			want:    "projects/619195795350/secrets/core/versions/latest",
			wantErr: false,
		},
		{
			name: "unhappy path: too few URI elements",
			args: args{
				uri: "projects/619195795350/secrets",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "unhappy path: corrupt secrets tag",
			args: args{
				uri: "projects/619195795350/secret/foo",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "unhappy path: corrupt project tag",
			args: args{
				uri: "project/619195795350/secrets/foo",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := latestVersionURI(tt.args.uri)
				if (err != nil) != tt.wantErr {
					t.Errorf("latestVersionURI() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("latestVersionURI() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}
