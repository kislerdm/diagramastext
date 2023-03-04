package configuration

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/kislerdm/diagramastext/server/core/openai"
	"github.com/kislerdm/diagramastext/server/core/secretsmanager"
	"github.com/kislerdm/diagramastext/server/core/storage"
	"github.com/kislerdm/diagramastext/server/core/utils"
)

// Config `core` client configuration.
type Config struct {
	StorageConfig        storage.PgConfig
	ModelInferenceConfig openai.ConfigOpenAI
}

type secret struct {
	storage.PgConfig
	OpenAiAPIKey string `json:"openai_api_key"`
}

const (
	tableWritePrompt          = "user_prompt"
	tableWriteModelPrediction = "openai_response"
)

// LoadDefaultConfig loads `core` configuration
func LoadDefaultConfig(ctx context.Context, awsOptFn ...func(*config.LoadOptions) error) (Config, error) {
	// defaults
	cfg := Config{
		StorageConfig: storage.PgConfig{
			TablePrompt:     tableWritePrompt,
			TablePrediction: tableWriteModelPrediction,
		},
	}

	loadEnvVarConfig(&cfg)

	if secretARN := os.Getenv("ACCESS_CREDENTIALS_ARN"); secretARN != "" {
		loadFromSecretsManager(ctx, &cfg, secretARN, awsOptFn...)
	}

	return cfg, nil

}

func loadFromSecretsManager(
	ctx context.Context, cfg *Config, secretARN string, awsOptFn ...func(*config.LoadOptions) error,
) {
	awsCfg, _ := config.LoadDefaultConfig(ctx, awsOptFn...)

	client := secretsmanager.NewAWSSecretManagerFromConfig(awsCfg)

	var s secret
	if err := client.ReadLatestSecret(ctx, secretARN, &s); err == nil {
		cfg.ModelInferenceConfig.Token = s.OpenAiAPIKey
		cfg.StorageConfig.DBHost = s.DBHost
		cfg.StorageConfig.DBName = s.DBName
		cfg.StorageConfig.DBUser = s.DBUser
		cfg.StorageConfig.DBPassword = s.DBPassword
	}
}

func loadEnvVarConfig(cfg *Config) {
	cfg.ModelInferenceConfig = openai.ConfigOpenAI{
		Token:     os.Getenv("OPENAI_API_KEY"),
		MaxTokens: utils.MustParseInt(os.Getenv("OPENAI_MAX_TOKENS")),
	}
	cfg.StorageConfig = storage.PgConfig{
		DBHost:     os.Getenv("DB_HOST"),
		DBName:     os.Getenv("DB_DBNAME"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
	}
	if v := os.Getenv("TABLE_PROMPT"); v != "" {
		cfg.StorageConfig.TablePrompt = v
	}
	if v := os.Getenv("TABLE_PREDICTION"); v != "" {
		cfg.StorageConfig.TablePrediction = v
	}
}
