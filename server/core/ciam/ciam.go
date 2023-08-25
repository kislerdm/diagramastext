// Package ciam to authN/Z users
package ciam

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/kislerdm/diagramastext/server/core/diagram"
	errs "github.com/kislerdm/diagramastext/server/core/errors"
	"github.com/kislerdm/diagramastext/server/core/internal/utils"
)

type Mock struct {
	LeadPath string
	Err      error
	O        []byte
	User     diagram.User
}

func (m Mock) LeadingPath() string {
	return m.LeadPath
}

func (m Mock) HTTPHandler(_ *http.Request) ([]byte, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.O, nil
}

func (m Mock) ParseAccessToken(_ context.Context, _ string) (diagram.User, error) {
	if m.Err != nil {
		return diagram.User{}, m.Err
	}
	return m.User, nil
}

type Client interface {
	LeadingPath() string
	HTTPHandler(req *http.Request) ([]byte, error)
	ParseAccessToken(ctx context.Context, token string) (diagram.User, error)
}

// NewClient initializes the CIAM client.
func NewClient(
	clientRepository RepositoryCIAM, clientEmail SMTPClient, privateKey ed25519.PrivateKey,
) (Client, error) {
	issuer, err := NewIssuer(privateKey)
	if err != nil {
		return nil, err
	}
	return client{
		clientRepository: clientRepository,
		clientEmail:      clientEmail,
		tokenIssuer:      issuer,
	}, nil
}

type client struct {
	clientRepository RepositoryCIAM
	clientEmail      SMTPClient
	tokenIssuer      Issuer
}

func (c client) HTTPHandler(r *http.Request) ([]byte, error) {
	path := strings.TrimPrefix(r.URL.Path, c.LeadingPath())
	switch path {
	// anonym's authentication flow:
	// Fingerprint found in DB -> No  -> Create \
	//							-> Yes ->  --	-> Generate id, refresh and access JWT -> Return generated tokens.
	case "/anonym":
		defer func() { _ = r.Body.Close() }()
		var req struct {
			Fingerprint string `json:"fingerprint"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return []byte(`{"error":"request parsing error"}`), errs.NewInputFormatValidationError(err)
		}
		if f, _ := regexp.MatchString(`^[a-f0-9]{40}$`, req.Fingerprint); !f {
			return []byte(`{"error":"invalid request"}`),
				errs.NewInputContentValidationError(errors.New("invalid fingerprint"))
		}

		userID, isActive, err := c.clientRepository.LookupUserByFingerprint(r.Context(), req.Fingerprint)
		if err != nil {
			return nil, err
		}

		if userID != "" && !isActive {
			return nil, errors.New("user was deactivated")
		}

		if userID == "" {
			userID = utils.NewUUID()
			role := uint8(diagram.RoleAnonymUser)
			if err := c.clientRepository.CreateUser(
				r.Context(), userID, "", req.Fingerprint, true, &role,
			); err != nil {
				return nil, err
			}
		}

		return c.issueTokens(r.Context(), diagram.User{ID: userID, Role: diagram.RoleAnonymUser}, "", req.Fingerprint)
	// 	TODO: case "/init", "/confirm", "/refresh", "/resend"
	default:
		return []byte(`{"error":"CIAM resource ` + r.URL.Path + ` not found"}`),
			errs.NewHandlerNotExistsError(errors.New("CIAM: " + r.URL.Path + " not found"))
	}
}

func (c client) LeadingPath() string {
	return "/auth"
}

// SigninUser executes user's authentication flow:
//
//	Email found in DB -> No  -> Create \
//			 	   	  -> Yes ->	--	  -> Generate secret and id JWT -> Send secret to email -> Return id JWT.
func (c client) SigninUser(ctx context.Context, email, fingerprint string) ([]byte, error) {
	if email == "" {
		return nil, errors.New("email must be provided")
	}

	const defaultExpirationSecret = 10 * time.Minute

	var (
		userID string
		err    error
	)

	userID, isActive, err := c.clientRepository.LookupUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	switch userID == "" {
	case true:
		userID = utils.NewUUID()
		role := uint8(diagram.RoleRegisteredUser)
		if err := c.clientRepository.CreateUser(
			ctx, userID, email, fingerprint, false, &role,
		); err != nil {
			return nil, err
		}
	default:
		if !isActive {
			return nil, errors.New("user was deactivated")
		}
		found, _, iat, err := c.clientRepository.ReadOneTimeSecret(ctx, userID)
		if err != nil {
			return nil, err
		}
		if found && iat.Add(defaultExpirationSecret).After(time.Now().UTC()) {
			tkn, err := c.tokenIssuer.NewIDToken(
				userID, email, fingerprint, WithValidityDuration(defaultExpirationSecret),
			)
			if err != nil {
				return nil, err
			}
			return []byte(tkn), nil
		}
	}

	secret := generateOnetimeSecret()
	iat := time.Now().UTC()

	if err := c.clientRepository.WriteOneTimeSecret(ctx, userID, secret, iat); err != nil {
		return nil, err
	}
	if err := c.clientEmail.SendSignInEmail(email, secret); err != nil {
		return nil, err
	}
	tkn, err := c.tokenIssuer.NewIDToken(
		userID, email, fingerprint, WithCustomIat(iat), WithValidityDuration(defaultExpirationSecret),
	)
	if err != nil {
		return nil, err
	}
	return []byte(tkn), nil
}

func (c client) IssueTokensAfterSecretConfirmation(ctx context.Context, identityToken, secret string) ([]byte, error) {
	userID, email, fingerprint, err := c.tokenIssuer.ParseIDToken(identityToken)
	if err != nil {
		return nil, err
	}

	found, secretRef, _, err := c.clientRepository.ReadOneTimeSecret(ctx, userID)
	if err != nil {
		return nil, err
	}

	if !found {
		return nil, errors.New("no secret was sent")
	}

	if secret != secretRef {
		return nil, errors.New("secret is wrong")
	}

	if err := c.clientRepository.UpdateUserSetActive(ctx, userID); err != nil {
		return nil, err
	}

	_ = c.clientRepository.DeleteOneTimeSecret(ctx, userID)

	return c.issueTokens(ctx, diagram.User{ID: userID, Role: diagram.RoleRegisteredUser}, email, fingerprint)
}

func (c client) issueTokens(_ context.Context, user diagram.User, email, fingerprint string) (
	[]byte, error,
) {
	iat := time.Now().UTC()

	idToken, err := c.tokenIssuer.NewIDToken(user.ID, email, fingerprint, WithCustomIat(iat))
	if err != nil {
		return nil, err
	}

	accessToken, err := c.tokenIssuer.NewAccessToken(user, WithCustomIat(iat))
	if err != nil {
		return nil, err
	}

	refreshToken, err := c.tokenIssuer.NewRefreshToken(user.ID, WithCustomIat(iat))
	if err != nil {
		return nil, err
	}

	return json.Marshal(
		struct {
			ID      *string `json:"id,omitempty"`
			Refresh *string `json:"refresh,omitempty"`
			Access  *string `json:"access,omitempty"`
		}{
			ID:      pointerStr(idToken),
			Refresh: pointerStr(refreshToken),
			Access:  pointerStr(accessToken),
		},
	)
}

func (c client) ParseAccessToken(_ context.Context, token string) (diagram.User, error) {
	return c.tokenIssuer.ParseAccessToken(token)
}

func (c client) RefreshTokens(ctx context.Context, refreshToken string) ([]byte, error) {
	userID, err := c.tokenIssuer.ParseRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}
	found, isActive, roleID, email, fingerprint, err := c.clientRepository.ReadUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, errors.New("user not found")
	}
	if !isActive {
		return nil, errors.New("user was deactivated")
	}
	role := diagram.Role(roleID)
	if email != "" && role == diagram.RoleAnonymUser {
		return nil, errors.New("user's email was not verified yet")
	}
	return c.issueTokens(ctx, diagram.User{ID: userID, Role: role}, email, fingerprint)
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

func GenerateCertificate() ed25519.PrivateKey {
	_, o, _ := ed25519.GenerateKey(rand.New(rand.NewSource(0)))
	return o
}
