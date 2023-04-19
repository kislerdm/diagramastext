package ciam

import (
	"context"
	"testing"

	"github.com/kislerdm/diagramastext/server/core/internal/utils"
)

func TestSigninAnonymFlow(t *testing.T) {
	t.Parallel()
	t.Run("shall create a user and issue tokens", testSigninAnonymFlowUserDidNotExist)
}

func testSigninAnonymFlowUserDidNotExist(t *testing.T) {
	// GIVEN
	const fingerprint = "foo"
	tokenSignClient := MockTokenSigningClient{
		Alg:       "EdDSA",
		Signature: "qux",
	}

	repoClient := &MockRepositoryCIAM{}

	smtpClient := SMTPClient(nil)

	ciamClient := NewClient(repoClient, tokenSignClient, smtpClient)

	// WHEN
	tokens, err := ciamClient.SigninAnonym(context.TODO(), fingerprint)

	// THEN
	if err != nil {
		t.Errorf("unexpected error")
	}

	validateToken(t, tokens.id, "id", tokenSignClient, defaultExpirationDurationIdentitySec)
	validateToken(t, tokens.access, "access", tokenSignClient, defaultExpirationDurationAccessSec)
	validateToken(t, tokens.refresh, "refresh", tokenSignClient, defaultExpirationDurationRefreshSec)

	if tokens.id.UserID() != tokens.refresh.UserID() || tokens.id.UserID() != tokens.access.UserID() {
		t.Errorf("sub does not match")
	}
	if err := utils.ValidateUUID(tokens.id.UserID()); err != nil {
		t.Errorf("wrong sub format: %+v", err)
	}

	found, isActive, emailVerified, email, gotFingerprint, err := repoClient.ReadUser(
		context.TODO(), tokens.id.UserID(),
	)
	if err != nil {
		t.Errorf("unexpected error: CIAM repository")
	}
	if !found {
		t.Errorf("user was not recorded to the CIAM repository")
	}
	if !isActive {
		t.Errorf("user's activity was not set corretly, true expected")
	}
	if emailVerified {
		t.Errorf("user email's verification was not set corretly, false expected")
	}
	if email != "" {
		t.Errorf("user's email was set incorrectly")
	}
	if fingerprint != gotFingerprint || fingerprint != tokens.id.UserDeviceFingerprint() {
		t.Errorf("user's fingerprint was set incorrectly")
	}
	if tokens.access.UserRole() != roleAnonymUser {
		t.Errorf("user's role was set incorrectly")
	}
	if tokens.id.UserEmail() != "" {
		t.Errorf("user's email was set incorrectly")
	}
}

func validateToken(t *testing.T, tkn JWT, tokenTyp string, clientSign MockTokenSigningClient, wantTTL int64) {
	if err := tkn.Validate(
		func(signingString, signature string) error {
			return clientSign.Verify(context.TODO(), signingString, signature)
		},
	); err != nil {
		t.Errorf("%s wrong signrature: %+v", tokenTyp, err)
	}

	tk, ok := tkn.(*token)
	if !ok {
		t.Errorf("%s wrong token format", tokenTyp)
	}
	if tk.header.Alg != clientSign.Alg {
		t.Errorf("%s wrong header: alg", tokenTyp)
	}
	if tk.header.Typ != typ {
		t.Errorf("%s wrong header: typ", tokenTyp)
	}
	if tk.payload.Aud != aud {
		t.Errorf("%s wrong payload: aud", tokenTyp)
	}
	if tk.payload.Iss != iss {
		t.Errorf("%s wrong payload: iss", tokenTyp)
	}
	if tk.payload.Exp-tk.payload.Iat != wantTTL {
		t.Errorf("%s wrong payload: iat and exp", tokenTyp)
	}
}
