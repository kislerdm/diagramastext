package ciam

import (
	"context"
	"errors"
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

type User struct {
	ID, Email, Fingerprint  string
	IsActive, EmailVerified bool
}

type Secret struct {
	Secret   string
	IssuedAt time.Time
}

type MockRepositoryCIAM struct {
	UserID          map[string]*User
	UserEmail       map[string]*User
	UserFingerprint map[string]*User
	Secret          map[string]Secret
	Err             error
}

func (m *MockRepositoryCIAM) setUser(u *User) {
	if m.UserEmail == nil {
		m.UserEmail = map[string]*User{}
	}
	m.UserEmail[u.Email] = u

	if m.UserFingerprint == nil {
		m.UserFingerprint = map[string]*User{}
	}
	m.UserFingerprint[u.Fingerprint] = u

	if m.UserID == nil {
		m.UserID = map[string]*User{}
	}
	m.UserID[u.ID] = u
}

func (m *MockRepositoryCIAM) CreateUser(_ context.Context, id, email, fingerprint string) error {
	if m.Err != nil {
		return m.Err
	}
	m.setUser(
		&User{
			ID:            id,
			Email:         email,
			Fingerprint:   fingerprint,
			IsActive:      false,
			EmailVerified: false,
		},
	)
	return nil
}

func (m *MockRepositoryCIAM) ReadUser(_ context.Context, id string) (
	found, isActive, emailVerified bool, email, fingerprint string, err error,
) {
	if m.Err != nil {
		return false, false, false, "", "", err
	}
	if u, ok := m.UserID[id]; ok {
		return true, u.IsActive, u.EmailVerified, u.Email, u.Fingerprint, nil
	}
	return false, false, false, "", "", nil
}

func (m *MockRepositoryCIAM) LookupUserByEmail(_ context.Context, email string) (id string, isActive bool, err error) {
	if m.Err != nil {
		return "", false, err
	}
	if u, ok := m.UserEmail[email]; ok {
		return u.ID, u.IsActive, nil
	}
	return "", false, nil
}

func (m *MockRepositoryCIAM) LookupUserByFingerprint(_ context.Context, fingerprint string) (
	id string, isActive bool, err error,
) {
	if m.Err != nil {
		return "", false, err
	}
	if u, ok := m.UserFingerprint[fingerprint]; ok {
		return u.ID, u.IsActive, nil
	}
	return "", false, nil
}

func (m *MockRepositoryCIAM) UpdateUserSetActiveStatus(ctx context.Context, id string, isActive bool) error {
	if m.Err != nil {
		return m.Err
	}
	found, _, emailVerified, email, fingerprint, _ := m.ReadUser(ctx, id)
	if !found {
		return errors.New("user not found")
	}
	m.setUser(
		&User{
			ID:            id,
			Email:         email,
			Fingerprint:   fingerprint,
			IsActive:      isActive,
			EmailVerified: emailVerified,
		},
	)
	return nil
}

func (m *MockRepositoryCIAM) UpdateUserSetEmailVerified(ctx context.Context, id string) error {
	if m.Err != nil {
		return m.Err
	}
	found, isActive, _, email, fingerprint, _ := m.ReadUser(ctx, id)
	if !found {
		return errors.New("user not found")
	}
	m.setUser(
		&User{
			ID:            id,
			Email:         email,
			Fingerprint:   fingerprint,
			IsActive:      isActive,
			EmailVerified: true,
		},
	)
	return nil
}

func (m *MockRepositoryCIAM) CreateOneTimeSecret(_ context.Context, userID, secret string, createdAt time.Time) error {
	if m.Err != nil {
		return m.Err
	}
	if m.Secret == nil {
		m.Secret = map[string]Secret{}
	}
	m.Secret[userID] = Secret{
		Secret:   secret,
		IssuedAt: createdAt,
	}
	return nil
}

func (m *MockRepositoryCIAM) ReadOneTimeSecret(_ context.Context, userID string) (
	found bool, secret string, issuedAt time.Time, err error,
) {
	if m.Err != nil {
		return false, "", time.Time{}, err
	}
	if v, ok := m.Secret[userID]; ok {
		return ok, v.Secret, v.IssuedAt, nil
	}
	return false, "", time.Time{}, nil
}

func (m *MockRepositoryCIAM) DeleteOneTimeSecret(_ context.Context, userID string) error {
	if m.Err != nil {
		return m.Err
	}
	delete(m.Secret, userID)
	return nil
}
