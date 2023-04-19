package ciam

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/kislerdm/diagramastext/server/core/internal/utils"
)

func mustNewToken(token JWT, err error) JWT {
	if err != nil {
		panic(err)
	}
	return token
}

func mustNewTokenStr(token JWT, err error) string {
	s, err := mustNewToken(token, err).String()
	if err != nil {
		panic(err)
	}
	return s
}

func validateToken(t *testing.T, tkn JWT, tokenTyp string, clientSign TokenSigningClient, wantTTL int64) {
	tk, ok := tkn.(*token)
	if !ok {
		t.Errorf("%s wrong token format", tokenTyp)
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

	if clientSign != nil {
		if err := tkn.Validate(
			func(signingString, signature string) error {
				return clientSign.Verify(context.TODO(), signingString, signature)
			},
		); err != nil {
			t.Errorf("%s wrong signrature: %+v", tokenTyp, err)
		}
		if c, ok := clientSign.(MockTokenSigningClient); ok {
			if tk.header.Alg != c.Alg {
				t.Errorf("%s wrong header: alg", tokenTyp)
			}
		}
	}
}

func Test_client_SigninAnonymFlow(t *testing.T) {
	t.Parallel()
	t.Run("shall create a user and issue tokens", testSigninAnonymFlowUserDidNotExist)
	t.Run("shall fetch existing user and issue tokens", testSigninAnonymFlowUserExisted)
	t.Run("shall fetch deactivated user and fail", testSigninAnonymFlowUserDeactivated)
	t.Run("shall fail if no fingerprint was provided", testSigninAnonymFlowMissingRequiredInput)
	t.Run("shall fail because of faulty interaction with CIAM repo", testSigninAnonymFlowFailedRepoInterface)
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

func testSigninAnonymFlowFailedRepoInterface(t *testing.T) {
	// GIVEN
	const fingerprint = "foo"

	repoComsError := errors.New("foobar")
	repoClient := &MockRepositoryCIAM{
		Err: repoComsError,
	}

	tokenSignClient := TokenSigningClient(nil)
	smtpClient := SMTPClient(nil)

	ciamClient := NewClient(repoClient, tokenSignClient, smtpClient)

	// WHEN
	tokens, err := ciamClient.SigninAnonym(context.TODO(), fingerprint)

	// THEN
	if !reflect.DeepEqual(err, repoComsError) {
		t.Errorf("unexpected error")
	}

	if !reflect.DeepEqual(tokens, Tokens{}) {
		t.Errorf("unexpected happy path's result")
	}
}

func Test_client_SigninUserFlow(t *testing.T) {
	t.Parallel()
	t.Run("shall create a user, generate and send a secret and issue ID token", testSigninUserFlowUserDidNotExist)
	t.Run(
		"shall fetch existing user, generate and send a secret and issue ID token",
		testSigninUserFlowActiveUserExisted,
	)
	t.Run("shall fetch deactivated user and fail", testSigninUserFlowDeactivatedUserExisted)
	t.Run("shall fail if no email was provided", testSigninUserFlowMissingRequiredInput)
	t.Run(
		"shall fetch existing user, existing non-expired secret and issue ID token",
		testSigninUserFlowUserExistedValidSecretExisted,
	)
	t.Run(
		"shall fetch existing user and expired secret, generate and sent new secret, and issue ID token",
		testSigninUserFlowUserExistedExpiredSecretExisted,
	)
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

	found, secret, _, _ := repoClient.ReadOneTimeSecret(context.TODO(), tokenID.UserID())
	if !found || smtpClient.Secret != secret {
		t.Errorf("one-time secret was not persisted")
	}
}

func testSigninUserFlowDeactivatedUserExisted(t *testing.T) {
	// GIVEN
	const (
		fingerprint = "foo"
		email       = "bar@baz.quxx"
	)
	tokenSignClient := MockTokenSigningClient{
		Alg:       "EdDSA",
		Signature: "qux",
	}

	u := &User{
		ID:            "4fa6ecab-1029-42aa-bce7-99800d6eb630",
		Email:         email,
		Fingerprint:   fingerprint,
		IsActive:      false,
		EmailVerified: true,
	}

	repoClient := &MockRepositoryCIAM{
		UserID: map[string]*User{
			"4fa6ecab-1029-42aa-bce7-99800d6eb630": u,
		},
		UserEmail: map[string]*User{
			email: u,
		},
		UserFingerprint: map[string]*User{
			fingerprint: u,
		},
	}

	smtpClient := &MockSMTPClient{}

	ciamClient := NewClient(repoClient, tokenSignClient, smtpClient)

	// WHEN
	tokenID, err := ciamClient.SigninUser(context.TODO(), email, fingerprint)

	// THEN
	if !reflect.DeepEqual(err, errors.New("user was deactivated")) {
		t.Errorf("unexpected error")
	}
	if tokenID != nil {
		t.Errorf("unexpected happy path's result")
	}
}

func testSigninUserFlowMissingRequiredInput(t *testing.T) {
	// GIVEN
	const email = ""
	ciamClient := NewClient(nil, nil, nil)

	// WHEN
	tokenID, err := ciamClient.SigninUser(context.TODO(), email, "")

	// THEN
	if !reflect.DeepEqual(
		err, errors.New("email must be provided"),
	) {
		t.Errorf("unexpected error")
	}
	if tokenID != nil {
		t.Errorf("unexpected happy path's result")
	}
}

func testSigninUserFlowActiveUserExisted(t *testing.T) {
	// GIVEN
	const (
		fingerprint = "foo"
		email       = "bar@baz.quxx"
		userID      = "4fa6ecab-1029-42aa-bce7-99800d6eb630"
	)
	tokenSignClient := MockTokenSigningClient{
		Alg:       "EdDSA",
		Signature: "qux",
	}

	u := &User{
		ID:            userID,
		Email:         email,
		Fingerprint:   fingerprint,
		IsActive:      true,
		EmailVerified: true,
	}

	repoClient := &MockRepositoryCIAM{
		UserID: map[string]*User{
			userID: u,
		},
		UserEmail: map[string]*User{
			email: u,
		},
		UserFingerprint: map[string]*User{
			fingerprint: u,
		},
	}

	smtpClient := &MockSMTPClient{}

	ciamClient := NewClient(repoClient, tokenSignClient, smtpClient)

	// WHEN
	tokenID, err := ciamClient.SigninUser(context.TODO(), email, fingerprint)

	// THEN
	if err != nil {
		t.Errorf("unexpected error")
	}

	validateToken(t, tokenID, "id", tokenSignClient, defaultExpirationDurationIdentitySec)

	if tokenID.UserID() != userID {
		t.Error("wrong userID was set format")
	}

	if fingerprint != tokenID.UserDeviceFingerprint() {
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

	found, secret, _, _ := repoClient.ReadOneTimeSecret(context.TODO(), tokenID.UserID())
	if !found || smtpClient.Secret != secret {
		t.Errorf("one-time secret was not persisted")
	}
}

func testSigninUserFlowUserExistedValidSecretExisted(t *testing.T) {
	// GIVEN
	const (
		fingerprint = "foo"
		email       = "bar@baz.quxx"
		userID      = "4fa6ecab-1029-42aa-bce7-99800d6eb630"
	)
	tokenSignClient := MockTokenSigningClient{
		Alg:       "EdDSA",
		Signature: "qux",
	}

	u := &User{
		ID:            userID,
		Email:         email,
		Fingerprint:   fingerprint,
		IsActive:      true,
		EmailVerified: true,
	}

	secret := Secret{
		Secret:   "foobar",
		IssuedAt: time.Now().UTC(),
	}

	repoClient := &MockRepositoryCIAM{
		UserID: map[string]*User{
			userID: u,
		},
		UserEmail: map[string]*User{
			email: u,
		},
		UserFingerprint: map[string]*User{
			fingerprint: u,
		},
		Secret: map[string]Secret{
			userID: secret,
		},
	}

	smtpClient := &MockSMTPClient{}

	ciamClient := NewClient(repoClient, tokenSignClient, smtpClient)

	// WHEN
	tokenID, err := ciamClient.SigninUser(context.TODO(), email, fingerprint)

	// THEN
	if err != nil {
		t.Errorf("unexpected error")
	}

	validateToken(t, tokenID, "id", tokenSignClient, defaultExpirationDurationIdentitySec)

	if tokenID.UserID() != userID {
		t.Error("wrong userID was set format")
	}

	if fingerprint != tokenID.UserDeviceFingerprint() {
		t.Errorf("user's fingerprint was set incorrectly")
	}
	if tokenID.UserEmail() != email {
		t.Errorf("user's email was set incorrectly")
	}

	if smtpClient.Secret != "" || smtpClient.Recipient != "" {
		t.Errorf("valid one-time secret was re-sent by mistake")
	}

	found, gotSecretStr, _, _ := repoClient.ReadOneTimeSecret(context.TODO(), tokenID.UserID())
	if !found || secret.Secret != gotSecretStr {
		t.Errorf("one-time secret was removed from the repo")
	}
}

func testSigninUserFlowUserExistedExpiredSecretExisted(t *testing.T) {
	// GIVEN
	const (
		fingerprint          = "foo"
		email                = "bar@baz.quxx"
		userID               = "4fa6ecab-1029-42aa-bce7-99800d6eb630"
		tokenIssuedTimeShift = -24 * 1000 * time.Hour
	)
	tokenSignClient := MockTokenSigningClient{
		Alg:       "EdDSA",
		Signature: "qux",
	}

	u := &User{
		ID:            userID,
		Email:         email,
		Fingerprint:   fingerprint,
		IsActive:      true,
		EmailVerified: true,
	}

	secret := Secret{
		Secret:   "foobar",
		IssuedAt: time.Now().UTC().Add(tokenIssuedTimeShift),
	}

	repoClient := &MockRepositoryCIAM{
		UserID: map[string]*User{
			userID: u,
		},
		UserEmail: map[string]*User{
			email: u,
		},
		UserFingerprint: map[string]*User{
			fingerprint: u,
		},
		Secret: map[string]Secret{
			userID: secret,
		},
	}

	smtpClient := &MockSMTPClient{}

	ciamClient := NewClient(repoClient, tokenSignClient, smtpClient)

	// WHEN
	tokenID, err := ciamClient.SigninUser(context.TODO(), email, fingerprint)

	// THEN
	if err != nil {
		t.Errorf("unexpected error")
	}

	validateToken(t, tokenID, "id", tokenSignClient, defaultExpirationDurationIdentitySec)

	if tokenID.UserID() != userID {
		t.Error("wrong userID was set format")
	}

	if fingerprint != tokenID.UserDeviceFingerprint() {
		t.Errorf("user's fingerprint was set incorrectly")
	}
	if tokenID.UserEmail() != email {
		t.Errorf("user's email was set incorrectly")
	}

	if smtpClient.Secret == "" || smtpClient.Recipient == "" {
		t.Errorf("one-time secret was not generated, or sent")
	}

	found, gotSecretStr, _, _ := repoClient.ReadOneTimeSecret(context.TODO(), tokenID.UserID())
	if !found || secret.Secret == "" || secret.Secret == gotSecretStr {
		t.Errorf("one-time secret was not re-generated and stored to the repo")
	}
}

func Test_client_IssueTokensAfterSecretConfirmationHappyPath(t *testing.T) {
	// GIVEN
	const (
		fingerprint = "foo"
		email       = "bar@baz.quxx"
		userID      = "4fa6ecab-1029-42aa-bce7-99800d6eb630"
		secretVal   = "foobar"
	)
	tokenSignClient := MockTokenSigningClient{
		Alg:       "EdDSA",
		Signature: "qux",
	}

	u := &User{
		ID:          userID,
		Email:       email,
		Fingerprint: fingerprint,
	}

	iat := time.Now().UTC()

	tID, err := NewIDToken(
		u.ID, u.Email, u.Fingerprint, u.EmailVerified, 0,
		WithCustomIat(iat), WithSignature(
			func(signingString string) (signature string, alg string, err error) {
				return tokenSignClient.Sign(context.TODO(), signingString)
			},
		),
	)
	if err != nil {
		panic(err)
	}
	tokenID, err := tID.String()
	if err != nil {
		panic(err)
	}

	secret := Secret{
		Secret:   secretVal,
		IssuedAt: iat,
	}

	repoClient := &MockRepositoryCIAM{
		UserID: map[string]*User{
			userID: u,
		},
		UserEmail: map[string]*User{
			email: u,
		},
		UserFingerprint: map[string]*User{
			fingerprint: u,
		},
		Secret: map[string]Secret{
			userID: secret,
		},
	}

	smtpClient := &MockSMTPClient{}

	ciamClient := NewClient(repoClient, tokenSignClient, smtpClient)

	// WHEN
	tokens, err := ciamClient.IssueTokensAfterSecretConfirmation(context.TODO(), tokenID, secretVal)

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

	if tokens.id.UserID() != userID {
		t.Error("wrong userID was set in the tokens")
	}

	found, isActive, emailVerified, gotEmail, gotFingerprint, err := repoClient.ReadUser(
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
	if !emailVerified {
		t.Errorf("user email's verification was not set corretly, true expected")
	}
	if email != tokens.id.UserEmail() || email != gotEmail {
		t.Errorf("user's email was persisted incorrectly")
	}
	if fingerprint != tokens.id.UserDeviceFingerprint() || fingerprint != gotFingerprint {
		t.Errorf("user's fingerprint was set incorrectly")
	}

	found, gotSecretStr, _, _ := repoClient.ReadOneTimeSecret(context.TODO(), userID)
	if found || secret.Secret == gotSecretStr {
		t.Errorf("one-time secret was not removed from the repo")
	}
}

func Test_client_ValidateToken(t *testing.T) {
	type fields struct {
		clientRepository RepositoryCIAM
		clientKMS        TokenSigningClient
		clientEmail      SMTPClient
	}
	type args struct {
		ctx   context.Context
		token string
	}

	signingClient := MockTokenSigningClient{
		Alg:       "EdDSA",
		Signature: "qux",
	}

	const userID = "4fa6ecab-1029-42aa-bce7-99800d6eb630"

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "token valid",
			fields: fields{
				clientKMS: signingClient,
			},
			args: args{
				ctx: context.TODO(),
				token: mustNewTokenStr(
					NewAccessToken(
						userID, true, WithSignature(
							func(signingString string) (signature string, alg string, err error) {
								return signingClient.Sign(context.TODO(), signingString)
							},
						),
					),
				),
			},
			wantErr: false,
		},
		{
			name: "token invalid: signature",
			fields: fields{
				clientKMS: signingClient,
			},
			args: args{
				ctx: context.TODO(),
				token: mustNewTokenStr(
					NewAccessToken(userID, true),
				),
			},
			wantErr: true,
		},
		{
			name: "token invalid: corrupt JWT",
			fields: fields{
				clientKMS: signingClient,
			},
			args: args{
				ctx:   context.TODO(),
				token: "foo",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c := client{
					clientRepository: tt.fields.clientRepository,
					clientKMS:        tt.fields.clientKMS,
					clientEmail:      tt.fields.clientEmail,
				}
				if err := c.ValidateToken(tt.args.ctx, tt.args.token); (err != nil) != tt.wantErr {
					t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
				}
			},
		)
	}
}

func Test_client_RefreshTokens(t *testing.T) {
	type fields struct {
		clientRepository RepositoryCIAM
		clientKMS        TokenSigningClient
		clientEmail      SMTPClient
	}
	type args struct {
		ctx          context.Context
		refreshToken string
	}
	const (
		fingerprint = "foo"
		email       = "bar@baz.quxx"
		userID      = "4fa6ecab-1029-42aa-bce7-99800d6eb630"
	)

	signingClient := MockTokenSigningClient{
		Alg:       "EdDSA",
		Signature: "qux",
	}

	t.Parallel()

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Tokens
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				clientRepository: &MockRepositoryCIAM{
					UserID: map[string]*User{
						userID: {
							ID:            userID,
							Email:         email,
							Fingerprint:   fingerprint,
							IsActive:      true,
							EmailVerified: true,
						},
					},
				},
				clientKMS: signingClient,
			},
			args: args{
				ctx: context.TODO(),
				refreshToken: mustNewTokenStr(
					NewRefreshToken(
						userID, WithSignature(
							func(signingString string) (signature string, alg string, err error) {
								return signingClient.Sign(context.TODO(), signingString)
							},
						),
					),
				),
			},
			wantErr: false,
		},
		{
			name: "unhappy path: user deactivated",
			fields: fields{
				clientRepository: &MockRepositoryCIAM{
					UserID: map[string]*User{
						userID: {
							ID:            userID,
							Email:         email,
							Fingerprint:   fingerprint,
							IsActive:      false,
							EmailVerified: true,
						},
					},
				},
				clientKMS: signingClient,
			},
			args: args{
				ctx: context.TODO(),
				refreshToken: mustNewTokenStr(
					NewRefreshToken(
						userID, WithSignature(
							func(signingString string) (signature string, alg string, err error) {
								return signingClient.Sign(context.TODO(), signingString)
							},
						),
					),
				),
			},
			wantErr: true,
		},
		{
			name: "unhappy path: user not found",
			fields: fields{
				clientRepository: &MockRepositoryCIAM{},
				clientKMS:        signingClient,
			},
			args: args{
				ctx: context.TODO(),
				refreshToken: mustNewTokenStr(
					NewRefreshToken(
						userID, WithSignature(
							func(signingString string) (signature string, alg string, err error) {
								return signingClient.Sign(context.TODO(), signingString)
							},
						),
					),
				),
			},
			wantErr: true,
		},
		{
			name: "unhappy path: registered user's email was not confirmed",
			fields: fields{
				clientRepository: &MockRepositoryCIAM{
					UserID: map[string]*User{
						userID: {
							ID:            userID,
							Email:         email,
							Fingerprint:   fingerprint,
							IsActive:      true,
							EmailVerified: false,
						},
					},
				},
				clientKMS: signingClient,
			},
			args: args{
				ctx: context.TODO(),
				refreshToken: mustNewTokenStr(
					NewRefreshToken(
						userID, WithSignature(
							func(signingString string) (signature string, alg string, err error) {
								return signingClient.Sign(context.TODO(), signingString)
							},
						),
					),
				),
			},
			wantErr: true,
		},
		{
			name: "unhappy path: failed interactions with CIAM repo",
			fields: fields{
				clientRepository: &MockRepositoryCIAM{
					Err: errors.New("foobar"),
				},
				clientKMS: signingClient,
			},
			args: args{
				ctx: context.TODO(),
				refreshToken: mustNewTokenStr(
					NewRefreshToken(
						userID, WithSignature(
							func(signingString string) (signature string, alg string, err error) {
								return signingClient.Sign(context.TODO(), signingString)
							},
						),
					),
				),
			},
			wantErr: true,
		},
		{
			name: "unhappy path: invalid token's signature",
			fields: fields{
				clientRepository: &MockRepositoryCIAM{
					Err: errors.New("foobar"),
				},
				clientKMS: signingClient,
			},
			args: args{
				ctx: context.TODO(),
				refreshToken: mustNewTokenStr(
					NewRefreshToken(
						userID, WithSignature(
							func(_ string) (signature string, alg string, err error) {
								return "foo", "bar", nil
							},
						),
					),
				),
			},
			wantErr: true,
		},
		{
			name: "unhappy path: invalid input refresh token",
			args: args{
				ctx:          context.TODO(),
				refreshToken: "foo",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c := client{
					clientRepository: tt.fields.clientRepository,
					clientKMS:        tt.fields.clientKMS,
					clientEmail:      tt.fields.clientEmail,
				}
				got, err := c.RefreshTokens(tt.args.ctx, tt.args.refreshToken)
				if (err != nil) != tt.wantErr {
					t.Errorf("RefreshTokens() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr {
					validateToken(
						t, got.id, "id", tt.fields.clientKMS,
						defaultExpirationDurationIdentitySec,
					)
					validateToken(
						t, got.refresh, "refresh",
						tt.fields.clientKMS, defaultExpirationDurationRefreshSec,
					)
					validateToken(
						t, got.access, "access",
						tt.fields.clientKMS, defaultExpirationDurationAccessSec,
					)

					if got.id.UserID() != got.access.UserID() || got.id.UserID() != got.refresh.UserID() {
						t.Errorf("userID/sub was set inconsistently across the tokens")
					}

					if got.id.UserID() != userID {
						t.Errorf("wront userID was set")
					}
				}
			},
		)
	}
}

func TestTokens_Serialize(t *testing.T) {
	type fields struct {
		id      JWT
		refresh JWT
		access  JWT
	}

	signingClient := MockTokenSigningClient{
		Alg:       "EdDSA",
		Signature: "qux",
	}

	const (
		fingerprint = "foo"
		email       = "bar@baz.quxx"
		userID      = "4fa6ecab-1029-42aa-bce7-99800d6eb630"
	)
	iat := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				id: mustNewToken(
					NewIDToken(
						userID, email, fingerprint, true, 0,
						WithCustomIat(iat),
						WithSignature(
							func(signingString string) (signature string, alg string, err error) {
								return signingClient.Sign(context.TODO(), signingString)
							},
						),
					),
				),
				refresh: mustNewToken(
					NewRefreshToken(
						userID, WithCustomIat(iat),
						WithSignature(
							func(signingString string) (signature string, alg string, err error) {
								return signingClient.Sign(context.TODO(), signingString)
							},
						),
					),
				),
				access: mustNewToken(
					NewAccessToken(
						userID, true, WithCustomIat(iat),
						WithSignature(
							func(signingString string) (signature string, alg string, err error) {
								return signingClient.Sign(context.TODO(), signingString)
							},
						),
					),
				),
			},
			want:    []byte(`{"id":"eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiZW1haWwiOiJiYXJAYmF6LnF1eHgiLCJmaW5nZXJwcmludCI6ImZvbyIsInN1YiI6IjRmYTZlY2FiLTEwMjktNDJhYS1iY2U3LTk5ODAwZDZlYjYzMCIsImlzcyI6Imh0dHBzOi8vY2lhbS5kaWFncmFtYXN0ZXh0LmRldiIsImF1ZCI6Imh0dHBzOi8vZGlhZ3JhbWFzdGV4dC5kZXYiLCJpYXQiOjAsImV4cCI6MzYwMH0.qux","refresh":"eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI0ZmE2ZWNhYi0xMDI5LTQyYWEtYmNlNy05OTgwMGQ2ZWI2MzAiLCJpc3MiOiJodHRwczovL2NpYW0uZGlhZ3JhbWFzdGV4dC5kZXYiLCJhdWQiOiJodHRwczovL2RpYWdyYW1hc3RleHQuZGV2IiwiaWF0IjowLCJleHAiOjg2NDAwMDB9.qux","access":"eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoxLCJzdWIiOiI0ZmE2ZWNhYi0xMDI5LTQyYWEtYmNlNy05OTgwMGQ2ZWI2MzAiLCJpc3MiOiJodHRwczovL2NpYW0uZGlhZ3JhbWFzdGV4dC5kZXYiLCJhdWQiOiJodHRwczovL2RpYWdyYW1hc3RleHQuZGV2IiwiaWF0IjowLCJleHAiOjM2MDB9.qux"}`),
			wantErr: false,
		},
	}

	t.Parallel()
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tkn := Tokens{
					id:      tt.fields.id,
					refresh: tt.fields.refresh,
					access:  tt.fields.access,
				}
				got, err := tkn.Serialize()
				if (err != nil) != tt.wantErr {
					t.Errorf("Serialize() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("Serialize() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}
