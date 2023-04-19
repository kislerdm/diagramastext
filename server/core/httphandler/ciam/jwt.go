package ciam

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// JWT defines the JWT interface.
// TODO(?): include prompt length, RPM and RPD in scopes
type JWT interface {
	String() (string, error)
	Validate(fn SignatureVerificationFn) error
	UserID() string
	UserEmail() string
	UserDeviceFingerprint() string
	UserRole() Role
}

type OptFn func(JWT) error
type SigningFn func(signingString string) (signature string, alg string, err error)
type SignatureVerificationFn func(signingString, signature string) error

func WithCustomIat(iat time.Time) OptFn {
	return func(jwt JWT) error {
		jwt.(*token).payload.Iat = iat.Unix()
		return nil
	}
}

func WithSignature(signFn SigningFn) OptFn {
	return func(jwt JWT) (err error) {
		if signFn == nil {
			return errors.New("signing function must be provided")
		}
		signingString, err := jwt.(*token).signingString()
		if err != nil {
			return
		}
		jwt.(*token).signature, jwt.(*token).header.Alg, err = signFn(signingString)
		return
	}
}

func NewIDToken(userID, email, fingerprint string, emailVerified bool, durationSec int64, optFns ...OptFn) (
	JWT, error,
) {
	o, err := defaultToken(userID, optFns...)
	if err != nil {
		return nil, err
	}
	o.payload.Email = &email
	o.payload.Fingerprint = &fingerprint
	o.payload.EmailVerified = emailVerified
	if durationSec == 0 {
		durationSec = defaultExpirationDurationIdentitySec
	}
	o.payload.Exp = o.payload.Iat + durationSec
	return o, nil
}

func NewRefreshToken(userID string, optFns ...OptFn) (JWT, error) {
	o, err := defaultToken(userID, optFns...)
	if err != nil {
		return nil, err
	}
	o.payload.Exp = o.payload.Iat + defaultExpirationDurationRefreshSec
	return o, nil
}

func NewAccessToken(userID string, emailVerified bool, optFns ...OptFn) (JWT, error) {
	o, err := defaultToken(userID, optFns...)
	if err != nil {
		return nil, err
	}
	tmp := roleAnonymUser
	if emailVerified {
		tmp = roleRegisteredUser
	}
	o.payload.Role = &tmp
	o.payload.Exp = o.payload.Iat + defaultExpirationDurationAccessSec
	return o, nil
}

func defaultToken(userID string, optFns ...OptFn) (*token, error) {
	o := &token{
		header: JWTHeader{
			Alg: algNone,
			Typ: typ,
		},
		payload: JWTPayload{
			Iat: time.Now().UTC().Unix(),
			Iss: iss,
			Aud: aud,
			Sub: userID,
		},
	}
	for _, optFn := range optFns {
		if err := optFn(o); err != nil {
			return nil, err
		}
	}
	return o, nil
}

const (
	typ     = "JWT"
	algNone = "none"
	iss     = "https://ciam.diagramastext.dev"
	aud     = "https://diagramastext.dev"
)

var (
	// OKTA defaults: https://support.okta.com/help/s/article/What-is-the-lifetime-of-the-JWT-tokens
	defaultExpirationDurationIdentitySec = int64(time.Hour.Seconds())
	defaultExpirationDurationAccessSec   = int64(time.Hour.Seconds())
	defaultExpirationDurationRefreshSec  = 100 * 24 * int64(time.Hour.Seconds())
)

type JWTHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type JWTPayload struct {
	EmailVerified bool    `json:"email_verified,omitempty"`
	Email         *string `json:"email,omitempty"`
	Fingerprint   *string `json:"fingerprint,omitempty"`
	Role          *Role   `json:"role,omitempty"`
	Sub           string  `json:"sub"`
	Iss           string  `json:"iss"`
	Aud           string  `json:"aud"`
	Iat           int64   `json:"iat"`
	Exp           int64   `json:"exp"`
}

type token struct {
	header    JWTHeader
	payload   JWTPayload
	signature string
}

func (t token) UserEmail() string {
	if t.payload.Email == nil {
		return ""
	}
	return *t.payload.Email
}

func (t token) UserDeviceFingerprint() string {
	if t.payload.Fingerprint == nil {
		return ""
	}
	return *t.payload.Fingerprint
}

func (t token) UserID() string {
	return t.payload.Sub
}

func (t token) UserRole() Role {
	return *t.payload.Role
}

func (t token) String() (string, error) {
	signingString, err := t.signingString()
	if err != nil {
		return "", err
	}

	switch t.signature == "" {
	case true:
		if t.header.Alg != algNone {
			return "", errors.New("signature is missing")
		}
		return signingString, nil
	default:
		if t.header.Alg == algNone || t.header.Alg == "" {
			return "", errors.New("JWT header corrupt: alg value")
		}
		return signingString + "." + t.signature, nil
	}
}

func (t token) Validate(fn SignatureVerificationFn) error {
	if err := t.verifySignature(fn); err != nil {
		return err
	}
	if t.isExpired() {
		return errors.New("token is expired")
	}
	return nil
}

func (t token) signingString() (string, error) {
	header, err := json.Marshal(t.header)
	if err != nil {
		return "", err
	}
	payload, err := json.Marshal(t.payload)
	if err != nil {
		return "", err
	}
	return encodeSegment(header) + "." + encodeSegment(payload), nil
}

func (t token) verifySignature(verificationFn SignatureVerificationFn) error {
	if (t.header.Alg != algNone && verificationFn == nil) || (t.header.Alg == algNone && t.signature != "") {
		return errors.New("corrupt JWT: alg does not match the signature")
	}
	signingString, err := t.signingString()
	if err != nil {
		return err
	}
	if err := verificationFn(signingString, t.signature); err != nil {
		return err
	}
	return nil
}

func (t token) isExpired() bool {
	return t.payload.Exp <= t.payload.Iat || t.payload.Exp < time.Now().UTC().Unix()
}

func ParseToken(s string) (JWT, error) {
	elements := strings.SplitN(s, ".", 3)
	if len(elements) < 2 {
		return nil, errors.New("wrong JWT format")
	}

	var o token
	headerBytes, err := decodeSegment(elements[0])
	if err != nil {
		return nil, errors.New("wrong JWT header encoding")
	}
	if err := json.Unmarshal(headerBytes, &o.header); err != nil {
		return nil, errors.New("wrong JWT header format")
	}

	payloadBytes, err := decodeSegment(elements[1])
	if err != nil {
		return nil, errors.New("wrong JWT payload encoding")
	}
	if err := json.Unmarshal(payloadBytes, &o.payload); err != nil {
		return nil, errors.New("wrong JWT payload format")
	}

	if len(elements) == 3 {
		o.signature = elements[2]
	}

	return &o, nil
}

func encodeSegment(seg []byte) string {
	return base64.RawURLEncoding.EncodeToString(seg)
}

func decodeSegment(seg string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(seg)
}
