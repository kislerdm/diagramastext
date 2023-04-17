package ciam

import "context"

type KMSClient interface {
	Verify(ctx context.Context, signingString, signature string) error
	Sign(ctx context.Context, signingString string) (signature string, alg string, err error)
}
