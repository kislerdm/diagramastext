package ciam

import (
	"crypto/ed25519"
	"math/rand"
	"reflect"
	"testing"

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

	userIDWant := utils.NewUUID()
	const (
		email = "foo@bar.baz"
	)

	t.Parallel()
	t.Run(
		"shall parse generated id token", func(t *testing.T) {
			tknStr, err := issuer.NewIDToken(userIDWant, email, "")
			if err != nil {
				t.Fatalf("failed to generate token: %v", err)
			}

			userIDGot, err := issuer.ParseIDToken(tknStr)
			if err != nil {
				t.Fatalf("failed to parse generated token: %v", err)
			}

			if userIDWant != userIDGot {
				t.Errorf("wront userID extracted from the token. want: %s, got: %s", userIDWant, userIDGot)
			}
		},
	)

	t.Run(
		"shall parse generated access token", func(t *testing.T) {
			const roleWant = roleRegisteredUser
			quotasWant := QuotasRegisteredUser

			tknStr, err := issuer.NewAccessToken(userIDWant, roleWant)
			if err != nil {
				t.Fatalf("failed to generate token: %v", err)
			}

			userIDGot, roleGot, quotasGot, err := issuer.ParseAccessToken(tknStr)
			if err != nil {
				t.Fatalf("failed to parse generated token: %v", err)
			}

			if userIDWant != userIDGot {
				t.Errorf("wront userIDWant extracted from the token. want: %s, got: %s", userIDWant, userIDGot)
			}

			if roleGot != roleWant {
				t.Errorf("wront role extracted from the token. want: %v, got: %v", roleWant, roleGot)
			}

			if !reflect.DeepEqual(quotasGot, quotasWant) {
				t.Errorf("wront quotas extracted from the token. want: %v, got: %v", quotasWant, quotasGot)
			}
		},
	)

	t.Run(
		"shall parse generated refresh token", func(t *testing.T) {
			tknStr, err := issuer.NewRefreshToken(userIDWant)
			if err != nil {
				t.Fatalf("failed to generate token: %v", err)
			}

			userIDGot, err := issuer.ParseRefreshToken(tknStr)
			if err != nil {
				t.Fatalf("failed to parse generated token: %v", err)
			}

			if userIDWant != userIDGot {
				t.Errorf("wront userIDWant extracted from the token. want: %s, got: %s", userIDWant, userIDGot)
			}
		},
	)
}
