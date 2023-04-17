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
	SigninUser(ctx context.Context, email, fingerprint string) (Tokens, error)

	// SigninAnonym executes anonym's authentication flow.
	SigninAnonym(ctx context.Context, fingerprint string) (Tokens, error)

	// RefreshAccessToken refreshes access token given the refresh token.
	RefreshAccessToken(ctx context.Context, refreshToken string) (JWT, error)

	// ValidateAccessToken validates JWT.
	ValidateAccessToken(ctx context.Context, token string) error
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
	ReadOneTimeSecret(ctx context.Context, userID string) (string, time.Time, error)
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

	iat := time.Now().UTC()

	opts := []OptFn{
		WithCustomIat(iat), WithSignature(
			func(signingString string) (signature string, alg string, err error) {
				return c.clientKMS.Sign(ctx, signature)
			},
		),
	}
	refreshToken, err := NewRefreshToken(userID, opts...)
	if err != nil {
		return Tokens{}, err
	}

	accessToken, err := NewAccessToken(userID, false, opts...)
	if err != nil {
		return Tokens{}, err
	}

	return Tokens{
		Refresh: refreshToken,
		Access:  accessToken,
	}, nil
}

// SigninUser executes user's authentication flow:
//
//	Email found in DB -> No  -> Create \
//			 	   	  -> Yes ->	--	  -> Generate secret and ID JWT -> Send secret to email -> Return ID JWT.
func (c client) SigninUser(ctx context.Context, email, fingerprint string) (Tokens, error) {
	if email == "" {
		return Tokens{}, errors.New("email must be provided")
	}

	var (
		userID string
		err    error
	)

	userID, _, err = c.clientRepository.LookupUserByEmail(ctx, email)
	if err != nil {
		return Tokens{}, err
	}

	switch userID == "" {
	case true:
		userID = utils.NewUUID()
		if err := c.clientRepository.CreateUser(ctx, userID, email, fingerprint); err != nil {
			return Tokens{}, err
		}
	default:
		_, iat, err := c.clientRepository.ReadOneTimeSecret(ctx, userID)
		if err != nil {
			return Tokens{}, err
		}
		if iat.Add(time.Duration(expirationDurationIdentitySec)).After(time.Now().UTC()) {
			idToken, err := NewIDToken(
				userID, email, fingerprint, WithCustomIat(iat), WithSignature(
					func(signingString string) (signature string, alg string, err error) {
						return c.clientKMS.Sign(ctx, signature)
					},
				),
			)
			if err != nil {
				return Tokens{}, err
			}
			return Tokens{
				ID: idToken,
			}, nil
		}
		if err := c.clientRepository.DeleteOneTimeSecret(ctx, userID); err != nil {
			return Tokens{}, err
		}
	}

	secret := generateOnetimeSecret()
	iat := time.Now().UTC()

	if err := c.clientEmail.SendSignInEmail(email, secret); err != nil {
		return Tokens{}, err
	}
	if err := c.clientRepository.CreateOneTimeSecret(ctx, userID, secret, iat); err != nil {
		return Tokens{}, err
	}

	idToken, err := NewIDToken(
		userID, email, fingerprint, WithCustomIat(iat), WithSignature(
			func(signingString string) (signature string, alg string, err error) {
				return c.clientKMS.Sign(ctx, signature)
			},
		),
	)
	if err != nil {
		return Tokens{}, err
	}

	return Tokens{
		ID: idToken,
	}, nil
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

func (c client) ValidateAccessToken(ctx context.Context, accessToken string) error {
	t, err := ParseToken(accessToken)
	if err != nil {
		return err
	}
	return t.Validate(
		func(signingString, signature string) error {
			return c.clientKMS.Verify(ctx, signingString, signature)
		},
	)
}

func (c client) RefreshAccessToken(ctx context.Context, refreshToken string) (JWT, error) {
	t, err := ParseToken(refreshToken)
	if err != nil {
		return nil, err
	}
	if err := t.Validate(
		func(signingString, signature string) error {
			return c.clientKMS.Verify(ctx, signingString, signature)
		},
	); err != nil {
		return nil, err
	}

	found, isActive, isPremium, _, _, _, err := c.clientRepository.ReadUser(ctx, t.Sub())
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, errors.New("no user exists")
	}
	if !isActive {
		return nil, errors.New("user was deactivated")
	}

	return NewAccessToken(
		t.Sub(), isPremium, WithSignature(
			func(signingString string) (signature string, alg string, err error) {
				return c.clientKMS.Sign(ctx, signingString)
			},
		),
	)
}
