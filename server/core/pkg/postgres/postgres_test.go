package postgres

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestConfig_Validate(t *testing.T) {
	type fields struct {
		DBHost             string
		DBName             string
		DBUser             string
		DBPassword         string
		TablePrompt        string
		TablePrediction    string
		TableSuccessStatus string
		TableUsers         string
		TableTokens        string
		TableOneTimeSecret string
		SSLMode            string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr error
	}{
		{
			name: "valid",
			fields: fields{
				DBHost:             "localhost",
				DBName:             "postgres",
				DBUser:             "postgres",
				DBPassword:         "postgres",
				TablePrompt:        "foo",
				TablePrediction:    "bar",
				TableSuccessStatus: "qux",
				TableUsers:         "quxx",
				TableTokens:        "baz",
				TableOneTimeSecret: "foobar",
			},
			wantErr: nil,
		},
		{
			name: "valid: ssl - full verification",
			fields: fields{
				DBHost:             "localhost",
				DBName:             "postgres",
				DBUser:             "postgres",
				DBPassword:         "postgres",
				TablePrompt:        "foo",
				TablePrediction:    "bar",
				TableSuccessStatus: "qux",
				TableUsers:         "quxx",
				TableTokens:        "baz",
				TableOneTimeSecret: "foobar",
				SSLMode:            "verify-full",
			},
			wantErr: nil,
		},
		{
			name: "invalid: host is missing",
			fields: fields{
				DBHost:             "",
				DBName:             "postgres",
				DBUser:             "postgres",
				DBPassword:         "postgres",
				TablePrompt:        "foo",
				TablePrediction:    "bar",
				TableSuccessStatus: "qux",
				TableUsers:         "quxx",
				TableTokens:        "baz",
				TableOneTimeSecret: "foobar",
			},
			wantErr: errors.New("host must be provided"),
		},
		{
			name: "invalid: dbname is missing",
			fields: fields{
				DBHost:             "localhost",
				DBName:             "",
				DBUser:             "postgres",
				DBPassword:         "postgres",
				TablePrompt:        "foo",
				TablePrediction:    "bar",
				TableSuccessStatus: "qux",
				TableUsers:         "quxx",
				TableTokens:        "baz",
				TableOneTimeSecret: "foobar",
			},
			wantErr: errors.New("dbname must be provided"),
		},
		{
			name: "invalid: user is missing",
			fields: fields{
				DBHost:             "localhost",
				DBName:             "postgres",
				DBUser:             "",
				DBPassword:         "postgres",
				TablePrompt:        "foo",
				TablePrediction:    "bar",
				TableSuccessStatus: "qux",
				TableUsers:         "",
				TableTokens:        "baz",
				TableOneTimeSecret: "foobar",
			},
			wantErr: errors.New("user must be provided"),
		},
		{
			name: "invalid: table_prompt is missing",
			fields: fields{
				DBHost:             "localhost",
				DBName:             "postgres",
				DBUser:             "postgres",
				DBPassword:         "postgres",
				TablePrompt:        "",
				TablePrediction:    "bar",
				TableSuccessStatus: "qux",
				TableUsers:         "users",
				TableTokens:        "tokens",
				TableOneTimeSecret: "foobar",
			},
			wantErr: errors.New("table_prompt must be provided"),
		},
		{
			name: "invalid: table_prediction is missing",
			fields: fields{
				DBHost:             "localhost",
				DBName:             "postgres",
				DBUser:             "postgres",
				DBPassword:         "postgres",
				TablePrompt:        "foo",
				TablePrediction:    "",
				TableSuccessStatus: "qux",
				TableUsers:         "users",
				TableTokens:        "tokens",
				TableOneTimeSecret: "foobar",
			},
			wantErr: errors.New("table_prediction must be provided"),
		},
		{
			name: "invalid: table_success_status is missing",
			fields: fields{
				DBHost:             "localhost",
				DBName:             "postgres",
				DBUser:             "postgres",
				DBPassword:         "postgres",
				TablePrompt:        "foo",
				TablePrediction:    "bar",
				TableUsers:         "users",
				TableTokens:        "tokens",
				TableSuccessStatus: "",
				TableOneTimeSecret: "foobar",
			},
			wantErr: errors.New("table_success_status must be provided"),
		},
		{
			name: "invalid: table_one_time_secret is missing",
			fields: fields{
				DBHost:             "localhost",
				DBName:             "postgres",
				DBUser:             "postgres",
				DBPassword:         "postgres",
				TablePrompt:        "foo",
				TablePrediction:    "bar",
				TableSuccessStatus: "qux",
				TableUsers:         "users",
				TableTokens:        "tokens",
				TableOneTimeSecret: "",
			},
			wantErr: errors.New("table_one_time_secret must be provided"),
		},
		{
			name: "invalid: table_tokens is missing",
			fields: fields{
				DBHost:             "localhost",
				DBName:             "postgres",
				DBUser:             "postgres",
				DBPassword:         "postgres",
				TablePrompt:        "foo",
				TablePrediction:    "bar",
				TableSuccessStatus: "qux",
				TableUsers:         "users",
				TableTokens:        "",
				TableOneTimeSecret: "quxx",
			},
			wantErr: errors.New("table_tokens must be provided"),
		},
		{
			name: "invalid: ssl mode is wrong",
			fields: fields{
				DBHost:             "localhost",
				DBName:             "postgres",
				DBUser:             "postgres",
				DBPassword:         "postgres",
				TablePrompt:        "foo",
				TablePrediction:    "bar",
				SSLMode:            "qux",
				TableSuccessStatus: "quxx",
				TableUsers:         "quxx",
				TableTokens:        "baz",
				TableOneTimeSecret: "foobar",
			},
			wantErr: errors.New("ssl mode qux is not supported"),
		},
	}

	t.Parallel()

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				cfg := Config{
					DBHost:             tt.fields.DBHost,
					DBName:             tt.fields.DBName,
					DBUser:             tt.fields.DBUser,
					DBPassword:         tt.fields.DBPassword,
					TablePrompt:        tt.fields.TablePrompt,
					TablePrediction:    tt.fields.TablePrediction,
					TableSuccessStatus: tt.fields.TableSuccessStatus,
					TableUsers:         tt.fields.TableUsers,
					TableTokens:        tt.fields.TableTokens,
					TableOneTimeSecret: tt.fields.TableOneTimeSecret,
					SSLMode:            tt.fields.SSLMode,
				}
				err := cfg.Validate()
				if !reflect.DeepEqual(err, tt.wantErr) {
					t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				}
			},
		)
	}
}

func TestNewRepositoryPostgres(t *testing.T) {
	type args struct {
		ctx context.Context
		cfg Config
	}
	tests := []struct {
		name    string
		args    args
		want    *Client
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				ctx: context.TODO(),
				cfg: Config{
					DBHost:             "mock",
					DBName:             "postgres",
					DBUser:             "postgres",
					DBPassword:         "foo",
					TablePrompt:        "bar",
					TablePrediction:    "baz",
					TableSuccessStatus: "qux",
					TableUsers:         "quxx",
					TableTokens:        "baz",
					TableOneTimeSecret: "quxxx",
				},
			},
			want: &Client{
				c:                         &mockDbClient{},
				tableWritePrompt:          "bar",
				tableWriteModelPrediction: "baz",
				tableWriteSuccessFlag:     "qux",
				tableUsers:                "quxx",
				tableTokens:               "baz",
				tableOneTimeSecret:        "quxxx",
			},
			wantErr: false,
		},
		{
			name: "unhappy path: no host",
			args: args{
				ctx: context.TODO(),
				cfg: Config{
					DBHost:          "",
					DBName:          "postgres",
					DBUser:          "postgres",
					TablePrompt:     "foo",
					TablePrediction: "bar",
				},
			},
			want:    nil,
			wantErr: true,
		},
	}

	t.Parallel()

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := NewPostgresClient(tt.args.ctx, tt.args.cfg)
				if (err != nil) != tt.wantErr {
					t.Errorf("NewPostgresClient() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("NewPostgresClient() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestClient_WritePrompt(t *testing.T) {
	type fields struct {
		c dbClient
	}
	type args struct {
		ctx                       context.Context
		requestID, userID, prompt string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name:   "happy path",
			fields: fields{&mockDbClient{}},
			args: args{
				ctx:       context.TODO(),
				prompt:    "c4 diagram of four boxes",
				requestID: "693a35ba-e42c-4168-8afc-5a7c359d1d05",
				userID:    "c40bad11-0822-4d84-9f61-44b9a97b0432",
			},
			wantErr: nil,
		},
		{
			name:   "unhappy path: no request id",
			fields: fields{&mockDbClient{}},
			args: args{
				ctx:    context.TODO(),
				prompt: "c4 diagram of four boxes",
				userID: "c40bad11-0822-4d84-9f61-44b9a97b0432",
			},
			wantErr: errors.New("request_id is required"),
		},
		{
			name:   "unhappy path: no prompt",
			fields: fields{&mockDbClient{}},
			args: args{
				ctx:       context.TODO(),
				requestID: "693a35ba-e42c-4168-8afc-5a7c359d1d05",
				userID:    "c40bad11-0822-4d84-9f61-44b9a97b0432",
			},
			wantErr: errors.New("prompt is required"),
		},
		{
			name: "unhappy path: no table found",
			fields: fields{
				&mockDbClient{
					err: errors.New(`pq: relation "foo" does not exist`),
				},
			},
			args: args{
				ctx:       context.TODO(),
				prompt:    "c4 diagram of four boxes",
				requestID: "693a35ba-e42c-4168-8afc-5a7c359d1d05",
				userID:    "c40bad11-0822-4d84-9f61-44b9a97b0432",
			},
			wantErr: errors.New(`pq: relation "foo" does not exist`),
		},
	}

	t.Parallel()

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c := Client{
					c:                         tt.fields.c,
					tableWritePrompt:          "foo",
					tableWriteModelPrediction: "bar",
				}
				if err := c.WriteInputPrompt(
					tt.args.ctx, tt.args.requestID, tt.args.userID, tt.args.prompt,
				); !reflect.DeepEqual(err, tt.wantErr) {
					t.Errorf("WriteInputPrompt() error = %v, wantErr %v", err, tt.wantErr)
				}
			},
		)
	}
}

func TestClient_WriteModelResult(t *testing.T) {
	type fields struct {
		c dbClient
	}
	type args struct {
		ctx                                                 context.Context
		requestID, userID, predictionRaw, prediction, model string
		usageTokensPrompt, usageTokensCompletions           uint16
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "happy path",
			fields: fields{
				c: &mockDbClient{},
			},
			args: args{
				ctx:                    context.TODO(),
				requestID:              "693a35ba-e42c-4168-8afc-5a7c359d1d05",
				userID:                 "c40bad11-0822-4d84-9f61-44b9a97b0432",
				predictionRaw:          `{"nodes":[{"id":"0"},{"id":"1"}]}`,
				prediction:             `{"nodes":[{"id":"0"},{"id":"1"}]}`,
				model:                  "foobar",
				usageTokensPrompt:      100,
				usageTokensCompletions: 50,
			},
			wantErr: nil,
		},
		{
			name: "unhappy path: no request id",
			fields: fields{
				c: &mockDbClient{},
			},
			args: args{
				ctx:                    context.TODO(),
				userID:                 "c40bad11-0822-4d84-9f61-44b9a97b0432",
				predictionRaw:          `{"nodes":[{"id":"0"},{"id":"1"}]}`,
				prediction:             `{"nodes":[{"id":"0"},{"id":"1"}]}`,
				model:                  "foobar",
				usageTokensPrompt:      100,
				usageTokensCompletions: 50,
			},
			wantErr: errors.New("request_id is required"),
		},
		{
			name: "unhappy path: no request id",
			fields: fields{
				c: &mockDbClient{},
			},
			args: args{
				ctx:                    context.TODO(),
				requestID:              "693a35ba-e42c-4168-8afc-5a7c359d1d05",
				predictionRaw:          `{"nodes":[{"id":"0"},{"id":"1"}]}`,
				prediction:             `{"nodes":[{"id":"0"},{"id":"1"}]}`,
				model:                  "foobar",
				usageTokensPrompt:      100,
				usageTokensCompletions: 50,
			},
			wantErr: errors.New("user_id is required"),
		},
		{
			name: "unhappy path: no response",
			fields: fields{
				c: &mockDbClient{},
			},
			args: args{
				ctx:                    context.TODO(),
				requestID:              "693a35ba-e42c-4168-8afc-5a7c359d1d05",
				userID:                 "c40bad11-0822-4d84-9f61-44b9a97b0432",
				predictionRaw:          `{"nodes":[{"id":"0"},{"id":"1"}]}`,
				model:                  "foobar",
				usageTokensPrompt:      100,
				usageTokensCompletions: 50,
			},
			wantErr: errors.New("response is required"),
		},
		{
			name: "unhappy path: no raw response",
			fields: fields{
				c: &mockDbClient{},
			},
			args: args{
				ctx:                    context.TODO(),
				requestID:              "693a35ba-e42c-4168-8afc-5a7c359d1d05",
				userID:                 "c40bad11-0822-4d84-9f61-44b9a97b0432",
				prediction:             `{"nodes":[{"id":"0"},{"id":"1"}]}`,
				model:                  "foobar",
				usageTokensPrompt:      100,
				usageTokensCompletions: 50,
			},
			wantErr: errors.New("raw response is required"),
		},
		{
			name: "unhappy path: no model",
			fields: fields{
				c: &mockDbClient{},
			},
			args: args{
				ctx:                    context.TODO(),
				requestID:              "693a35ba-e42c-4168-8afc-5a7c359d1d05",
				userID:                 "c40bad11-0822-4d84-9f61-44b9a97b0432",
				predictionRaw:          `{"nodes":[{"id":"0"},{"id":"1"}]}`,
				prediction:             `{"nodes":[{"id":"0"},{"id":"1"}]}`,
				usageTokensPrompt:      100,
				usageTokensCompletions: 50,
			},
			wantErr: errors.New("model is required"),
		},
		{
			name: "unhappy path: no relation found",
			fields: fields{
				c: &mockDbClient{
					err: errors.New(`pq: relation "bar" does not exist`),
				},
			},
			args: args{
				ctx:                    context.TODO(),
				requestID:              "693a35ba-e42c-4168-8afc-5a7c359d1d05",
				userID:                 "c40bad11-0822-4d84-9f61-44b9a97b0432",
				predictionRaw:          `{"nodes":[{"id":"0"},{"id":"1"}]}`,
				prediction:             `{"nodes":[{"id":"0"},{"id":"1"}]}`,
				model:                  "foobar",
				usageTokensPrompt:      100,
				usageTokensCompletions: 50,
			},
			wantErr: errors.New(`pq: relation "bar" does not exist`),
		},
	}

	t.Parallel()

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c := Client{
					c:                         tt.fields.c,
					tableWritePrompt:          "foo",
					tableWriteModelPrediction: "bar",
				}
				err := c.WriteModelResult(
					tt.args.ctx, tt.args.requestID, tt.args.userID, tt.args.predictionRaw, tt.args.prediction,
					tt.args.model, tt.args.usageTokensPrompt, tt.args.usageTokensCompletions,
				)
				if !reflect.DeepEqual(tt.wantErr, err) {
					t.Errorf("WriteModelResult() error = %v, wantErr %v", err, tt.wantErr)
				}
			},
		)
	}
}

func TestClient_Close(t *testing.T) {
	t.Parallel()

	t.Run(
		"happy path", func(t *testing.T) {
			// GIVEN
			c := Client{
				c: &mockDbClient{},
			}
			// WHEN
			err := c.Close(context.TODO())
			// THEN
			if err != nil {
				t.Errorf("unexpected error")
			}
		},
	)
	t.Run(
		"unhappy path", func(t *testing.T) {
			// GIVEN
			c := Client{
				c: &mockDbClient{err: errors.New("error disconnecting")},
			}
			// WHEN
			err := c.Close(context.TODO())
			// THEN
			if err == nil {
				t.Errorf("expected error")
			}
		},
	)
}

func TestClient_WriteSuccessFlag(t *testing.T) {
	type fields struct {
		c dbClient
	}
	type args struct {
		ctx                      context.Context
		requestID, userID, token string
	}

	const table = "qux"

	tests := []struct {
		name                      string
		fields                    fields
		args                      args
		wantExecutedQueryTemplate string
		wantErr                   error
	}{
		{
			name:   "happy path",
			fields: fields{&mockDbClient{}},
			args: args{
				ctx:       context.TODO(),
				requestID: "693a35ba-e42c-4168-8afc-5a7c359d1d05",
				userID:    "c40bad11-0822-4d84-9f61-44b9a97b0432",
				token:     "1410904f-f646-488f-ae08-cc341dfb321c",
			},
			wantExecutedQueryTemplate: `INSERT INTO ` + table +
				` (
	 request_id
   , user_id
   , timestamp
   , token
 ) VALUES (
		   $1
		 , $2
	     , $3
	     , $4
)`,
			wantErr: nil,
		},
		{
			name:   "happy path: no token",
			fields: fields{&mockDbClient{}},
			args: args{
				ctx:       context.TODO(),
				requestID: "693a35ba-e42c-4168-8afc-5a7c359d1d05",
				userID:    "c40bad11-0822-4d84-9f61-44b9a97b0432",
			},
			wantExecutedQueryTemplate: `INSERT INTO ` + table +
				` (
	request_id
  , user_id
  , timestamp
) VALUES (
		  $1
    	, $2
    	, $3
)`,
			wantErr: nil,
		},
		{
			name: "unhappy path: no request id",
			fields: fields{
				c: &mockDbClient{},
			},
			args: args{
				ctx:    context.TODO(),
				userID: "c40bad11-0822-4d84-9f61-44b9a97b0432",
				token:  "1410904f-f646-488f-ae08-cc341dfb321c",
			},
			wantErr: errors.New("request_id is required"),
		},
		{
			name: "unhappy path: no user id",
			fields: fields{
				c: &mockDbClient{},
			},
			args: args{
				ctx:       context.TODO(),
				requestID: "693a35ba-e42c-4168-8afc-5a7c359d1d05",
				token:     "1410904f-f646-488f-ae08-cc341dfb321c",
			},
			wantErr: errors.New("user_id is required"),
		},
	}

	t.Parallel()

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c := Client{
					c:                     tt.fields.c,
					tableWriteSuccessFlag: table,
				}
				err := c.WriteSuccessFlag(
					tt.args.ctx, tt.args.requestID, tt.args.userID, tt.args.token,
				)
				if !reflect.DeepEqual(err, tt.wantErr) {
					t.Errorf("WriteSuccessFlag() error = %v, wantErr %v", err, tt.wantErr)
				}
				gotQueryExecuted := c.c.(*mockDbClient).query
				if gotQueryExecuted != tt.wantExecutedQueryTemplate {
					t.Errorf(
						"WriteSuccessFlag() executes wrong query = %s, want = %s",
						gotQueryExecuted, tt.wantExecutedQueryTemplate,
					)
				}
			},
		)
	}
}

func TestClient_GetActiveUserIDByActiveTokenID(t *testing.T) {
	type fields struct {
		c dbClient
	}
	type args struct {
		ctx   context.Context
		token string
	}

	const (
		tableUsers  = "foo"
		tableTokens = "bar"
	)

	tests := []struct {
		name                      string
		fields                    fields
		args                      args
		wantExecutedQueryTemplate string
		want                      string
		wantErr                   error
	}{
		{
			name: "happy path",
			fields: fields{
				&mockDbClient{
					v: &mockRows{
						tag: pgconn.NewCommandTag("SELECT"),
						v:   [][]any{{"c40bad11-0822-4d84-9f61-44b9a97b0432"}},
						s:   &sync.RWMutex{},
					},
				},
			},
			args: args{
				ctx:   context.TODO(),
				token: "1410904f-f646-488f-ae08-cc341dfb321c",
			},
			want: "c40bad11-0822-4d84-9f61-44b9a97b0432",
			wantExecutedQueryTemplate: `SELECT u.user_id 
FROM ` + tableUsers + ` AS u 
INNER JOIN ` + tableTokens + ` AS t USING (user_id) 
WHERE t.token = $1 AND t.is_active AND u.is_active`,
			wantErr: nil,
		},
		{
			name: "unhappy path: table not found",
			fields: fields{
				&mockDbClient{
					err: errors.New("foobar"),
				},
			},
			args: args{
				ctx:   context.TODO(),
				token: "1410904f-f646-488f-ae08-cc341dfb321c",
			},
			wantExecutedQueryTemplate: `SELECT u.user_id 
FROM ` + tableUsers + ` AS u 
INNER JOIN ` + tableTokens + ` AS t USING (user_id) 
WHERE t.token = $1 AND t.is_active AND u.is_active`,
			wantErr: errors.New("foobar"),
		},
	}

	t.Parallel()

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c := Client{
					c:           tt.fields.c,
					tableUsers:  tableUsers,
					tableTokens: tableTokens,
				}
				got, err := c.GetActiveUserIDByActiveTokenID(
					tt.args.ctx, tt.args.token,
				)
				if !reflect.DeepEqual(err, tt.wantErr) {
					t.Errorf("WriteSuccessFlag() error = %v, wantErr %v", err, tt.wantErr)
				}
				gotQueryExecuted := c.c.(*mockDbClient).query
				if gotQueryExecuted != tt.wantExecutedQueryTemplate {
					t.Errorf(
						"GetActiveUserIDByActiveTokenID() executes wrong query = %s, want = %s",
						gotQueryExecuted, tt.wantExecutedQueryTemplate,
					)
				}
				if got != tt.want {
					t.Errorf("GetActiveUserIDByActiveTokenID() unexpected result = %s, want = %s", got, tt.want)
				}
			},
		)
	}
}

func TestConfig_ConnectionString(t *testing.T) {
	type fields struct {
		DBHost             string
		DBName             string
		DBUser             string
		DBPassword         string
		TablePrompt        string
		TablePrediction    string
		TableSuccessStatus string
		TableUsers         string
		TableTokens        string
		SSLMode            string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "neon-like example",
			fields: fields{
				DBHost:     "aa-foobar-quxquxx-123456.us-east-2.aws.neon.tech",
				DBName:     "foo",
				DBUser:     "bar",
				DBPassword: "baz",
				SSLMode:    "verify-full",
			},
			want: "postgres://bar:baz@aa-foobar-quxquxx-123456.us-east-2.aws.neon.tech/foo?sslmode=verify-full",
		},
		{
			name: "localenv",
			fields: fields{
				DBHost:     "localhost:5432",
				DBName:     "postgres",
				DBUser:     "postgres",
				DBPassword: "postgres",
			},
			want: "postgres://postgres:postgres@localhost:5432/postgres",
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				cfg := Config{
					DBHost:             tt.fields.DBHost,
					DBName:             tt.fields.DBName,
					DBUser:             tt.fields.DBUser,
					DBPassword:         tt.fields.DBPassword,
					TablePrompt:        tt.fields.TablePrompt,
					TablePrediction:    tt.fields.TablePrediction,
					TableSuccessStatus: tt.fields.TableSuccessStatus,
					TableUsers:         tt.fields.TableUsers,
					TableTokens:        tt.fields.TableTokens,
					SSLMode:            tt.fields.SSLMode,
				}
				if got := cfg.ConnectionString(); got != tt.want {
					t.Errorf("ConnectionString() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestClient_GetDailySuccessfulResultsTimestampsByUserID(t *testing.T) {
	type fields struct {
		c                         dbClient
		tableWritePrompt          string
		tableWriteModelPrediction string
		tableWriteSuccessFlag     string
		tableUsers                string
		tableTokens               string
	}
	type args struct {
		ctx    context.Context
		userID string
	}
	tests := []struct {
		name                      string
		fields                    fields
		args                      args
		want                      []time.Time
		wantErr                   bool
		wantExecutedQueryTemplate string
	}{
		{
			name: "empty array",
			fields: fields{
				c: &mockDbClient{
					query: "",
					v: &mockRows{
						tag: pgconn.NewCommandTag("SELECT"),
						s:   &sync.RWMutex{},
						v:   nil,
					},
				},
				tableWriteSuccessFlag: "foo",
			},
			args: args{
				ctx:    context.TODO(),
				userID: "",
			},
			want:                      nil,
			wantErr:                   false,
			wantExecutedQueryTemplate: `SELECT timestamp FROM foo WHERE timestamp::date = current_date AND user_id = $1`,
		},
		{
			name: "array with two elements",
			fields: fields{
				c: &mockDbClient{
					query: "",
					v: &mockRows{
						tag: pgconn.NewCommandTag("SELECT"),
						s:   &sync.RWMutex{},
						v: [][]any{
							{time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)},
							{time.Date(2023, 1, 1, 0, 10, 0, 0, time.UTC)},
						},
					},
				},
				tableWriteSuccessFlag: "foo",
			},
			args: args{
				ctx:    context.TODO(),
				userID: "",
			},
			want: []time.Time{
				time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				time.Date(2023, 1, 1, 0, 10, 0, 0, time.UTC),
			},
			wantErr:                   false,
			wantExecutedQueryTemplate: `SELECT timestamp FROM foo WHERE timestamp::date = current_date AND user_id = $1`,
		},
		{
			name: "unhappy path",
			fields: fields{
				c: &mockDbClient{
					err: errors.New("foobat"),
				},
				tableWriteSuccessFlag: "foo",
			},
			args: args{
				ctx:    context.TODO(),
				userID: "",
			},
			wantErr:                   true,
			wantExecutedQueryTemplate: `SELECT timestamp FROM foo WHERE timestamp::date = current_date AND user_id = $1`,
		},
		{
			name: "unhappy path: while reading a raw",
			fields: fields{
				c: &mockDbClient{
					query: "",
					v: &mockRows{
						tag: pgconn.NewCommandTag("SELECT"),
						s:   &sync.RWMutex{},
						err: errors.New("foobar"),
						v: [][]any{
							{time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)},
						},
					},
				},
				tableWriteSuccessFlag: "foo",
			},
			args: args{
				ctx:    context.TODO(),
				userID: "",
			},
			wantErr:                   true,
			wantExecutedQueryTemplate: `SELECT timestamp FROM foo WHERE timestamp::date = current_date AND user_id = $1`,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c := Client{
					c:                     tt.fields.c,
					tableWriteSuccessFlag: tt.fields.tableWriteSuccessFlag,
				}
				got, err := c.GetDailySuccessfulResultsTimestampsByUserID(tt.args.ctx, tt.args.userID)
				if (err != nil) != tt.wantErr {
					t.Errorf("GetDailySuccessfulResultsTimestampsByUserID() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetDailySuccessfulResultsTimestampsByUserID() got = %v, want %v", got, tt.want)
				}
				gotQueryExecuted := c.c.(*mockDbClient).query
				if gotQueryExecuted != tt.wantExecutedQueryTemplate {
					t.Errorf(
						"GetDailySuccessfulResultsTimestampsByUserID() executes wrong query = %s, want = %s",
						gotQueryExecuted, tt.wantExecutedQueryTemplate,
					)
				}
			},
		)
	}
}

func TestClient_CreateUser(t *testing.T) {
	type fields struct {
		c                         dbClient
		tableWritePrompt          string
		tableWriteModelPrediction string
		tableWriteSuccessFlag     string
		tableUsers                string
		tableTokens               string
	}
	type args struct {
		ctx         context.Context
		id          string
		email       string
		fingerprint string
		isActive    bool
		role        *uint8
	}
	tests := []struct {
		name   string
		fields fields
		args   args

		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				c:                         &mockDbClient{},
				tableWritePrompt:          "foo",
				tableWriteModelPrediction: "bar",
				tableWriteSuccessFlag:     "baz",
				tableUsers:                "qux",
				tableTokens:               "quxx",
			},
			args: args{
				ctx:  context.TODO(),
				id:   "ccb42cbf-92c5-4069-bd01-ae25d49d9727",
				role: pointerUint(0),
			},
			wantErr: false,
		},
		{
			name: "shall fail: user ID is missing",
			fields: fields{
				c:                         &mockDbClient{},
				tableWritePrompt:          "foo",
				tableWriteModelPrediction: "bar",
				tableWriteSuccessFlag:     "baz",
				tableUsers:                "qux",
				tableTokens:               "quxx",
			},
			args: args{
				ctx: context.TODO(),
				id:  "",
			},
			wantErr: true,
		},
		{
			name: "shall fail: user role is missing",
			fields: fields{
				c:                         &mockDbClient{},
				tableWritePrompt:          "foo",
				tableWriteModelPrediction: "bar",
				tableWriteSuccessFlag:     "baz",
				tableUsers:                "qux",
				tableTokens:               "quxx",
			},
			args: args{
				ctx: context.TODO(),
				id:  "ccb42cbf-92c5-4069-bd01-ae25d49d9727",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c := Client{
					c:                         tt.fields.c,
					tableWritePrompt:          tt.fields.tableWritePrompt,
					tableWriteModelPrediction: tt.fields.tableWriteModelPrediction,
					tableWriteSuccessFlag:     tt.fields.tableWriteSuccessFlag,
					tableUsers:                tt.fields.tableUsers,
					tableTokens:               tt.fields.tableTokens,
				}
				if err := c.CreateUser(
					tt.args.ctx, tt.args.id, tt.args.email, tt.args.fingerprint, tt.args.isActive, tt.args.role,
				); (err != nil) != tt.wantErr {
					t.Errorf("CreateUser() error = %v, wantErr %v", err, tt.wantErr)
				}
			},
		)
	}
}

func pointerUint(i uint8) *uint8 {
	return &i
}

func TestClient_ReadUser(t *testing.T) {
	type fields struct {
		c                         dbClient
		tableWritePrompt          string
		tableWriteModelPrediction string
		tableWriteSuccessFlag     string
		tableUsers                string
		tableTokens               string
	}
	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		wantFound       bool
		wantIsActive    bool
		wantRole        uint8
		wantEmail       string
		wantFingerprint string
		wantErr         bool
	}{
		{
			name: "happy path: user found",
			fields: fields{
				c: &mockDbClient{
					query: "",
					v: &mockRows{
						tag: pgconn.NewCommandTag("SELECT"),
						s:   &sync.RWMutex{},
						v: [][]any{
							{true, "foo@bar.baz", 1, "qux"},
						},
					},
				},
				tableUsers: "users",
			},
			args: args{
				ctx: context.TODO(),
				id:  "ccb42cbf-92c5-4069-bd01-ae25d49d9727",
			},
			wantFound:       true,
			wantIsActive:    true,
			wantRole:        1,
			wantEmail:       "foo@bar.baz",
			wantFingerprint: "qux",
			wantErr:         false,
		},
		{
			name: "happy path: user not found",
			fields: fields{
				c: &mockDbClient{
					query: "",
					v: &mockRows{
						tag: pgconn.NewCommandTag("SELECT"),
						s:   &sync.RWMutex{},
					},
				},
				tableUsers: "users",
			},
			args: args{
				ctx: context.TODO(),
				id:  "ccb42cbf-92c5-4069-bd01-ae25d49d9727",
			},
			wantFound:       false,
			wantIsActive:    false,
			wantRole:        0,
			wantEmail:       "",
			wantFingerprint: "",
			wantErr:         false,
		},
		{
			name:    "empty user id provided",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c := Client{
					c:                         tt.fields.c,
					tableWritePrompt:          tt.fields.tableWritePrompt,
					tableWriteModelPrediction: tt.fields.tableWriteModelPrediction,
					tableWriteSuccessFlag:     tt.fields.tableWriteSuccessFlag,
					tableUsers:                tt.fields.tableUsers,
					tableTokens:               tt.fields.tableTokens,
				}
				gotFound, gotIsActive, gotRole, gotEmail, gotFingerprint, err := c.ReadUser(
					tt.args.ctx, tt.args.id,
				)
				if (err != nil) != tt.wantErr {
					t.Errorf("ReadUser() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if gotFound != tt.wantFound {
					t.Errorf("ReadUser() gotFound = %v, want %v", gotFound, tt.wantFound)
				}
				if gotIsActive != tt.wantIsActive {
					t.Errorf("ReadUser() gotIsActive = %v, want %v", gotIsActive, tt.wantIsActive)
				}
				if gotRole != tt.wantRole {
					t.Errorf("ReadUser() gotRole = %v, want %v", gotRole, tt.wantRole)
				}
				if gotEmail != tt.wantEmail {
					t.Errorf("ReadUser() gotEmail = %v, want %v", gotEmail, tt.wantEmail)
				}
				if gotFingerprint != tt.wantFingerprint {
					t.Errorf("ReadUser() gotFingerprint = %v, want %v", gotFingerprint, tt.wantFingerprint)
				}
			},
		)
	}
}

func TestClient_LookupUserByEmail(t *testing.T) {
	type fields struct {
		c                         dbClient
		tableWritePrompt          string
		tableWriteModelPrediction string
		tableWriteSuccessFlag     string
		tableUsers                string
		tableTokens               string
	}
	type args struct {
		ctx   context.Context
		email string
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantId       string
		wantIsActive bool
		wantErr      bool
		wantQuery    string
	}{
		{
			name: "happy path: user found",
			fields: fields{
				c: &mockDbClient{
					v: &mockRows{
						s:   &sync.RWMutex{},
						tag: pgconn.NewCommandTag("SELECT"),
						v: [][]any{
							{
								// user_id
								"ccb42cbf-92c5-4069-bd01-ae25d49d9727",
								// is_active
								true,
							},
							{
								// user_id
								"97375f24-91ed-4a70-adb6-4a5a4b19191c",
								// is_active
								false,
							},
						},
					},
				},
				tableUsers: "users",
			},
			args: args{
				ctx:   context.TODO(),
				email: "foo@bar.baz",
			},
			wantId:       "ccb42cbf-92c5-4069-bd01-ae25d49d9727",
			wantQuery:    "SELECT user_id, is_active FROM users WHERE email = $1 ORDER BY created_at LIMIT 1",
			wantIsActive: true,
			wantErr:      false,
		},
		{
			name: "happy path: user not found",
			fields: fields{
				c: &mockDbClient{
					v: &mockRows{
						s:   &sync.RWMutex{},
						tag: pgconn.NewCommandTag("SELECT"),
					},
				},
				tableUsers: "users",
			},
			args: args{
				ctx:   context.TODO(),
				email: "foo@bar.baz",
			},
			wantId:       "",
			wantQuery:    "SELECT user_id, is_active FROM users WHERE email = $1 ORDER BY created_at LIMIT 1",
			wantIsActive: false,
			wantErr:      false,
		},
		{
			name: "unhappy path: email not provided",
			args: args{
				ctx:   context.TODO(),
				email: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c := Client{
					c:                         tt.fields.c,
					tableWritePrompt:          tt.fields.tableWritePrompt,
					tableWriteModelPrediction: tt.fields.tableWriteModelPrediction,
					tableWriteSuccessFlag:     tt.fields.tableWriteSuccessFlag,
					tableUsers:                tt.fields.tableUsers,
					tableTokens:               tt.fields.tableTokens,
				}
				gotId, gotIsActive, err := c.LookupUserByEmail(tt.args.ctx, tt.args.email)
				if (err != nil) != tt.wantErr {
					t.Errorf("LookupUserByEmail() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if gotId != tt.wantId {
					t.Errorf("LookupUserByEmail() gotId = %v, want %v", gotId, tt.wantId)
				}
				if gotIsActive != tt.wantIsActive {
					t.Errorf("LookupUserByEmail() gotIsActive = %v, want %v", gotIsActive, tt.wantIsActive)
				}
				if err == nil && c.c.(*mockDbClient).query != tt.wantQuery {
					t.Error("LookupUserByEmail() executed unexpected query")
				}
			},
		)
	}
}

func TestClient_LookupUserByFingerprint(t *testing.T) {
	type fields struct {
		c                         dbClient
		tableWritePrompt          string
		tableWriteModelPrediction string
		tableWriteSuccessFlag     string
		tableUsers                string
		tableTokens               string
	}
	type args struct {
		ctx         context.Context
		fingerprint string
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantId       string
		wantIsActive bool
		wantErr      bool
		wantQuery    string
	}{
		{
			name: "happy path: user found",
			fields: fields{
				c: &mockDbClient{
					v: &mockRows{
						s:   &sync.RWMutex{},
						tag: pgconn.NewCommandTag("SELECT"),
						v: [][]any{
							{
								// user_id
								"ccb42cbf-92c5-4069-bd01-ae25d49d9727",
								// is_active
								true,
							},
							{
								// user_id
								"97375f24-91ed-4a70-adb6-4a5a4b19191c",
								// is_active
								false,
							},
						},
					},
				},
				tableUsers: "users",
			},
			args: args{
				ctx:         context.TODO(),
				fingerprint: "foo",
			},
			wantQuery:    "SELECT user_id, is_active FROM users WHERE web_fingerprint = $1 ORDER BY created_at LIMIT 1",
			wantId:       "ccb42cbf-92c5-4069-bd01-ae25d49d9727",
			wantIsActive: true,
			wantErr:      false,
		},
		{
			name: "happy path: user not found",
			fields: fields{
				c: &mockDbClient{
					v: &mockRows{
						s:   &sync.RWMutex{},
						tag: pgconn.NewCommandTag("SELECT"),
					},
				},
				tableUsers: "users",
			},
			args: args{
				ctx:         context.TODO(),
				fingerprint: "foo",
			},
			wantQuery:    "SELECT user_id, is_active FROM users WHERE web_fingerprint = $1 ORDER BY created_at LIMIT 1",
			wantId:       "",
			wantIsActive: false,
			wantErr:      false,
		},
		{
			name: "unhappy path: fingerprint not provided",
			args: args{
				ctx:         context.TODO(),
				fingerprint: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c := Client{
					c:                         tt.fields.c,
					tableWritePrompt:          tt.fields.tableWritePrompt,
					tableWriteModelPrediction: tt.fields.tableWriteModelPrediction,
					tableWriteSuccessFlag:     tt.fields.tableWriteSuccessFlag,
					tableUsers:                tt.fields.tableUsers,
					tableTokens:               tt.fields.tableTokens,
				}
				gotId, gotIsActive, err := c.LookupUserByFingerprint(tt.args.ctx, tt.args.fingerprint)
				if (err != nil) != tt.wantErr {
					t.Errorf("LookupUserByFingerprint() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if gotId != tt.wantId {
					t.Errorf("LookupUserByFingerprint() gotId = %v, want %v", gotId, tt.wantId)
				}
				if gotIsActive != tt.wantIsActive {
					t.Errorf("LookupUserByFingerprint() gotIsActive = %v, want %v", gotIsActive, tt.wantIsActive)
				}
				if err == nil && c.c.(*mockDbClient).query != tt.wantQuery {
					t.Error("LookupUserByFingerprint() executed unexpected query")
				}
			},
		)
	}
}

func TestClient_UpdateUserSetActive(t *testing.T) {
	type fields struct {
		c                         dbClient
		tableWritePrompt          string
		tableWriteModelPrediction string
		tableWriteSuccessFlag     string
		tableUsers                string
		tableTokens               string
	}
	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				c:          &mockDbClient{},
				tableUsers: "users",
			},
			args: args{
				ctx: context.TODO(),
				id:  "ccb42cbf-92c5-4069-bd01-ae25d49d9727",
			},
			wantErr: false,
		},
		{
			name: "unhappy path: no user id provided",
			fields: fields{
				c:          &mockDbClient{},
				tableUsers: "users",
			},
			args: args{
				ctx: context.TODO(),
				id:  "",
			},
			wantErr: true,
		},
		{
			name: "unhappy path: db error",
			fields: fields{
				c: &mockDbClient{
					err: errors.New("foobar"),
				},
				tableUsers: "users",
			},
			args: args{
				ctx: context.TODO(),
				id:  "ccb42cbf-92c5-4069-bd01-ae25d49d9727",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c := Client{
					c:                         tt.fields.c,
					tableWritePrompt:          tt.fields.tableWritePrompt,
					tableWriteModelPrediction: tt.fields.tableWriteModelPrediction,
					tableWriteSuccessFlag:     tt.fields.tableWriteSuccessFlag,
					tableUsers:                tt.fields.tableUsers,
					tableTokens:               tt.fields.tableTokens,
				}
				if err := c.UpdateUserSetActive(tt.args.ctx, tt.args.id); (err != nil) != tt.wantErr {
					t.Errorf("UpdateUserSetActive() error = %v, wantErr %v", err, tt.wantErr)
				}
			},
		)
	}
}

func TestClient_WriteOneTimeSecret(t *testing.T) {
	type fields struct {
		c                         dbClient
		tableWritePrompt          string
		tableWriteModelPrediction string
		tableWriteSuccessFlag     string
		tableUsers                string
		tableTokens               string
		tableOneTimeSecret        string
	}
	type args struct {
		ctx       context.Context
		userID    string
		secret    string
		createdAt time.Time
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantErr   bool
		wantQuery string
	}{
		{
			name: "happy path",
			fields: fields{
				c:                  &mockDbClient{},
				tableOneTimeSecret: "secret",
			},
			args: args{
				ctx:       context.TODO(),
				userID:    "ccb42cbf-92c5-4069-bd01-ae25d49d9727",
				secret:    "123456",
				createdAt: time.Now().UTC().Add(-1 * time.Minute),
			},
			wantErr:   false,
			wantQuery: "INSERT INTO secret (user_id, secret, created_at) VALUES ($1, $2, $3) ON CONFLICT DO UPDATE SET user_id = $1, secret = $2, created_at = $3",
		},
		{
			name: "unhappy path: no user id provided",
			fields: fields{
				c:                  &mockDbClient{},
				tableOneTimeSecret: "secret",
			},
			args: args{
				ctx:       context.TODO(),
				secret:    "123456",
				createdAt: time.Now().UTC().Add(-1 * time.Minute),
			},
			wantErr: true,
		},
		{
			name: "unhappy path: no secret provided",
			fields: fields{
				c:                  &mockDbClient{},
				tableOneTimeSecret: "secret",
			},
			args: args{
				ctx:       context.TODO(),
				userID:    "ccb42cbf-92c5-4069-bd01-ae25d49d9727",
				createdAt: time.Now().UTC().Add(-1 * time.Minute),
			},
			wantErr: true,
		},
		{
			name: "unhappy path: no createdAt provided",
			fields: fields{
				c:                  &mockDbClient{},
				tableOneTimeSecret: "secret",
			},
			args: args{
				ctx:    context.TODO(),
				userID: "ccb42cbf-92c5-4069-bd01-ae25d49d9727",
				secret: "123456",
			},
			wantErr: true,
		},
	}
	t.Parallel()
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c := Client{
					c:                  tt.fields.c,
					tableOneTimeSecret: tt.fields.tableOneTimeSecret,
				}
				err := c.WriteOneTimeSecret(
					tt.args.ctx, tt.args.userID, tt.args.secret, tt.args.createdAt,
				)
				if (err != nil) != tt.wantErr {
					t.Errorf("WriteOneTimeSecret() error = %v, wantErr %v", err, tt.wantErr)
				}
				if err == nil && tt.wantQuery != "" && c.c.(*mockDbClient).query != tt.wantQuery {
					t.Error("WriteOneTimeSecret() executed unexpected query")
				}
			},
		)
	}
}

func TestClient_ReadOneTimeSecret(t *testing.T) {
	type fields struct {
		c                         dbClient
		tableWritePrompt          string
		tableWriteModelPrediction string
		tableWriteSuccessFlag     string
		tableUsers                string
		tableTokens               string
		tableOneTimeSecret        string
	}
	type args struct {
		ctx    context.Context
		userID string
	}
	iat := time.Now().UTC().Add(-1 * time.Minute)
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantFound    bool
		wantSecret   string
		wantIssuedAt time.Time
		wantErr      bool
		wantQuery    string
	}{
		{
			name: "happy path",
			fields: fields{
				c: &mockDbClient{
					v: &mockRows{
						tag: pgconn.NewCommandTag("SELECT"),
						s:   &sync.RWMutex{},
						v:   [][]any{{"123456", iat}},
					},
				},
				tableOneTimeSecret: "secret",
			},
			args: args{
				ctx:    context.TODO(),
				userID: "ccb42cbf-92c5-4069-bd01-ae25d49d9727",
			},
			wantFound:    true,
			wantSecret:   "123456",
			wantIssuedAt: iat,
			wantErr:      false,
			wantQuery:    "SELECT secret, created_at FROM secret WHERE user_id = $1",
		},
		{
			name:    "unhappy path: no user ID provided",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c := Client{
					c:                         tt.fields.c,
					tableWritePrompt:          tt.fields.tableWritePrompt,
					tableWriteModelPrediction: tt.fields.tableWriteModelPrediction,
					tableWriteSuccessFlag:     tt.fields.tableWriteSuccessFlag,
					tableUsers:                tt.fields.tableUsers,
					tableTokens:               tt.fields.tableTokens,
					tableOneTimeSecret:        tt.fields.tableOneTimeSecret,
				}
				gotFound, gotSecret, gotIssuedAt, err := c.ReadOneTimeSecret(tt.args.ctx, tt.args.userID)
				if (err != nil) != tt.wantErr {
					t.Errorf("ReadOneTimeSecret() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if gotFound != tt.wantFound {
					t.Errorf("ReadOneTimeSecret() gotFound = %v, want %v", gotFound, tt.wantFound)
				}
				if gotSecret != tt.wantSecret {
					t.Errorf("ReadOneTimeSecret() gotSecret = %v, want %v", gotSecret, tt.wantSecret)
				}
				if !reflect.DeepEqual(gotIssuedAt, tt.wantIssuedAt) {
					t.Errorf("ReadOneTimeSecret() gotIssuedAt = %v, want %v", gotIssuedAt, tt.wantIssuedAt)
				}
				if err == nil && tt.wantQuery != "" && c.c.(*mockDbClient).query != tt.wantQuery {
					t.Error("WriteOneTimeSecret() executed unexpected query")
				}
			},
		)
	}
}

func TestClient_DeleteOneTimeSecret(t *testing.T) {
	type fields struct {
		c                         dbClient
		tableWritePrompt          string
		tableWriteModelPrediction string
		tableWriteSuccessFlag     string
		tableUsers                string
		tableTokens               string
		tableOneTimeSecret        string
	}
	type args struct {
		ctx    context.Context
		userID string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantErr   bool
		wantQuery string
	}{
		{
			name: "happy path",
			fields: fields{
				c:                  &mockDbClient{},
				tableOneTimeSecret: "secret",
			},
			args: args{
				ctx:    context.TODO(),
				userID: "ccb42cbf-92c5-4069-bd01-ae25d49d9727",
			},
			wantErr:   false,
			wantQuery: "DELETE FROM secret WHERE user_id = $1",
		},
		{
			name: "unhappy path: no user ID provided",
			fields: fields{
				c:                  &mockDbClient{},
				tableOneTimeSecret: "secret",
			},
			args: args{
				ctx:    context.TODO(),
				userID: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c := Client{
					c:                         tt.fields.c,
					tableWritePrompt:          tt.fields.tableWritePrompt,
					tableWriteModelPrediction: tt.fields.tableWriteModelPrediction,
					tableWriteSuccessFlag:     tt.fields.tableWriteSuccessFlag,
					tableUsers:                tt.fields.tableUsers,
					tableTokens:               tt.fields.tableTokens,
					tableOneTimeSecret:        tt.fields.tableOneTimeSecret,
				}
				err := c.DeleteOneTimeSecret(tt.args.ctx, tt.args.userID)
				if (err != nil) != tt.wantErr {
					t.Errorf("DeleteOneTimeSecret() error = %v, wantErr %v", err, tt.wantErr)
				}
				if err == nil && tt.wantQuery != "" && c.c.(*mockDbClient).query != tt.wantQuery {
					t.Error("WriteOneTimeSecret() executed unexpected query")
				}
			},
		)
	}
}
