package ciam

import (
	"crypto/ed25519"
	"math/rand"
	"reflect"
	"testing"

	"github.com/kislerdm/diagramastext/server/core/diagram"
	"github.com/kislerdm/diagramastext/server/core/internal/utils"
)

func TestCircular(t *testing.T) {
	_, priv, err := ed25519.GenerateKey(rand.New(rand.NewSource(0)))
	if err != nil {
		t.Fatal(err)
	}

	issuer, err := NewIssuer(priv)
	if err != nil {
		t.Fatal(err)
	}

	userWant := diagram.User{
		ID:   utils.NewUUID(),
		Role: diagram.RoleRegisteredUser,
	}
	const (
		email       = "foo@bar.baz"
		fingerprint = "qux"
	)

	t.Parallel()
	t.Run(
		"shall parse generated id token", func(t *testing.T) {
			tknStr, err := issuer.NewIDToken(userWant.ID, email, fingerprint)
			if err != nil {
				t.Fatalf("failed to generate token: %v", err)
			}

			userIDGot, emailGot, fingerprintGot, err := issuer.ParseIDToken(tknStr)
			if err != nil {
				t.Fatalf("failed to parse generated token: %v", err)
			}

			if userWant.ID != userIDGot {
				t.Errorf("wront userID extracted from the token. want: %s, got: %s", userWant.ID, userIDGot)
			}

			if emailGot != email {
				t.Errorf("wront email extracted from the token. want: %s, got: %s", email, emailGot)
			}

			if fingerprintGot != fingerprint {
				t.Errorf(
					"wront fingerprint extracted from the token. want: %s, got: %s",
					fingerprint, fingerprintGot,
				)
			}
		},
	)

	t.Run(
		"shall parse generated access token", func(t *testing.T) {
			tknStr, err := issuer.NewAccessToken(userWant)
			if err != nil {
				t.Fatalf("failed to generate token: %v", err)
			}

			got, err := issuer.ParseAccessToken(tknStr)
			if err != nil {
				t.Fatalf("failed to parse generated token: %v", err)
			}

			if !reflect.DeepEqual(userWant, got) {
				t.Errorf("wront user data extracted from the token. want: %v, got: %v", userWant, got)
			}
		},
	)

	t.Run(
		"shall parse generated refresh token", func(t *testing.T) {
			tknStr, err := issuer.NewRefreshToken(userWant.ID)
			if err != nil {
				t.Fatalf("failed to generate token: %v", err)
			}

			userIDGot, err := issuer.ParseRefreshToken(tknStr)
			if err != nil {
				t.Fatalf("failed to parse generated token: %v", err)
			}

			if userWant.ID != userIDGot {
				t.Errorf("wront userIDWant extracted from the token. want: %s, got: %s", userWant.ID, userIDGot)
			}
		},
	)
}
