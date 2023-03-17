package gcpsecretsmanager

import (
	"context"
	"encoding/json"
	"errors"
	"hash/crc32"

	secretsmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/googleapis/gax-go/v2"
)

// NewSecretmanager initialises the GCP secretsmanager Client.
func NewSecretmanager(ctx context.Context) (
	*Client, error,
) {
	c, err := secretsmanager.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return &Client{c}, nil
}

// Client gcpsecretsmanager client.
type Client struct {
	c gcpSecretsmanagerClient
}

func (c Client) ReadLastVersion(ctx context.Context, uri string, output interface{}) error {
	res, err := c.c.AccessSecretVersion(
		ctx, &secretmanagerpb.AccessSecretVersionRequest{
			Name: uri,
		},
	)
	if err != nil {
		return err
	}

	if err := isValidResponse(res); err != nil {
		return err
	}

	return json.Unmarshal(res.Payload.Data, output)
}

func isValidResponse(res *secretmanagerpb.AccessSecretVersionResponse) error {
	crc32c := crc32.MakeTable(crc32.Castagnoli)
	checksum := int64(crc32.Checksum(res.GetPayload().GetData(), crc32c))
	if checksum != res.GetPayload().GetDataCrc32C() {
		return errors.New("data corruption detected")
	}
	return nil
}

type gcpSecretsmanagerClient interface {
	AccessSecretVersion(
		ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption,
	) (*secretmanagerpb.AccessSecretVersionResponse, error)
}

type mockGCPSecretsmanagerClient struct {
	v   *secretmanagerpb.AccessSecretVersionResponse
	err error
}

func (m mockGCPSecretsmanagerClient) AccessSecretVersion(
	_ context.Context, _ *secretmanagerpb.AccessSecretVersionRequest, _ ...gax.CallOption,
) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.v, nil
}
