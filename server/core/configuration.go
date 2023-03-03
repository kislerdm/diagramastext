package core

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	errs "github.com/kislerdm/diagramastext/server/core/errors"
	"github.com/kislerdm/diagramastext/server/core/openai"
	"github.com/kislerdm/diagramastext/server/core/secretsmanager"
	"github.com/kislerdm/diagramastext/server/core/utils"
)

// Config `core` client configuration.
type Config struct {
	StorageConfig        storageConfig
	ModelInferenceConfig openai.ConfigOpenAI
}

type storageConfig struct {
	DBHost     string `json:"db_host"`
	DBName     string `json:"db_name"`
	DBUser     string `json:"db_user"`
	DBPassword string `json:"db_password"`
}

type secret struct {
	storageConfig
	OpenAiAPIKey string `json:"openai_api_key"`
}

// LoadDefaultConfig loads `core` configuration
func LoadDefaultConfig(ctx context.Context, awsOptFn ...func(*config.LoadOptions) error) (Config, error) {
	cfg := loadEnvVarConfig()

	secretARN := os.Getenv("ACCESS_CREDENTIALS_ARN")

	if secretARN == "" {
		return cfg, nil
	}

	awsCfg, err := config.LoadDefaultConfig(ctx, awsOptFn...)
	if err != nil {
		return Config{}, errs.Error{
			Service: errs.ServiceAWSConfig,
			Message: err.Error(),
		}
	}

	client := secretsmanager.NewAWSSecretManagerFromConfig(awsCfg)

	var s secret
	if err := client.ReadLatestSecret(ctx, secretARN, &s); err != nil {
		return Config{}, errs.Error{
			Service: errs.ServiceSecretsManager,
			Message: err.Error(),
		}
	}

	cfg.ModelInferenceConfig.Token = s.OpenAiAPIKey
	cfg.StorageConfig = s.storageConfig

	return cfg, nil
}

func loadEnvVarConfig() Config {
	return Config{
		ModelInferenceConfig: openai.ConfigOpenAI{
			Token:     os.Getenv("OPENAI_API_KEY"),
			MaxTokens: utils.MustParseInt(os.Getenv("OPENAI_MAX_TOKENS")),
		},
		StorageConfig: storageConfig{
			DBHost:     os.Getenv("DB_HOST"),
			DBName:     os.Getenv("DB_DBNAME"),
			DBUser:     os.Getenv("DB_USER"),
			DBPassword: os.Getenv("DB_PASSWORD"),
		},
	}
}
