package ciam

import (
	"context"
	"crypto/ed25519"
	"errors"
	"reflect"
	"testing"
)

func generateKeyPairEd25519() ([]byte, []byte) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic(err)
	}
	return priv, pub
}

func TestNewTokenSigningClientEd25519(t *testing.T) {
	type args func() ([]byte, []byte)

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "happy path",
			args:    generateKeyPairEd25519,
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: func() ([]byte, []byte) {
				k0, _ := generateKeyPairEd25519()
				_, k1 := generateKeyPairEd25519()
				return k0, k1
			},
			wantErr: true,
		},
	}
	t.Parallel()
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				_, err := NewTokenSigningClientEd25519(tt.args())
				if (err != nil) != tt.wantErr {
					t.Errorf("NewTokenSigningClientEd25519() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			},
		)
	}
}

func TestTracingSignAndVerify(t *testing.T) {
	t.Parallel()
	t.Run(
		"shall validate the string which it signed", func(t *testing.T) {
			// GIVEN
			client, err := NewTokenSigningClientEd25519(generateKeyPairEd25519())
			if err != nil {
				t.Fatal(err)
			}
			const signingString = "foo"

			// WHEN
			signature, alg, err := client.Sign(context.TODO(), signingString)

			// THEN
			if err != nil {
				t.Errorf("unexpected signig error: %+v", err)
			}
			if alg != "EdDSA" {
				t.Errorf("unexpected signing alg: %s", alg)
			}

			if err := client.Verify(context.TODO(), signingString, signature); err != nil {
				t.Errorf("unexpected verification error: %+v", err)
			}
		},
	)
	t.Run(
		"shall fail validation of the string signed with other key", func(t *testing.T) {
			// GIVEN
			clientSign, err := NewTokenSigningClientEd25519(generateKeyPairEd25519())
			if err != nil {
				t.Fatal(err)
			}

			clientVerify, err := NewTokenSigningClientEd25519(generateKeyPairEd25519())
			if err != nil {
				t.Fatal(err)
			}

			const signingString = "foo"

			// WHEN
			signature, alg, err := clientSign.Sign(context.TODO(), signingString)

			// THEN
			if err != nil {
				t.Errorf("unexpected signig error: %+v", err)
			}
			if alg != "EdDSA" {
				t.Errorf("unexpected signing alg: %s", alg)
			}

			if err := clientVerify.Verify(context.TODO(), signingString, signature); !reflect.DeepEqual(
				err, errors.New("failed signature verification"),
			) {
				t.Errorf("verification error expected: got=%+v", err)
			}
		},
	)
}
