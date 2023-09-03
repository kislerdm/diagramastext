package config

import (
	"context"
	"crypto/ed25519"
	"os"
	"strings"

	"github.com/kislerdm/diagramastext/server/core/ciam"
	"github.com/kislerdm/diagramastext/server/core/diagram"
	"github.com/kislerdm/diagramastext/server/core/internal/utils"
)

const (
	defaultSSLMode = "verify-full"

	tableWritePrompt          = "user_prompts"
	tableWriteModelPrediction = "openai_responses"
	tableWriteSuccessStatus   = "successful_requests"
	tableLookupUser           = "users"
	tableLookupApiTokens      = "api_tokens"
	tableOneTimeSecret        = "user_auth_secrets"

	defaultSenderEmail = "support@diagramastext.dev"
	defaultSMPTPort    = "587"
)

type repositoryPredictionConfig struct {
	DBHost             string `json:"db_host"`
	DBName             string `json:"db_name"`
	DBUser             string `json:"db_user"`
	DBPassword         string `json:"db_password"`
	TablePrompt        string `json:"table_prompt"`
	TablePrediction    string `json:"table_prediction"`
	TableSuccessStatus string `json:"table_success_status"`
	TableUsers         string `json:"table_users"`
	TableAPITokens     string `json:"table_api_tokens"`
	SSLMode            string `json:"ssl_mode"`
}

type ciamConfigStore struct {
	PrivateKey         string `json:"private_key"`
	SmtpUser           string `json:"smtp_user"`
	SmtpPassword       string `json:"smtp_password"`
	SmtpHost           string `json:"smtp_host"`
	SmtpPort           string `json:"smtp_port"`
	SmtpSenderEmail    string `json:"smtp_sender_email"`
	TableOneTimeSecret string `json:"table_one_time_secret"`
}

type secret struct {
	repositoryPredictionConfig
	ciamConfigStore
	APIKey string `json:"model_api_key"`
}

type modelInferenceConfig struct {
	Token     string
	MaxTokens int
}

type ciamCfg struct {
	PrivateKey         ed25519.PrivateKey
	TableOneTimeSecret string
	SmtpUser           string
	SmtpPassword       string
	SmtpHost           string
	SmtpPort           string
	SmtpSenderEmail    string
}

type Config struct {
	RepositoryPredictionConfig repositoryPredictionConfig
	CIAM                       ciamCfg
	ModelInferenceConfig       modelInferenceConfig
}

func LoadDefaultConfig(ctx context.Context, clientSecretsManager diagram.RepositorySecretsVault) *Config {
	// defaults
	cfg := Config{
		RepositoryPredictionConfig: repositoryPredictionConfig{
			TablePrompt:        tableWritePrompt,
			TablePrediction:    tableWriteModelPrediction,
			TableSuccessStatus: tableWriteSuccessStatus,
			TableUsers:         tableLookupUser,
			TableAPITokens:     tableLookupApiTokens,
			SSLMode:            defaultSSLMode,
		},
		CIAM: ciamCfg{
			TableOneTimeSecret: tableOneTimeSecret,
			SmtpSenderEmail:    defaultSenderEmail,
			SmtpPort:           defaultSMPTPort,
		},
	}

	loadEnvVarConfig(&cfg)

	if secretARN := os.Getenv("ACCESS_CREDENTIALS_URI"); secretARN != "" && clientSecretsManager != nil {
		loadFromSecretsManager(ctx, &cfg, secretARN, clientSecretsManager)
	}

	return &cfg
}

func loadFromSecretsManager(
	ctx context.Context, cfg *Config, secretURI string, client diagram.RepositorySecretsVault,
) {
	var s secret
	if err := client.ReadLastVersion(ctx, secretURI, &s); err == nil {
		cfg.ModelInferenceConfig.Token = s.APIKey
		cfg.RepositoryPredictionConfig.DBHost = s.DBHost
		cfg.RepositoryPredictionConfig.DBName = s.DBName
		cfg.RepositoryPredictionConfig.DBUser = s.DBUser
		cfg.RepositoryPredictionConfig.DBPassword = s.DBPassword

		if s.PrivateKey != "" {
			var err error
			cfg.CIAM.PrivateKey, err = ciam.ReadPrivateKey(s.PrivateKey)
			if err != nil {
				panic("cannot read key from secret: " + err.Error())
			}
		}

		cfg.CIAM.SmtpUser = s.SmtpUser
		cfg.CIAM.SmtpPassword = s.SmtpPassword
		cfg.CIAM.SmtpHost = s.SmtpHost

		if s.SmtpSenderEmail != "" {
			cfg.CIAM.SmtpSenderEmail = s.SmtpSenderEmail
		}

		if s.SmtpPort != "" {
			cfg.CIAM.SmtpPort = s.SmtpPort
		}
	}
}

func loadEnvVarConfig(cfg *Config) {
	cfg.ModelInferenceConfig.MaxTokens = utils.MustParseInt(os.Getenv("MODEL_MAX_TOKENS"))
	cfg.ModelInferenceConfig.Token = os.Getenv("MODEL_API_KEY")
	cfg.RepositoryPredictionConfig.DBHost = os.Getenv("DB_HOST")
	cfg.RepositoryPredictionConfig.DBName = os.Getenv("DB_DBNAME")
	cfg.RepositoryPredictionConfig.DBUser = os.Getenv("DB_USER")
	cfg.RepositoryPredictionConfig.DBPassword = os.Getenv("DB_PASSWORD")

	if v := os.Getenv("SSL_MODE"); v != "" {
		cfg.RepositoryPredictionConfig.SSLMode = v
	}

	if v := os.Getenv("TABLE_PROMPT"); v != "" {
		cfg.RepositoryPredictionConfig.TablePrompt = v
	}

	if v := os.Getenv("TABLE_PREDICTION"); v != "" {
		cfg.RepositoryPredictionConfig.TablePrediction = v
	}

	if v := os.Getenv("TABLE_SUCCESS_STATUS"); v != "" {
		cfg.RepositoryPredictionConfig.TableSuccessStatus = v
	}

	if v := os.Getenv("TABLE_USERS"); v != "" {
		cfg.RepositoryPredictionConfig.TableUsers = v
	}

	if v := os.Getenv("TABLE_API_TOKENS"); v != "" {
		cfg.RepositoryPredictionConfig.TableAPITokens = v
	}

	if v := os.Getenv("TABLE_ONE_TIME_SECRET"); v != "" {
		cfg.CIAM.TableOneTimeSecret = v
	}

	if v := os.Getenv("ENV"); strings.HasPrefix(strings.ToLower(v), "dev") {
		cfg.CIAM.PrivateKey = ciam.GenerateCertificate()
	}

	if v := os.Getenv("CIAM_SMTP_USER"); v != "" {
		cfg.CIAM.SmtpUser = v
	}

	if v := os.Getenv("CIAM_SMTP_PASSWORD"); v != "" {
		cfg.CIAM.SmtpPassword = v
	}

	if v := os.Getenv("CIAM_SMTP_HOST"); v != "" {
		cfg.CIAM.SmtpHost = v
	}

	if v := os.Getenv("CIAM_SMTP_PORT"); v != "" {
		cfg.CIAM.SmtpPort = v
	}

	if v := os.Getenv("CIAM_SMTP_SENDER_EMAIL"); v != "" {
		cfg.CIAM.SmtpSenderEmail = v
	}
}
