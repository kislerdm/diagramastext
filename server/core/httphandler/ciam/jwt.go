package ciam

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/kislerdm/diagramastext/server/core/internal/utils"
)

type Tokens struct {
	ID      JWT `json:"id,omitempty"`
	Refresh JWT `json:"refresh,omitempty"`
	Access  JWT `json:"access,omitempty"`
}

func (t Tokens) Serialize() ([]byte, error) {
	return json.Marshal(t)
}

type JWTHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type JWTPayload struct {
	Iat           int64  `json:"iat"`
	Exp           int64  `json:"exp"`
	Iss           string `json:"iss"`
	Sub           string `json:"sub"`
	Aud           string `json:"aud"`
	IsPremium     bool   `json:"isPremium,omitempty"`
	Email         string `json:"email,omitempty"`
	Fingerprint   string `json:"fingerprint,omitempty"`
	EmailVerified bool   `json:"email_verified,omitempty"`
}

type OptFn func(JWT)

func WithCustomIat(iat time.Time) OptFn {
	return func(jwt JWT) {
		jwt.(*token).Payload.Iat = iat.Unix()
	}
}

func NewIDToken(userID, email, fingerprint string, optFns ...OptFn) (JWT, error) {
	o := defaultToken(userID)
	o.Payload.Email = email
	o.Payload.Fingerprint = fingerprint

	for _, optFn := range optFns {
		optFn(o)
	}

	o.Payload.Exp = o.Payload.Iat + expirationDurationIdentitySec

	return o, nil
}

func NewRefreshToken(userID string, optFns ...OptFn) (JWT, error) {
	o := defaultToken(userID)

	for _, optFn := range optFns {
		optFn(o)
	}

	o.Payload.Exp = o.Payload.Iat + expirationDurationRefreshSec

	return o, nil
}

func NewAccessToken(userID string, isPremium bool, optFns ...OptFn) (JWT, error) {
	o := defaultToken(userID)
	o.Payload.IsPremium = isPremium

	for _, optFn := range optFns {
		optFn(o)
	}

	o.Payload.Exp = o.Payload.Iat + expirationDurationAccessSec

	return o, nil
}

const (
	typ     = "JWT"
	algNone = "none"
	iss     = "https://ciam.diagramastext.dev"
	aud     = "https://diagramastext.dev"
)

func defaultToken(userID string) *token {
	return &token{
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
}

type JWT interface {
	Validate() error
	SigningString() (string, error)
	String() (string, error)
	SetAlg(alg string)
	GetAlg() string
}

type token struct {
	Header    JWTHeader
	Payload   JWTPayload
	Signature string
}

func (t *token) SetAlg(alg string) {
	t.Header.Alg = alg
}

func (t *token) GetAlg() string {
	return t.Header.Alg
}

func (t *token) String() (string, error) {
	signingString, err := t.SigningString()
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

func (t *token) SigningString() (string, error) {
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

func (t *token) Validate() error {
	if t.Header.Typ != typ {
		return errors.New("corrupt JWT header: typ")
	}

	if t.Payload.Iss != iss {
		return errors.New("corrupt JWT payload: iss")
	}
	if t.Payload.Aud != aud {
		return errors.New("corrupt JWT payload: aud")
	}
	if err := utils.ValidateUUID(t.Payload.Sub); err != nil {
		return errors.New("corrupt JWT payload: sub")
	}
	if t.Payload.Exp < t.Payload.Iat {
		return errors.New("corrupt JWT payload: exp, iat")
	}

	now := time.Now().UTC().Unix()
	if t.Payload.Exp == t.Payload.Iat || t.Payload.Exp < now {
		return errors.New("JWT has expired")
	}

	return nil
}

func ParseToken(v string) (JWT, error) {
	elements := strings.SplitN(v, ".", 3)
	if len(elements) != 3 {
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

	o.Signature = elements[2]

	if err := o.Validate(); err != nil {
		return nil, err
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
	expirationDurationIdentitySec = 10 * int64(time.Minute)
	expirationDurationRefreshSec  = 7 * 24 * int64(time.Hour)
	expirationDurationAccessSec   = 30 * int64(time.Minute)
)
