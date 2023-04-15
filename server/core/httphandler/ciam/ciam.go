// Package ciam to authN/Z users
package ciam

import (
	"context"
	"errors"

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

// RepositoryUsers defines the communication port to persistence layer hosting users' data.
type RepositoryUsers interface {
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
}

func NewCIAMClient(clientRepository RepositoryUsers, clientKMS KMSClient, clientEmail SMTPClient) CIAMClient {
	return &client{
		clientRepository: clientRepository,
		clientKMS:        clientKMS,
		clientEmail:      clientEmail,
	}
}

type client struct {
	clientRepository RepositoryUsers
	clientKMS        KMSClient
	clientEmail      SMTPClient
}

// SigninAnonym executes anonym's authentication flow:
//
//	Fingerprint found in DB -> No  -> Create \
//							   -> Yes ->  --	-> Generate Refresh and Access JWT -> Return generates JWT.
func (c client) SigninAnonym(ctx context.Context, fingerprint string) (Tokens, error) {
	//TODO implement me
	panic("implement me")
}

// SigninUser executes user's authentication flow:
//
//	Email found in DB -> No  -> Create \
//			 	   		 -> Yes ->	--	  -> Generate secret and ID JWT -> Send secret to email -> Return ID JWT.
func (c client) SigninUser(ctx context.Context, email, fingerprint string) (Tokens, error) {
	if email == "" {
		return Tokens{}, errors.New("email must be provided")
	}

	var (
		clientID string
		err      error
	)

	clientID, err = c.clientRepository.LookupUserByEmail(ctx, email)
	if err != nil {
		return Tokens{}, err
	}

	if clientID == "" {
		clientID = utils.NewUUID()
		if err := c.clientRepository.CreateUser(ctx, clientID, email, fingerprint); err != nil {
			return Tokens{}, err
		}
	}

	//iat := time.Now().UTC()
	//secret := generateOnetimeSecret()

	return Tokens{}, nil
}

func generateOnetimeSecret() string {
	panic("todo")
}

func (c client) ValidateToken(ctx context.Context, accessToken string) error {
	//TODO implement me
	panic("implement me")
}

func (c client) RefreshAccessToken(ctx context.Context, token string) (JWT, error) {
	//TODO implement me
	panic("implement me")
}
