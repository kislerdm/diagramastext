// Package ciam to authN/Z users
package ciam

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/kislerdm/diagramastext/server/core/internal/utils"
)

// CIAMClient defines the CIAM client.
type CIAMClient interface {
	// SigninUser executes user's authentication flow.
	SigninUser(ctx context.Context, email, fingerprint string) (Tokens, error)

	// SigninAnonym executes anonym's authentication flow.
	SigninAnonym(ctx context.Context, fingerprint string) (Tokens, error)

	// RefreshAccessToken refreshes access token given the refresh token.
	RefreshAccessToken(ctx context.Context, refreshToken string) (JWT, error)

	// ValidateToken validates JWT.
	ValidateToken(ctx context.Context, token string) error
}

type KMSClient interface {
	Verify(ctx context.Context, token JWT) error
	Sign(ctx context.Context, token JWT) error
}

// RepositoryCIAM defines the communication port to persistence layer hosting users' data.
type RepositoryCIAM interface {
	GetUser(ctx context.Context, id string) (
		email string, fingerprint string, emailVerified, isActive bool, err error,
	)
	LookupUserByEmail(ctx context.Context, email string) (id string, err error)
	LookupUserByFingerprint(ctx context.Context, fingerprint string) (id string, err error)
	CreateUser(ctx context.Context, id, email, fingerprint string) error
	UpdateUserEmail(ctx context.Context, id, email string) error
	UpdateUserFignerprint(ctx context.Context, id, fingerprint string) error
	UpdateUserActiveStatus(ctx context.Context, id string, isActive bool) error
	SetUserEmailVerified(ctx context.Context, id string) error

	LookupOneTimeSecret(ctx context.Context, userID string) (time.Time, error)
	WriteOneTimeSecret(ctx context.Context, userID, secret string, createdAt time.Time) error
	DeleteOneTimeSecret(ctx context.Context, userID string) error
}

func NewCIAMClient(clientRepository RepositoryCIAM, clientKMS KMSClient, clientEmail SMTPClient) CIAMClient {
	return &client{
		clientRepository:      clientRepository,
		clientKMS:             clientKMS,
		clientEmail:           clientEmail,
		oneTimeSecretDuration: 10 * time.Minute,
	}
}

type client struct {
	clientRepository      RepositoryCIAM
	clientKMS             KMSClient
	clientEmail           SMTPClient
	oneTimeSecretDuration time.Duration
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

	userID, err = c.clientRepository.LookupUserByFingerprint(ctx, fingerprint)
	if err != nil {
		return Tokens{}, err
	}

	if userID == "" {
		userID = utils.NewUUID()
		if err := c.clientRepository.CreateUser(ctx, userID, "", fingerprint); err != nil {
			return Tokens{}, err
		}
	}

	iat := time.Now().UTC()

	return Tokens{
		Refresh: NewRefreshToken(userID, WithCustomIat(iat)),
		Access:  NewAccessToken(userID, false, WithCustomIat(iat)),
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

	userID, err = c.clientRepository.LookupUserByEmail(ctx, email)
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
		iat, err := c.clientRepository.LookupOneTimeSecret(ctx, userID)
		if err != nil {
			return Tokens{}, err
		}
		if iat.Add(time.Duration(expirationDurationIdentitySec)).After(time.Now().UTC()) {
			return Tokens{
				ID: NewIDToken(userID, email, fingerprint, WithCustomIat(iat)),
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
	if err := c.clientRepository.WriteOneTimeSecret(ctx, userID, secret, iat); err != nil {
		return Tokens{}, err
	}

	return Tokens{
		ID: NewIDToken(userID, email, fingerprint, WithCustomIat(iat)),
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

func (c client) ValidateToken(ctx context.Context, accessToken string) error {
	//TODO implement me
	panic("implement me")
}

func (c client) RefreshAccessToken(ctx context.Context, token string) (JWT, error) {
	//TODO implement me
	panic("implement me")
}
