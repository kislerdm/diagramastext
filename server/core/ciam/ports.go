package ciam

import (
	"context"
	"errors"
	"time"
)

// RepositoryCIAM defines the communication port to persistence layer hosting users' data.
type RepositoryCIAM interface {
	CreateUser(ctx context.Context, id, email, fingerprint string, isActive bool, role *uint8) error
	ReadUser(ctx context.Context, id string) (
		found, isActive bool, role uint8, email, fingerprint string, err error,
	)

	LookupUserByEmail(ctx context.Context, email string) (id string, isActive bool, err error)
	LookupUserByFingerprint(ctx context.Context, fingerprint string) (id string, isActive bool, err error)

	// UpdateUserSetActive user active.
	UpdateUserSetActive(ctx context.Context, userID string) error

	// WriteOneTimeSecret creates a new, or updates existing one-time secret.
	WriteOneTimeSecret(ctx context.Context, userID, secret string, createdAt time.Time) error
	ReadOneTimeSecret(ctx context.Context, userID string) (found bool, secret string, issuedAt time.Time, err error)
	DeleteOneTimeSecret(ctx context.Context, userID string) error

	// GetDailySuccessfulResultsTimestampsByUserID reads the timestamps of all user's successful requests
	// which led to successful diagrams generation over the last 24 hours / day.
	GetDailySuccessfulResultsTimestampsByUserID(ctx context.Context, userID string) ([]time.Time, error)

	// GetActiveUserIDByActiveTokenID reads userID from the repository given the tokenID.
	// It returns a non-empty value if and only if the token and user are active.
	GetActiveUserIDByActiveTokenID(ctx context.Context, token string) (userID string, err error)
}

type userContainer struct {
	ID, Email, Fingerprint string
	IsActive               bool
	RoleID                 uint8
}

type Secret struct {
	Secret   string
	IssuedAt time.Time
}

type MockRepositoryCIAM struct {
	UserID          map[string]*userContainer
	UserEmail       map[string]*userContainer
	UserFingerprint map[string]*userContainer
	Secret          map[string]Secret
	Err             error
	Timestamps      []time.Time
	UserToken       map[string]string
}

func (m *MockRepositoryCIAM) CreateUser(
	_ context.Context, id, email, fingerprint string, isActive bool, role *uint8,
) error {
	if m.Err != nil {
		return m.Err
	}
	m.setUser(
		&userContainer{
			ID:          id,
			Email:       email,
			Fingerprint: fingerprint,
			IsActive:    isActive,
			RoleID:      *role,
		},
	)
	return nil
}

func (m *MockRepositoryCIAM) UpdateUserSetActive(_ context.Context, userID string) error {
	if m.Err != nil {
		return m.Err
	}
	if _, ok := m.UserID[userID]; !ok {
		return errors.New("user not found")
	}
	m.UserID[userID].IsActive = true
	return nil
}

func (m *MockRepositoryCIAM) setUser(u *userContainer) {
	if m.UserEmail == nil {
		m.UserEmail = map[string]*userContainer{}
	}
	m.UserEmail[u.Email] = u

	if m.UserFingerprint == nil {
		m.UserFingerprint = map[string]*userContainer{}
	}
	m.UserFingerprint[u.Fingerprint] = u

	if m.UserID == nil {
		m.UserID = map[string]*userContainer{}
	}
	m.UserID[u.ID] = u
}

func (m *MockRepositoryCIAM) ReadUser(_ context.Context, id string) (
	found, isActive bool, role uint8, email, fingerprint string, err error,
) {
	if m.Err != nil {
		return false, false, 0, "", "", m.Err
	}
	if u, ok := m.UserID[id]; ok {
		return true, u.IsActive, u.RoleID, u.Email, u.Fingerprint, nil
	}
	return false, false, 0, "", "", nil
}

func (m *MockRepositoryCIAM) LookupUserByEmail(_ context.Context, email string) (id string, isActive bool, err error) {
	if m.Err != nil {
		return "", false, m.Err
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
		return "", false, m.Err
	}
	if u, ok := m.UserFingerprint[fingerprint]; ok {
		return u.ID, u.IsActive, nil
	}
	return "", false, nil
}

func (m *MockRepositoryCIAM) WriteOneTimeSecret(_ context.Context, userID, secret string, createdAt time.Time) error {
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
		return false, "", time.Time{}, m.Err
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

func (m *MockRepositoryCIAM) GetDailySuccessfulResultsTimestampsByUserID(_ context.Context, _ string) (
	[]time.Time, error,
) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Timestamps, nil
}

func (m *MockRepositoryCIAM) GetActiveUserIDByActiveTokenID(_ context.Context, token string) (string, error) {
	if m.Err != nil {
		return "", m.Err
	}
	if v, ok := m.UserToken[token]; ok {
		return v, nil
	}
	return "", errors.New("token not found")
}
