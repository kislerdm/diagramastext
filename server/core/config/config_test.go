package config

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/kislerdm/diagramastext/server/core/diagram"
)

func mustSerialize(v interface{}) []byte {
	o, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return o
}

func Test_loadDefaultConfig(t *testing.T) {
	type args struct {
		ctx                  context.Context
		clientSecretsManager diagram.RepositorySecretsVault
	}

	keyPriv, keyPub := generateDevCertificateKeysPair()

	tests := []struct {
		name    string
		args    args
		envVars map[string]string
		want    *Config
	}{
		{
			name: "load from secretsmanager",
			args: args{
				ctx: context.TODO(),
				clientSecretsManager: diagram.MockRepositorySecretsVault{
					V: mustSerialize(
						secret{
							repositoryPredictionConfig: repositoryPredictionConfig{
								DBHost:     "localhost",
								DBName:     "postgres",
								DBUser:     "postgres",
								DBPassword: "postgres",
							},
							ciamConfigStore: ciamConfigStore{
								PrivateKey:      keyPriv,
								PublicKey:       keyPub,
								SmtpUser:        "foo@bar.baz",
								SmtpPassword:    "qux",
								SmtpHost:        "smtphost",
								SmtpPort:        "573",
								SmtpSenderEmail: "support@bar.baz",
							},
							APIKey: "foobar",
						},
					),
				},
			},
			envVars: map[string]string{
				"ACCESS_CREDENTIALS_URI": "foobar",
			},
			want: &Config{
				RepositoryPredictionConfig: repositoryPredictionConfig{
					DBHost:             "localhost",
					DBName:             "postgres",
					DBUser:             "postgres",
					DBPassword:         "postgres",
					TablePrompt:        tableWritePrompt,
					TablePrediction:    tableWriteModelPrediction,
					TableSuccessStatus: tableWriteSuccessStatus,
					TableUsers:         tableLookupUser,
					TableAPITokens:     tableLookupApiTokens,
					SSLMode:            defaultSSLMode,
				},
				ModelInferenceConfig: modelInferenceConfig{
					Token: "foobar",
				},
				CIAM: ciamCfg{
					TableOneTimeSecret: tableOneTimeSecret,
					SmtpUser:           "foo@bar.baz",
					SmtpPassword:       "qux",
					SmtpHost:           "smtphost",
					SmtpPort:           "573",
					SmtpSenderEmail:    "support@bar.baz",
					PrivateKey:         keyPriv,
					PublicKey:          keyPub,
				},
			},
		},
		{
			name: "load from secretsmanager, overwrite some from env variables",
			args: args{
				ctx: context.TODO(),
				clientSecretsManager: diagram.MockRepositorySecretsVault{
					V: mustSerialize(
						secret{
							repositoryPredictionConfig: repositoryPredictionConfig{
								DBHost:     "localhost",
								DBName:     "postgres",
								DBUser:     "postgres",
								DBPassword: "postgres",
							},
							ciamConfigStore: ciamConfigStore{
								PrivateKey:      keyPriv,
								PublicKey:       keyPub,
								SmtpUser:        "foo@bar.baz",
								SmtpPassword:    "qux",
								SmtpHost:        "smtphost",
								SmtpPort:        "573",
								SmtpSenderEmail: "support@bar.baz",
							},
							APIKey: "foobar",
						},
					),
				},
			},
			envVars: map[string]string{
				"ACCESS_CREDENTIALS_URI": "bazz",
				"MODEL_API_KEY":          "key",
				"DB_HOST":                "dbh",
				"DB_DBNAME":              "dbn",
				"DB_USER":                "dbu",
				"DB_PASSWORD":            "dbpass",
				"MODEL_MAX_TOKENS":       "100",
				"TABLE_PROMPT":           "foo",
				"TABLE_PREDICTION":       "bar",
				"TABLE_SUCCESS_STATUS":   "qux",
				"TABLE_USERS":            "u",
				"TABLE_API_TOKENS":       "t",
				"TABLE_ONE_TIME_SECRET":  "s",
				"SSL_MODE":               "disable",
				"CIAM_SMTP_USER":         "r",
				"CIAM_SMTP_PASSWORD":     "t",
				"CIAM_SMTP_HOST":         "yy",
				"CIAM_SMTP_PORT":         "44",
				"CIAM_SMTP_SENDER_EMAIL": "dfdf",
			},
			want: &Config{
				RepositoryPredictionConfig: repositoryPredictionConfig{
					DBHost:             "localhost",
					DBName:             "postgres",
					DBUser:             "postgres",
					DBPassword:         "postgres",
					TablePrompt:        "foo",
					TablePrediction:    "bar",
					TableSuccessStatus: "qux",
					TableUsers:         "u",
					TableAPITokens:     "t",
					SSLMode:            "disable",
				},
				CIAM: ciamCfg{
					TableOneTimeSecret: "s",
					SmtpUser:           "foo@bar.baz",
					SmtpPassword:       "qux",
					SmtpHost:           "smtphost",
					SmtpPort:           "573",
					SmtpSenderEmail:    "support@bar.baz",
					PrivateKey:         keyPriv,
					PublicKey:          keyPub,
				},
				ModelInferenceConfig: modelInferenceConfig{
					Token:     "foobar",
					MaxTokens: 100,
				},
			},
		},
		{
			name: "env variables only",
			args: args{
				ctx: context.TODO(),
			},
			envVars: map[string]string{
				"MODEL_API_KEY":          "foobar",
				"MODEL_MAX_TOKENS":       "100",
				"DB_HOST":                "localhost",
				"DB_DBNAME":              "postgres",
				"DB_USER":                "postgres",
				"DB_PASSWORD":            "postgres",
				"TABLE_PROMPT":           "foo",
				"TABLE_PREDICTION":       "bar",
				"TABLE_SUCCESS_STATUS":   "qux",
				"TABLE_USERS":            "u",
				"TABLE_ONE_TIME_SECRET":  "s",
				"TABLE_API_TOKENS":       "t",
				"CIAM_SMTP_USER":         "r",
				"CIAM_SMTP_PASSWORD":     "t",
				"CIAM_SMTP_HOST":         "yy",
				"CIAM_SMTP_PORT":         "44",
				"CIAM_SMTP_SENDER_EMAIL": "dfdf",
			},
			want: &Config{
				RepositoryPredictionConfig: repositoryPredictionConfig{
					DBHost:             "localhost",
					DBName:             "postgres",
					DBUser:             "postgres",
					DBPassword:         "postgres",
					TablePrompt:        "foo",
					TablePrediction:    "bar",
					TableSuccessStatus: "qux",
					TableUsers:         "u",
					TableAPITokens:     "t",
					SSLMode:            defaultSSLMode,
				},
				ModelInferenceConfig: modelInferenceConfig{
					Token:     "foobar",
					MaxTokens: 100,
				},
				CIAM: ciamCfg{
					TableOneTimeSecret: "s",
					SmtpUser:           "r",
					SmtpPassword:       "t",
					SmtpHost:           "yy",
					SmtpPort:           "44",
					SmtpSenderEmail:    "dfdf",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				for k, v := range tt.envVars {
					t.Setenv(k, v)
				}

				if got := LoadDefaultConfig(tt.args.ctx, tt.args.clientSecretsManager); !reflect.DeepEqual(
					got, tt.want,
				) {
					t.Errorf("LoadDefaultConfig() = %v, want %v", got, tt.want)
				}
			},
		)
	}

	t.Run(
		"shall set CIAM keys if ENV envvar is set to 'dev'", func(t *testing.T) {
			// GIVEN
			t.Setenv("ENV", "dev")

			// WHEN
			got := LoadDefaultConfig(context.TODO(), nil)

			// THEN
			if got.CIAM.PublicKey == nil || got.CIAM.PrivateKey == nil {
				t.Error("CIAM signing keys are not defined for dev environment")
			}
		},
	)
	t.Run(
		"shall set CIAM keys if ENV envvar is set to 'development'", func(t *testing.T) {
			// GIVEN
			t.Setenv("ENV", "development")

			// WHEN
			got := LoadDefaultConfig(context.TODO(), nil)

			// THEN
			if got.CIAM.PublicKey == nil || got.CIAM.PrivateKey == nil {
				t.Error("CIAM signing keys are not defined for dev environment")
			}
		},
	)
}
