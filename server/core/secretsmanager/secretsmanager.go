package secretsmanager

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/kislerdm/diagramastext/server/core/contract"
)

type client struct {
	awsSecretsManager
}

type awsSecretsManager interface {
	GetSecretValue(
		ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options),
	) (
		*secretsmanager.GetSecretValueOutput, error,
	)
}

func (c client) ReadLatestSecret(ctx context.Context, uri string, output interface{}) error {
	v, err := c.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{SecretId: &uri})
	if err != nil {
		return err
	}

	if v.SecretString == nil {
		return errors.New("no secret found")
	}

	return json.Unmarshal([]byte(aws.ToString(v.SecretString)), output)
}

// NewAWSSecretManagerFromConfig initiates the AWS Secretsmanager.
func NewAWSSecretManagerFromConfig(config aws.Config) contract.ClientSecretsmanager {
	return client{
		secretsmanager.NewFromConfig(config),
	}
}

// MockAWSSecretsmanagerClient mock of the AWS Secretsmanager's client.
type MockAWSSecretsmanagerClient struct {
	Output *secretsmanager.GetSecretValueOutput
}

func (m MockAWSSecretsmanagerClient) GetSecretValue(
	_ context.Context, _ *secretsmanager.GetSecretValueInput, _ ...func(*secretsmanager.Options),
) (*secretsmanager.GetSecretValueOutput, error) {
	if m.Output == nil {
		return nil, errors.New("no secret found")
	}
	return m.Output, nil
}
