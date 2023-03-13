package core

//import (
//	"context"
//	"os"
//
//	"github.com/kislerdm/diagramastext/server/core/contract"
//	"github.com/kislerdm/diagramastext/server/core/openai"
//	"github.com/kislerdm/diagramastext/server/core/port"
//	"github.com/kislerdm/diagramastext/server/core/postgres"
//	"github.com/kislerdm/diagramastext/server/core/utils"
//)
//
//// InitEntrypoint initialises the `core` entrypoint function.
//func InitEntrypoint(ctx context.Context, clientSecretsManager contract.ClientSecretsmanager) (
//	contract.Entrypoint, error,
//) {
//	clientModelInference, _, err := initCoreClients(ctx, clientSecretsManager)
//	if err != nil {
//		return nil, err
//	}
//
//	return func(ctx context.Context, inquery port.Input, diagramImplementationHandler contract.DiagramHandler) (
//		[]byte, error,
//	) {
//		if err := inquery.Validate(); err != nil {
//			return nil, err
//		}
//
//		modelAttrs := diagramImplementationHandler.GetModelRequest(inquery)
//
//		diagramPrediction, err := clientModelInference.Do(ctx, modelAttrs.Prompt, modelAttrs.Model)
//		if err != nil {
//			return nil, err
//		}
//
//		//storePrediction(ctx, clientStorage, inquery.UserID, inquery.Prompt, string(diagramPrediction))
//
//		diagramSVG, err := diagramImplementationHandler.RenderPredictionResultsDiagramSVG(ctx, diagramPrediction)
//		if err != nil {
//			return nil, err
//		}
//
//		if err := utils.ValidateSVG(diagramSVG); err != nil {
//			return nil, err
//		}
//
//		return diagramSVG, nil
//
//	}, nil
//}
//
//func storePrediction(
//	ctx context.Context, clientStorage contract.ClientStorage, userID string, prompt string, graphPrediction string,
//) {
//	requestID := utils.NewUUID()
//	// FIXME: combine both transactions into a single atomic transaction.
//	if err := clientStorage.WritePrompt(ctx, requestID, prompt, userID); err == nil {
//		_ = clientStorage.WriteModelPrediction(ctx, requestID, graphPrediction, userID)
//	}
//}
//
//// initCoreClients initialises `core` clients.
//func initCoreClients(ctx context.Context, clientSecretsManager contract.ClientSecretsmanager) (
//	contract.ClientModelInference, contract.ClientStorage, error,
//) {
//	cfg := loadDefaultConfig(ctx, clientSecretsManager)
//
//	// TODO: add custom http client with the retry-backoff mechanism.
//	clientModelInference, err := openai.NewClient(cfg.modelInferenceConfig)
//	if err != nil {
//		return nil, nil, err
//	}
//
//	clientStorage, err := postgres.NewClient(ctx, cfg.storageConfig)
//	if err != nil {
//		return nil, nil, err
//	}
//
//	return clientModelInference, clientStorage, nil
//}
//
//type config struct {
//	modelInferenceConfig openai.ConfigOpenAI
//	storageConfig        postgres.Config
//}
//
//type secret struct {
//	DBHost       string `json:"db_host"`
//	DBName       string `json:"db_name"`
//	DBUser       string `json:"db_user"`
//	DBPassword   string `json:"db_password"`
//	OpenAiAPIKey string `json:"openai_api_key"`
//}
//
//const (
//	tableWritePrompt          = "user_prompt"
//	tableWriteModelPrediction = "openai_response"
//)
//
//func loadDefaultConfig(ctx context.Context, clientSecretsManager contract.ClientSecretsmanager) *config {
//	// defaults
//	cfg := config{
//		storageConfig: postgres.Config{
//			TablePrompt:     tableWritePrompt,
//			TablePrediction: tableWriteModelPrediction,
//		},
//	}
//
//	loadEnvVarConfig(&cfg)
//
//	if secretARN := os.Getenv("ACCESS_CREDENTIALS_ARN"); secretARN != "" && clientSecretsManager != nil {
//		loadFromSecretsManager(ctx, &cfg, secretARN, clientSecretsManager)
//	}
//
//	return &cfg
//}
//
//func loadFromSecretsManager(
//	ctx context.Context, cfg *config, secretARN string, client contract.ClientSecretsmanager,
//) {
//	var s secret
//	if err := client.ReadLatestSecret(ctx, secretARN, &s); err == nil {
//		cfg.modelInferenceConfig.Token = s.OpenAiAPIKey
//		cfg.storageConfig.DBHost = s.DBHost
//		cfg.storageConfig.DBName = s.DBName
//		cfg.storageConfig.DBUser = s.DBUser
//		cfg.storageConfig.DBPassword = s.DBPassword
//	}
//}
//
//func loadEnvVarConfig(cfg *config) {
//	cfg.modelInferenceConfig = openai.ConfigOpenAI{
//		Token:     os.Getenv("OPENAI_API_KEY"),
//		MaxTokens: utils.MustParseInt(os.Getenv("OPENAI_MAX_TOKENS")),
//	}
//	cfg.storageConfig = postgres.Config{
//		DBHost:     os.Getenv("DB_HOST"),
//		DBName:     os.Getenv("DB_DBNAME"),
//		DBUser:     os.Getenv("DB_USER"),
//		DBPassword: os.Getenv("DB_PASSWORD"),
//	}
//	if v := os.Getenv("TABLE_PROMPT"); v != "" {
//		cfg.storageConfig.TablePrompt = v
//	}
//	if v := os.Getenv("TABLE_PREDICTION"); v != "" {
//		cfg.storageConfig.TablePrediction = v
//	}
//}
