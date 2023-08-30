package ciam

import (
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"time"
)

const (
	iss = "https://ciam.diagramastext.dev"
	aud = "https://diagramastext.dev"
)

var (
	// OKTA defaults: https://support.okta.com/help/s/article/What-is-the-lifetime-of-the-JWT-tokens
	defaultExpirationDurationIdentity = time.Hour
	defaultExpirationDurationAccess   = time.Hour
	defaultExpirationDurationRefresh  = 100 * 24 * time.Hour
)

type stdClaims struct {
	Sub string `json:"sub"`
	Iss string `json:"iss"`
	Aud string `json:"aud"`
	Iat int64  `json:"iat"`
	Exp int64  `json:"exp"`
}

func (s stdClaims) IsValidToken() error {
	if s.Exp < time.Now().Unix() {
		return errors.New("token expired")
	}
	return nil
}

func newStdClaims(userID string, duration time.Duration, fnOps ...ClaimsOps) stdClaims {
	c := stdClaims{
		Sub: userID,
		Iss: iss,
		Aud: aud,
		Iat: time.Now().Unix(),
	}
	setExp(&c, duration)
	for _, fn := range fnOps {
		fn(&c)
	}
	return c
}

type ClaimsOps func(claims *stdClaims)

type idTokenClaims struct {
	stdClaims
	Email       string `json:"email,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
}

type accessTokenClaims struct {
	stdClaims
	Role   Role   `json:"role"`
	Quotas Quotas `json:"quotas"`
}

type refreshTokenClaims struct {
	stdClaims
}

func setExp(claims *stdClaims, d time.Duration) {
	claims.Exp = claims.Iat + d.Milliseconds()/1000
}

func WithCustomIat(iat time.Time) ClaimsOps {
	return func(claims *stdClaims) {
		d := claims.Exp - claims.Iat
		claims.Iat = iat.Unix()
		setExp(claims, time.Duration(d)*time.Second)
	}
}

func WithValidityDuration(d time.Duration) ClaimsOps {
	return func(claims *stdClaims) {
		setExp(claims, d)
	}
}

// Issuer defines the interface which issuer and parses JWT.
type Issuer interface {
	// NewIDToken issuer id JWT.
	NewIDToken(userID, email, fingerprint string, fnOps ...ClaimsOps) (string, error)
	// NewAccessToken issuer access JWT.
	NewAccessToken(user User, fnOps ...ClaimsOps) (string, error)
	// NewRefreshToken issuer refresh JWT.
	NewRefreshToken(userID string, fnOps ...ClaimsOps) (string, error)
	// ParseIDToken parses id JWT.
	ParseIDToken(token string) (userID, email, fingerprint string, err error)
	// ParseRefreshToken parses refresh JWT.
	ParseRefreshToken(token string) (userID string, err error)
	// ParseAccessToken parses access JWT.
	ParseAccessToken(token string) (user User, err error)
}

func NewIssuer(key ed25519.PrivateKey) (Issuer, error) {
	if key == nil {
		return nil, errors.New("no valid ed25519 private key provided")
	}
	pubKey, ok := key.Public().(ed25519.PublicKey)
	if !ok || len(pubKey) != ed25519.PublicKeySize {
		return nil, errors.New("key is invalid")
	}

	h := struct {
		Alg string `json:"alg"`
		Typ string `json:"typ"`
	}{
		Alg: "EdDSA",
		Typ: "JWT",
	}
	header, _ := json.Marshal(h)

	return issuer{
		privKey: key,
		pubKey:  pubKey,
		header:  encodeSegment(header),
	}, nil
}

type issuer struct {
	privKey ed25519.PrivateKey
	pubKey  ed25519.PublicKey
	header  string
}

func (i issuer) serializeAndSign(tkn interface{}) (string, error) {
	payload, err := json.Marshal(tkn)
	if err != nil {
		return "", err
	}

	signingStr := i.header + "." + encodeSegment(payload)

	signature, err := i.privKey.Sign(rand.Reader, []byte(signingStr), crypto.Hash(0))
	if err != nil {
		return "", err
	}

	return signingStr + "." + encodeSegment(signature), nil
}

func (i issuer) NewIDToken(userID, email, fingerprint string, fnOps ...ClaimsOps) (string, error) {
	tkn := idTokenClaims{
		Email:       email,
		Fingerprint: fingerprint,
		stdClaims:   newStdClaims(userID, defaultExpirationDurationIdentity, fnOps...),
	}
	return i.serializeAndSign(tkn)
}

func (i issuer) NewAccessToken(user User, fnOps ...ClaimsOps) (string, error) {
	tkn := accessTokenClaims{
		Role:      user.Role,
		Quotas:    user.Role.Quotas(),
		stdClaims: newStdClaims(user.ID, defaultExpirationDurationAccess, fnOps...),
	}
	return i.serializeAndSign(tkn)
}

func (i issuer) NewRefreshToken(userID string, fnOps ...ClaimsOps) (string, error) {
	tkn := refreshTokenClaims{
		stdClaims: newStdClaims(userID, defaultExpirationDurationRefresh, fnOps...),
	}
	return i.serializeAndSign(tkn)
}

func (i issuer) parseToken(token string, tkn interface{}) error {
	els := strings.Split(token, ".")
	if len(els) < 3 {
		return errors.New("wrong token format")
	}

	sig, err := decodeSegment(els[2])
	if err != nil {
		return errors.New("wrong signature format")
	}

	signingStr := els[0] + "." + els[1]

	if !ed25519.Verify(i.pubKey, []byte(signingStr), sig) {
		return errors.New("wrong signature")
	}

	payload, err := decodeSegment(els[1])
	if err != nil {
		return errors.New("wrong payload format")
	}

	if err := json.Unmarshal(payload, &tkn); err != nil {
		return errors.New("cannot deserialize payload")
	}

	return nil
}

func (i issuer) ParseIDToken(token string) (userID, email, fingerprint string, err error) {
	var tkn idTokenClaims
	if err := i.parseToken(token, &tkn); err != nil {
		return "", "", "", err
	}
	if err := tkn.IsValidToken(); err != nil {
		return "", "", "", err
	}
	return tkn.Sub, tkn.Email, tkn.Fingerprint, nil
}

func (i issuer) ParseRefreshToken(token string) (userID string, err error) {
	var tkn refreshTokenClaims
	if err := i.parseToken(token, &tkn); err != nil {
		return "", err
	}
	if err := tkn.IsValidToken(); err != nil {
		return "", err
	}
	return tkn.Sub, nil
}

func (i issuer) ParseAccessToken(token string) (user User, err error) {
	var tkn accessTokenClaims
	if err = i.parseToken(token, &tkn); err != nil {
		return
	}
	if err = tkn.IsValidToken(); err != nil {
		return
	}

	if !reflect.DeepEqual(tkn.Quotas, tkn.Role.Quotas()) {
		err = errors.New("quotas from the token are not up to date")
		return
	}

	user = User{ID: tkn.Sub, Role: tkn.Role}
	return
}

func encodeSegment(seg []byte) string {
	return base64.RawURLEncoding.EncodeToString(seg)
}

func decodeSegment(seg string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(seg)
}
