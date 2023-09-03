package ciam

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"math/rand"
)

func GenerateCertificate() ed25519.PrivateKey {
	_, o, _ := ed25519.GenerateKey(rand.New(rand.NewSource(0)))
	return o
}

func ReadPrivateKey(s string) (ed25519.PrivateKey, error) {
	block, _ := pem.Decode([]byte(s))
	if block == nil || block.Type != "PRIVATE KEY" {
		return nil, errors.New("cannot read the key")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	o, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, errors.New("wrong key type")
	}

	return o, nil
}

func MarshalKey(key ed25519.PrivateKey) ([]byte, error) {
	v, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, err
	}

	return pem.EncodeToMemory(
		&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: v,
		},
	), nil
}
