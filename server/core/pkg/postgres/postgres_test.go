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
				TableSuccessStatus: "",
			},
			wantErr: errors.New("table_success_status must be provided"),
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
				},
			},
			want: &Client{
				c:                         &mockDbClient{},
				tableWritePrompt:          "bar",
				tableWriteModelPrediction: "baz",
				tableWriteSuccessFlag:     "qux",
				tableUsers:                "quxx",
				tableTokens:               "baz",
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
