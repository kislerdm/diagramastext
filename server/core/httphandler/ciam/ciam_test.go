package ciam

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/kislerdm/diagramastext/server/core/internal/utils"
)

func TestSigninAnonymFlow(t *testing.T) {
	t.Parallel()
	t.Run("shall create a user and issue tokens", testSigninAnonymFlowUserDidNotExist)
	t.Run("shall fetch existing user and issue tokens", testSigninAnonymFlowUserExisted)
	t.Run("shall fetch deactivated user and fail", testSigninAnonymFlowUserDeactivated)
	t.Run("shall fail if no fingerprint was provided", testSigninAnonymFlowMissingRequiredInput)
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

func testSigninAnonymFlowUserExisted(t *testing.T) {
	// GIVEN
	const fingerprint = "foo"
	tokenSignClient := MockTokenSigningClient{
		Alg:       "EdDSA",
		Signature: "qux",
	}

	u := &User{
		ID:          "4fa6ecab-1029-42aa-bce7-99800d6eb630",
		Fingerprint: fingerprint,
		IsActive:    true,
	}

	repoClient := &MockRepositoryCIAM{
		UserID: map[string]*User{
			"4fa6ecab-1029-42aa-bce7-99800d6eb630": u,
		},
		UserFingerprint: map[string]*User{
			fingerprint: u,
		},
	}

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
	if tokens.id.UserID() != u.ID {
		t.Errorf("sub was set to a faulty value")
	}
	if fingerprint != tokens.id.UserDeviceFingerprint() {
		t.Errorf("user's fingerprint was set incorrectly")
	}
	if tokens.access.UserRole() != roleAnonymUser {
		t.Errorf("user's role was set incorrectly")
	}
	if tokens.id.UserEmail() != "" {
		t.Errorf("user's email was set incorrectly")
	}
}

func testSigninAnonymFlowUserDeactivated(t *testing.T) {
	// GIVEN
	const fingerprint = "foo"
	tokenSignClient := MockTokenSigningClient{
		Alg:       "EdDSA",
		Signature: "qux",
	}

	u := &User{
		ID:          "4fa6ecab-1029-42aa-bce7-99800d6eb630",
		Fingerprint: fingerprint,
		IsActive:    false,
	}

	repoClient := &MockRepositoryCIAM{
		UserID: map[string]*User{
			"4fa6ecab-1029-42aa-bce7-99800d6eb630": u,
		},
		UserFingerprint: map[string]*User{
			fingerprint: u,
		},
	}

	smtpClient := SMTPClient(nil)

	ciamClient := NewClient(repoClient, tokenSignClient, smtpClient)

	// WHEN
	tokens, err := ciamClient.SigninAnonym(context.TODO(), fingerprint)

	// THEN
	if !reflect.DeepEqual(err, errors.New("user was deactivated")) {
		t.Errorf("unexpected error")
	}
	if !reflect.DeepEqual(tokens, Tokens{}) {
		t.Errorf("unexpected happy path's result")
	}
}

func testSigninAnonymFlowMissingRequiredInput(t *testing.T) {
	// GIVEN
	const fingerprint = ""
	ciamClient := NewClient(nil, nil, nil)

	// WHEN
	tokens, err := ciamClient.SigninAnonym(context.TODO(), fingerprint)

	// THEN
	if !reflect.DeepEqual(
		err, errors.New("fingerprint must be provided"),
	) {
		t.Errorf("unexpected error")
	}
	if !reflect.DeepEqual(tokens, Tokens{}) {
		t.Errorf("unexpected happy path's result")
	}
}

func TestSigninUserFlow(t *testing.T) {
	t.Parallel()
	t.Run("shall create a user, generate and send a secret and issue ID token", testSigninUserFlowUserDidNotExist)
	t.Run(
		"shall fetch existing user, generate and send a secret and issue ID token",
		testSigninUserFlowUserExisted,
	)
	t.Run("shall fetch deactivated user and fail", testSigninUserFlowUserDeactivated)
	t.Run("shall fail if no email was provided", testSigninUserFlowMissingRequiredInput)
	t.Run(
		"shall fetch existing user, and existing non-expired secret and issue ID token",
		testSigninUserFlowUserExistedValidSecretExisted,
	)
}

func testSigninUserFlowUserExisted(t *testing.T) {
	t.Skip("todo")
}

func testSigninUserFlowUserDidNotExist(t *testing.T) {
	// GIVEN
	const (
		fingerprint = "foo"
		email       = "bar@baz.quxx"
	)
	tokenSignClient := MockTokenSigningClient{
		Alg:       "EdDSA",
		Signature: "qux",
	}

	repoClient := &MockRepositoryCIAM{}

	smtpClient := &MockSMTPClient{}

	ciamClient := NewClient(repoClient, tokenSignClient, smtpClient)

	// WHEN
	tokenID, err := ciamClient.SigninUser(context.TODO(), email, fingerprint)

	// THEN
	if err != nil {
		t.Errorf("unexpected error")
	}

	validateToken(t, tokenID, "id", tokenSignClient, defaultExpirationDurationIdentitySec)

	if err := utils.ValidateUUID(tokenID.UserID()); err != nil {
		t.Errorf("wrong sub format: %+v", err)
	}

	found, isActive, emailVerified, gotEmail, gotFingerprint, err := repoClient.ReadUser(
		context.TODO(), tokenID.UserID(),
	)
	if err != nil {
		t.Errorf("unexpected error: CIAM repository")
	}
	if !found {
		t.Errorf("user was not recorded to the CIAM repository")
	}
	if isActive {
		t.Errorf("user's activity was not set corretly, false expected")
	}
	if emailVerified {
		t.Errorf("user email's verification was not set corretly, false expected")
	}
	if email != gotEmail {
		t.Errorf("user's email was set incorrectly")
	}
	if fingerprint != gotFingerprint || fingerprint != tokenID.UserDeviceFingerprint() {
		t.Errorf("user's fingerprint was set incorrectly")
	}
	if tokenID.UserEmail() != email {
		t.Errorf("user's email was set incorrectly")
	}

	if smtpClient.Secret == "" {
		t.Errorf("one-time secret was generated and sent")
	}
	if smtpClient.Recipient != email {
		t.Errorf("one-time secret was not sent to the right recipient")
	}
}

func testSigninUserFlowUserDeactivated(t *testing.T) {
	t.Skip("todo")
}

func testSigninUserFlowMissingRequiredInput(t *testing.T) {
	t.Skip("todo")
}

func testSigninUserFlowUserExistedValidSecretExisted(t *testing.T) {
	t.Skip("todo")
}
