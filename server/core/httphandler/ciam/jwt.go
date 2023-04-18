package ciam

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type Tokens struct {
	ID      JWT
	Refresh JWT
	Access  JWT
}

func (t Tokens) Serialize() ([]byte, error) {
	var (
		temp struct {
			ID      *string `json:"id,omitempty"`
			Refresh *string `json:"refresh,omitempty"`
			Access  *string `json:"access,omitempty"`
		}
		s   string
		err error
	)
	s, err = t.ID.String()
	if err != nil {
		return nil, err
	}
	temp.ID = &s

	s, err = t.Refresh.String()
	if err != nil {
		return nil, err
	}
	temp.Refresh = &s

	s, err = t.Access.String()
	if err != nil {
		return nil, err
	}
	temp.Access = &s

	return json.Marshal(temp)
}

type JWTHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type JWTPayload struct {
	EmailVerified bool    `json:"email_verified,omitempty"`
	Email         *string `json:"email,omitempty"`
	Fingerprint   *string `json:"fingerprint,omitempty"`
	Role          Role    `json:"role"`
	Iss           string  `json:"iss"`
	Sub           string  `json:"sub"`
	Aud           string  `json:"aud"`
	Iat           int64   `json:"iat"`
	Exp           int64   `json:"exp"`
}

type Role uint8

func (r Role) IsRegisteredUser() bool {
	return r == roleRegisteredUser
}

const (
	roleAnonymUser Role = iota
	roleRegisteredUser
)

type OptFn func(JWT) error
type SigningFn func(signingString string) (signature string, alg string, err error)
type SignatureVerificationFn func(signingString, signature string) error

func WithCustomIat(iat time.Time) OptFn {
	return func(jwt JWT) error {
		jwt.(*token).Payload.Iat = iat.Unix()
		return nil
	}
}

func WithSignature(signFn SigningFn) OptFn {
	return func(jwt JWT) (err error) {
		signingString, err := jwt.(*token).signingString()
		if err != nil {
			return
		}
		jwt.(*token).Signature, jwt.(*token).Header.Alg, err = signFn(signingString)
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
	o.Payload.Email = &email
	o.Payload.Fingerprint = &fingerprint
	o.Payload.EmailVerified = emailVerified
	if durationSec == 0 {
		durationSec = defaultExpirationDurationRefreshSec
	}
	o.Payload.Exp = o.Payload.Iat + durationSec
	return o, nil
}

func NewRefreshToken(userID string, optFns ...OptFn) (JWT, error) {
	o, err := defaultToken(userID, optFns...)
	if err != nil {
		return nil, err
	}
	o.Payload.Exp = o.Payload.Iat + defaultExpirationDurationRefreshSec
	return o, nil
}

func NewAccessToken(userID string, emailVerified bool, optFns ...OptFn) (JWT, error) {
	o, err := defaultToken(userID, optFns...)
	if err != nil {
		return nil, err
	}
	o.Payload.Role = roleAnonymUser
	if emailVerified {
		o.Payload.Role = roleRegisteredUser
	}
	o.Payload.Exp = o.Payload.Iat + defaultExpirationDurationAccessSec
	return o, nil
}

const (
	typ     = "JWT"
	algNone = "none"
	iss     = "https://ciam.diagramastext.dev"
	aud     = "https://diagramastext.dev"
)

func defaultToken(userID string, optFns ...OptFn) (*token, error) {
	o := &token{
		Header: JWTHeader{
			Alg: algNone,
			Typ: typ,
		},
		Payload: JWTPayload{
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

type JWT interface {
	String() (string, error)
	Validate(fn SignatureVerificationFn) error
	Role() Role
	Sub() string
	Email() string
	Fingerprint() string
}

type token struct {
	Header    JWTHeader
	Payload   JWTPayload
	Signature string
}

func (t token) Email() string {
	if t.Payload.Email == nil {
		return ""
	}
	return *t.Payload.Email
}

func (t token) Fingerprint() string {
	if t.Payload.Fingerprint == nil {
		return ""
	}
	return *t.Payload.Fingerprint
}

func (t token) Sub() string {
	return t.Payload.Sub
}

func (t token) Role() Role {
	return t.Payload.Role
}

func (t token) String() (string, error) {
	signingString, err := t.signingString()
	if err != nil {
		return "", err
	}

	if t.Header.Alg == algNone {
		return signingString, nil
	}

	if t.Signature == "" {
		return "", errors.New("signature is missing")
	}
	return signingString + "." + t.Signature, nil
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
	header, err := json.Marshal(t.Header)
	if err != nil {
		return "", err
	}
	payload, err := json.Marshal(t.Payload)
	if err != nil {
		return "", err
	}
	return encodeSegment(header) + "." + encodeSegment(payload), nil
}

func (t token) verifySignature(verificationFn SignatureVerificationFn) error {
	if (t.Header.Alg != algNone && verificationFn == nil) || (t.Header.Alg == algNone && t.Signature == "") {
		return errors.New("corrupt JWT: alg does not match the signature")
	}
	signingString, err := t.signingString()
	if err != nil {
		return err
	}
	if err := verificationFn(signingString, t.Signature); err != nil {
		return err
	}
	return nil
}

func (t token) isExpired() bool {
	return t.Payload.Exp <= t.Payload.Iat || t.Payload.Exp < time.Now().UTC().Unix()
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
	if err := json.Unmarshal(headerBytes, &o.Header); err != nil {
		return nil, errors.New("wrong JWT header format")
	}

	payloadBytes, err := decodeSegment(elements[1])
	if err != nil {
		return nil, errors.New("wrong JWT payload encoding")
	}
	if err := json.Unmarshal(payloadBytes, &o.Payload); err != nil {
		return nil, errors.New("wrong JWT payload format")
	}

	if len(elements) == 3 {
		o.Signature = elements[2]
	}

	return &o, nil
}

func encodeSegment(seg []byte) string {
	return base64.RawURLEncoding.EncodeToString(seg)
}

func decodeSegment(seg string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(seg)
}

var (
	// OKTA defaults: https://support.okta.com/help/s/article/What-is-the-lifetime-of-the-JWT-tokens
	defaultExpirationDurationIdentitySec = 1 * int64(time.Hour)
	defaultExpirationDurationAccessSec   = 1 * int64(time.Hour)
	defaultExpirationDurationRefreshSec  = 100 * 24 * int64(time.Hour)
)
