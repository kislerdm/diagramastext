package secretsmanager

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// Client secrets vault client.
type Client interface {
	// ReadLatestSecret reads and deserializes the latest version of JSON-encoded secret.
	ReadLatestSecret(ctx context.Context, uri string, output interface{}) error
}

type AWSSecretManager struct {
	awsSecretsManager
}

type awsSecretsManager interface {
	GetSecretValue(
		ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options),
	) (
		*secretsmanager.GetSecretValueOutput, error,
	)
}

func (c AWSSecretManager) ReadLatestSecret(ctx context.Context, uri string, output interface{}) error {
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
func NewAWSSecretManagerFromConfig(config aws.Config) Client {
	return AWSSecretManager{
		secretsmanager.NewFromConfig(config),
	}
}
