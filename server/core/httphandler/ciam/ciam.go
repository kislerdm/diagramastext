// Package ciam to authN/Z users
package ciam

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/kislerdm/diagramastext/server/core/internal/utils"
)

// Client defines the CIAM client.
type Client interface {
	// SigninUser executes user's authentication flow.
	SigninUser(ctx context.Context, email, fingerprint string) (identityToken JWT, err error)

	// SigninAnonym executes anonym's authentication flow.
	SigninAnonym(ctx context.Context, fingerprint string) (Tokens, error)

	// IssueTokensAfterSecretConfirmation validates user's confirmation.
	IssueTokensAfterSecretConfirmation(ctx context.Context, identityToken, secret string) (Tokens, error)

	// RefreshTokens refreshes access token given the refresh token.
	RefreshTokens(ctx context.Context, refreshToken string) (Tokens, error)

	// ValidateToken validates JWT.
	ValidateToken(ctx context.Context, token string) error
}

// RepositoryCIAM defines the communication port to persistence layer hosting users' data.
type RepositoryCIAM interface {
	CreateUser(ctx context.Context, id, email, fingerprint string) error
	ReadUser(ctx context.Context, id string) (
		found, isActive, isPremium, emailVerified bool, email, fingerprint string, err error,
	)

	LookupUserByEmail(ctx context.Context, email string) (id string, isActive bool, err error)
	LookupUserByFingerprint(ctx context.Context, fingerprint string) (id string, isActive bool, err error)

	UpdateUserSetActiveStatus(ctx context.Context, id string, isActive bool) error
	UpdateUserSetEmailVerified(ctx context.Context, id string) error

	CreateOneTimeSecret(ctx context.Context, userID, secret string, createdAt time.Time) error
	ReadOneTimeSecret(ctx context.Context, userID string) (secret string, issuedAt time.Time, err error)
	DeleteOneTimeSecret(ctx context.Context, userID string) error
}

// NewClient initializes the CIAM client.
func NewClient(clientRepository RepositoryCIAM, clientKMS KMSClient, clientEmail SMTPClient) Client {
	return &client{
		clientRepository: clientRepository,
		clientKMS:        clientKMS,
		clientEmail:      clientEmail,
	}
}

type client struct {
	clientRepository RepositoryCIAM
	clientKMS        KMSClient
	clientEmail      SMTPClient
}

func (c client) IssueTokensAfterSecretConfirmation(ctx context.Context, identityToken, secret string) (Tokens, error) {
	t, err := ParseToken(identityToken)
	if err != nil {
		return Tokens{}, err
	}
	if err := t.Validate(
		func(signingString, signature string) error {
			return c.clientKMS.Verify(ctx, signingString, signature)
		},
	); err != nil {
		return Tokens{}, err
	}

	secretRef, _, err := c.clientRepository.ReadOneTimeSecret(ctx, t.Sub())
	if err != nil {
		return Tokens{}, err
	}

	if secret != secretRef {
		return Tokens{}, errors.New("secret is wrong")
	}

	if err := c.clientRepository.UpdateUserSetActiveStatus(ctx, t.Sub(), true); err != nil {
		return Tokens{}, err
	}

	if err := c.clientRepository.UpdateUserSetEmailVerified(ctx, t.Sub()); err != nil {
		return Tokens{}, err
	}

	return c.issueTokens(ctx, t.Sub(), t.Email(), t.Fingerprint(), false)
}

// SigninAnonym executes anonym's authentication flow:
//
//	Fingerprint found in DB -> No  -> Create \
//							-> Yes ->  --	-> Generate Refresh and Access JWT -> Return generates JWT.
func (c client) SigninAnonym(ctx context.Context, fingerprint string) (Tokens, error) {
	if fingerprint == "" {
		return Tokens{}, errors.New("fingerprint must be provided")
	}

	var (
		userID string
		err    error
	)

	userID, isActive, err := c.clientRepository.LookupUserByFingerprint(ctx, fingerprint)
	if err != nil {
		return Tokens{}, err
	}

	if userID != "" && !isActive {
		return Tokens{}, errors.New("user was deactivated")
	}

	if userID == "" {
		userID = utils.NewUUID()
		if err := c.clientRepository.CreateUser(ctx, userID, "", fingerprint); err != nil {
			return Tokens{}, err
		}
	}

	return c.issueTokens(ctx, userID, "", fingerprint, false)
}

func (c client) issueTokens(ctx context.Context, userID, email, fingerprint string, isPremium bool) (Tokens, error) {
	iat := time.Now().UTC()
	opts := []OptFn{
		WithCustomIat(iat), WithSignature(
			func(signingString string) (signature string, alg string, err error) {
				return c.clientKMS.Sign(ctx, signature)
			},
		),
	}
	idToken, err := NewIDToken(userID, email, fingerprint, 0, opts...)
	if err != nil {
		return Tokens{}, err
	}
	refreshToken, err := NewRefreshToken(userID, opts...)
	if err != nil {
		return Tokens{}, err
	}
	accessToken, err := NewAccessToken(userID, isPremium, opts...)
	if err != nil {
		return Tokens{}, err
	}
	return Tokens{
		ID:      idToken,
		Refresh: refreshToken,
		Access:  accessToken,
	}, nil
}

// SigninUser executes user's authentication flow:
//
//	Email found in DB -> No  -> Create \
//			 	   	  -> Yes ->	--	  -> Generate secret and ID JWT -> Send secret to email -> Return ID JWT.
func (c client) SigninUser(ctx context.Context, email, fingerprint string) (JWT, error) {
	if email == "" {
		return nil, errors.New("email must be provided")
	}

	var (
		userID string
		err    error
	)

	userID, _, err = c.clientRepository.LookupUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	switch userID == "" {
	case true:
		userID = utils.NewUUID()
		if err := c.clientRepository.CreateUser(ctx, userID, email, fingerprint); err != nil {
			return nil, err
		}
	default:
		_, iat, err := c.clientRepository.ReadOneTimeSecret(ctx, userID)
		if err != nil {
			return nil, err
		}
		if iat.Add(time.Duration(defaultExpirationDurationIdentitySec)).After(time.Now().UTC()) {
			return NewIDToken(
				userID, email, fingerprint, 0, WithCustomIat(iat), WithSignature(
					func(signingString string) (signature string, alg string, err error) {
						return c.clientKMS.Sign(ctx, signature)
					},
				),
			)
		}
		if err := c.clientRepository.DeleteOneTimeSecret(ctx, userID); err != nil {
			return nil, err
		}
	}

	secret := generateOnetimeSecret()
	iat := time.Now().UTC()

	if err := c.clientEmail.SendSignInEmail(email, secret); err != nil {
		return nil, err
	}
	if err := c.clientRepository.CreateOneTimeSecret(ctx, userID, secret, iat); err != nil {
		return nil, err
	}

	return NewIDToken(
		userID, email, fingerprint, 0, WithCustomIat(iat), WithSignature(
			func(signingString string) (signature string, alg string, err error) {
				return c.clientKMS.Sign(ctx, signature)
			},
		),
	)
}

func generateOnetimeSecret() string {
	const (
		charset = "0123456789abcdef"
		length  = 6
	)
	var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	var b = make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func (c client) ValidateToken(ctx context.Context, token string) error {
	t, err := ParseToken(token)
	if err != nil {
		return err
	}
	return t.Validate(
		func(signingString, signature string) error {
			return c.clientKMS.Verify(ctx, signingString, signature)
		},
	)
}

func (c client) RefreshTokens(ctx context.Context, refreshToken string) (Tokens, error) {
	t, err := ParseToken(refreshToken)
	if err != nil {
		return Tokens{}, err
	}
	if err := t.Validate(
		func(signingString, signature string) error {
			return c.clientKMS.Verify(ctx, signingString, signature)
		},
	); err != nil {
		return Tokens{}, err
	}
	found, isActive, isPremium, _, email, fingerprint, err := c.clientRepository.ReadUser(ctx, t.Sub())
	if err != nil {
		return Tokens{}, err
	}
	if !found {
		return Tokens{}, errors.New("user not found")
	}
	if !isActive {
		return Tokens{}, errors.New("user was deactivated")
	}
	return c.issueTokens(ctx, t.Sub(), email, fingerprint, isPremium)
}
