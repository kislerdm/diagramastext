package core

//func TestLoadDefaultConfig(t *testing.T) {
//	type args struct {
//		ctx      context.Context
//		awsOptFn []func(*config.LoadOptions) error
//	}
//	tests := []struct {
//		name    string
//		args    args
//		envVars map[string]string
//		want    Config
//		wantErr bool
//	}{
//		{
//			name: "happy path: envvar",
//			args: args{
//				ctx: context.TODO(),
//			},
//			envVars: map[string]string{
//				"OPENAI_API_KEY":    "sk-xxxxxx",
//				"OPENAI_MAX_TOKENS": "100",
//				"DB_HOST":           "localhost",
//				"DB_DBNAME":         "postgres",
//				"DB_USER":           "postgres",
//				"DB_PASSWORD":       "postgres",
//			},
//			want: Config{
//				StorageConfig: storageConfig{
//					DBHost:     "localhost",
//					DBName:     "postgres",
//					DBUser:     "postgres",
//					DBPassword: "postgres",
//				},
//				ModelInferenceConfig: openai.ConfigOpenAI{
//					MaxTokens: 100,
//					Token:     "sk-xxxxxx",
//				},
//			},
//			wantErr: false,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(
//			tt.name, func(t *testing.T) {
//				for k, v := range tt.envVars {
//					t.Setenv(k, v)
//				}
//				got, err := loadDefaultConfig(
//					tt.args.ctx,
//				)
//				if (err != nil) != tt.wantErr {
//					t.Errorf("loadDefaultConfig() error = %v, wantErr %v", err, tt.wantErr)
//					return
//				}
//				if !reflect.DeepEqual(got, tt.want) {
//					t.Errorf("loadDefaultConfig() got = %v, want %v", got, tt.want)
//				}
//			},
//		)
//	}
//}
