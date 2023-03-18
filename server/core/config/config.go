package config

import (
	"context"
	"os"

	"github.com/kislerdm/diagramastext/server/core/port"
	"github.com/kislerdm/diagramastext/server/core/utils"
)

const (
	tableWritePrompt          = "user_prompt"
	tableWriteModelPrediction = "openai_response"
)

type repositoryPredictionConfig struct {
	DBHost          string `json:"db_host"`
	DBName          string `json:"db_name"`
	DBUser          string `json:"db_user"`
	DBPassword      string `json:"db_password"`
	TablePrompt     string `json:"table_prompt"`
	TablePrediction string `json:"table_prediction"`
}

type secret struct {
	repositoryPredictionConfig
	APIKey string `json:"model_api_key"`
}

type modelInferenceConfig struct {
	Token     string
	MaxTokens int
}

type Config struct {
	RepositoryPredictionConfig repositoryPredictionConfig
	ModelInferenceConfig       modelInferenceConfig
}

func LoadDefaultConfig(ctx context.Context, clientSecretsManager port.RepositorySecretsVault) *Config {
	// defaults
	cfg := Config{
		RepositoryPredictionConfig: repositoryPredictionConfig{
			TablePrompt:     tableWritePrompt,
			TablePrediction: tableWriteModelPrediction,
		},
	}

	loadEnvVarConfig(&cfg)

	if secretARN := os.Getenv("ACCESS_CREDENTIALS_URI"); secretARN != "" && clientSecretsManager != nil {
		loadFromSecretsManager(ctx, &cfg, secretARN, clientSecretsManager)
	}

	return &cfg
}

func loadFromSecretsManager(
	ctx context.Context, cfg *Config, secretURI string, client port.RepositorySecretsVault,
) {
	var s secret
	if err := client.ReadLastVersion(ctx, secretURI, &s); err == nil {
		cfg.ModelInferenceConfig.Token = s.APIKey
		cfg.RepositoryPredictionConfig.DBHost = s.DBHost
		cfg.RepositoryPredictionConfig.DBName = s.DBName
		cfg.RepositoryPredictionConfig.DBUser = s.DBUser
		cfg.RepositoryPredictionConfig.DBPassword = s.DBPassword
	}
}

func loadEnvVarConfig(cfg *Config) {
	cfg.ModelInferenceConfig.Token = os.Getenv("MODEL_API_KEY")
	cfg.ModelInferenceConfig.MaxTokens = utils.MustParseInt(os.Getenv("MODEL_MAX_TOKENS"))
	cfg.RepositoryPredictionConfig.DBHost = os.Getenv("DB_HOST")
	cfg.RepositoryPredictionConfig.DBName = os.Getenv("DB_DBNAME")
	cfg.RepositoryPredictionConfig.DBUser = os.Getenv("DB_USER")
	cfg.RepositoryPredictionConfig.DBPassword = os.Getenv("DB_PASSWORD")

	if v := os.Getenv("TABLE_PROMPT"); v != "" {
		cfg.RepositoryPredictionConfig.TablePrompt = v
	}

	if v := os.Getenv("TABLE_PREDICTION"); v != "" {
		cfg.RepositoryPredictionConfig.TablePrediction = v
	}
}
