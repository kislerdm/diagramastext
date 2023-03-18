package config

import (
	"context"
	"reflect"
	"testing"

	"github.com/kislerdm/diagramastext/server/core/diagram"
)

func Test_loadDefaultConfig(t *testing.T) {
	type args struct {
		ctx                  context.Context
		clientSecretsManager diagram.RepositorySecretsVault
	}
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
					V: []byte(`{
"db_host": "localhost",
"db_name": "postgres",
"db_user": "postgres",
"db_password": "postgres",
"model_api_key": "foobar"
}`),
				},
			},
			envVars: map[string]string{
				"ACCESS_CREDENTIALS_URI": "foobar",
			},
			want: &Config{
				RepositoryPredictionConfig: repositoryPredictionConfig{
					DBHost:          "localhost",
					DBName:          "postgres",
					DBUser:          "postgres",
					DBPassword:      "postgres",
					TablePrompt:     tableWritePrompt,
					TablePrediction: tableWriteModelPrediction,
				},
				ModelInferenceConfig: modelInferenceConfig{
					Token: "foobar",
				},
			},
		},
		{
			name: "load from secretsmanager, tables from env variables",
			args: args{
				ctx: context.TODO(),
				clientSecretsManager: diagram.MockRepositorySecretsVault{
					V: []byte(`{
"db_host": "localhost",
"db_name": "postgres",
"db_user": "postgres",
"db_password": "postgres",
"model_api_key": "foobar"
}`),
				},
			},
			envVars: map[string]string{
				"ACCESS_CREDENTIALS_URI": "foobar",
				"TABLE_PROMPT":           "foo",
				"TABLE_PREDICTION":       "bar",
				"MODEL_MAX_TOKENS":       "100",
			},
			want: &Config{
				RepositoryPredictionConfig: repositoryPredictionConfig{
					DBHost:          "localhost",
					DBName:          "postgres",
					DBUser:          "postgres",
					DBPassword:      "postgres",
					TablePrompt:     "foo",
					TablePrediction: "bar",
				},
				ModelInferenceConfig: modelInferenceConfig{
					Token:     "foobar",
					MaxTokens: 100,
				},
			},
		},
		{
			name: "env variables",
			args: args{
				ctx: context.TODO(),
			},
			envVars: map[string]string{
				"MODEL_API_KEY":    "foobar",
				"MODEL_MAX_TOKENS": "100",
				"DB_HOST":          "localhost",
				"DB_DBNAME":        "postgres",
				"DB_USER":          "postgres",
				"DB_PASSWORD":      "postgres",
				"TABLE_PROMPT":     "foo",
				"TABLE_PREDICTION": "bar",
			},
			want: &Config{
				RepositoryPredictionConfig: repositoryPredictionConfig{
					DBHost:          "localhost",
					DBName:          "postgres",
					DBUser:          "postgres",
					DBPassword:      "postgres",
					TablePrompt:     "foo",
					TablePrediction: "bar",
				},
				ModelInferenceConfig: modelInferenceConfig{
					Token:     "foobar",
					MaxTokens: 100,
				},
			},
		},
	}

	t.Parallel()

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
}
