package ciam

import (
	"context"
	"time"
)

// RepositoryCIAM defines the communication port to persistence layer hosting users' data.
type RepositoryCIAM interface {
	CreateUser(ctx context.Context, id, email, fingerprint string) error
	ReadUser(ctx context.Context, id string) (
		found, isActive, emailVerified bool, email, fingerprint string, err error,
	)

	LookupUserByEmail(ctx context.Context, email string) (id string, isActive bool, err error)
	LookupUserByFingerprint(ctx context.Context, fingerprint string) (id string, isActive bool, err error)

	UpdateUserSetActiveStatus(ctx context.Context, id string, isActive bool) error
	UpdateUserSetEmailVerified(ctx context.Context, id string) error

	CreateOneTimeSecret(ctx context.Context, userID, secret string, createdAt time.Time) error
	ReadOneTimeSecret(ctx context.Context, userID string) (found bool, secret string, issuedAt time.Time, err error)
	DeleteOneTimeSecret(ctx context.Context, userID string) error
}
