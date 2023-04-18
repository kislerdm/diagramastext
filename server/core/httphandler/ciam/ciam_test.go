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
	//GIVEN
	const fingerprint = "foo"
	tokenSignClient := MockTokenSigningClient{
		Alg:       "EdDSA",
		Signature: "qux",
	}

	repoClient := &MockRepositoryCIAM{}

	smtpClient := SMTPClient(nil)

	ciamClient := NewClient(repoClient, tokenSignClient, smtpClient)

	//WHEN
	tokens, err := ciamClient.SigninAnonym(context.TODO(), fingerprint)

	//THEN
	if err != nil {
		t.Errorf("unexpected error")
	}

	validateToken(t, tokens.ID, "id", tokenSignClient, defaultExpirationDurationIdentitySec)
	validateToken(t, tokens.Access, "access", tokenSignClient, defaultExpirationDurationAccessSec)
	validateToken(t, tokens.Refresh, "refresh", tokenSignClient, defaultExpirationDurationRefreshSec)

	if tokens.ID.Sub() != tokens.Refresh.Sub() || tokens.ID.Sub() != tokens.Access.Sub() {
		t.Errorf("sub does not match")
	}
	if err := utils.ValidateUUID(tokens.ID.Sub()); err != nil {
		t.Errorf("wrong sub format: %+v", err)
	}

	found, isActive, emailVerified, email, gotFingerprint, err := repoClient.ReadUser(context.TODO(), tokens.ID.Sub())
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
	if fingerprint != gotFingerprint || fingerprint != tokens.ID.Fingerprint() {
		t.Errorf("user's fingerprint was set incorrectly")
	}
	if tokens.Access.Role() != roleAnonymUser {
		t.Errorf("user's role was set incorrectly")
	}
	if tokens.ID.Email() != "" {
		t.Errorf("user's email was set incorrectly")
	}
}

func validateToken(t *testing.T, tkn JWT, tokenTyp string, clientSign MockTokenSigningClient, wantTTL int64) {
	if err := tkn.Validate(
		func(signingString, signature string) error {
			return clientSign.Verify(nil, signature, signature)
		},
	); err != nil {
		t.Errorf("%s wrong signrature: %+v", tokenTyp, err)
	}

	tk, ok := tkn.(*token)
	if !ok {
		t.Errorf("%s wrong token format", tokenTyp)
	}
	if tk.Header.Alg != clientSign.Alg {
		t.Errorf("%s wrong header: alg", tokenTyp)
	}
	if tk.Header.Typ != typ {
		t.Errorf("%s wrong header: typ", tokenTyp)
	}
	if tk.Payload.Aud != aud {
		t.Errorf("%s wrong payload: aud", tokenTyp)
	}
	if tk.Payload.Iss != iss {
		t.Errorf("%s wrong payload: iss", tokenTyp)
	}
	if tk.Payload.Exp-tk.Payload.Iat != wantTTL {
		t.Errorf("%s wrong payload: iat and exp", tokenTyp)
	}
}
