package ciam

import (
	"context"
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"errors"
)

// NewTokenSigningClientEd25519 initialises the client to sign JWT using ed25519 keys.
func NewTokenSigningClientEd25519(privateKey, publicKey []byte) (TokenSigningClient, error) {
	priv := ed25519.PrivateKey(privateKey)
	pub := ed25519.PublicKey(publicKey)
	if !pub.Equal(priv.Public()) {
		return nil, errors.New("faulty private/public keys pair provided")
	}
	return clientEd25519{
		alg:        "EdDSA",
		privateKey: priv,
		publicKey:  pub,
	}, nil
}

type clientEd25519 struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
	alg        string
}

type hasherEd25519 struct{}

func (h hasherEd25519) HashFunc() crypto.Hash {
	return 0
}

func (e clientEd25519) Sign(_ context.Context, signingString string) (signature []byte, alg string, err error) {
	signatureBytes, err := e.privateKey.Sign(rand.Reader, []byte(signingString), hasherEd25519{})

	// The error is always nil because the signingString is ASSUMED to be not hashed - hasherEd25519 definition
	// FIXME: would validation of hashing be required?
	if err != nil {
		return nil, "EdDSA", err
	}

	return signatureBytes, e.alg, nil
}

func (e clientEd25519) Verify(_ context.Context, signingString string, signature []byte) error {
	if !ed25519.Verify(e.publicKey, []byte(signingString), signature) {
		return errors.New("failed signature verification")
	}
	return nil
}
