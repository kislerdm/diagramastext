package postgres

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	type fields struct {
		DBHost          string
		DBName          string
		DBUser          string
		DBPassword      string
		TablePrompt     string
		TablePrediction string
		SSLMode         string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr error
	}{
		{
			name: "valid",
			fields: fields{
				DBHost:          "localhost",
				DBName:          "postgres",
				DBUser:          "postgres",
				DBPassword:      "postgres",
				TablePrompt:     "foo",
				TablePrediction: "bar",
			},
			wantErr: nil,
		},
		{
			name: "valid: ssl - full verification",
			fields: fields{
				DBHost:          "localhost",
				DBName:          "postgres",
				DBUser:          "postgres",
				DBPassword:      "postgres",
				TablePrompt:     "foo",
				TablePrediction: "bar",
				SSLMode:         "verify-full",
			},
			wantErr: nil,
		},
		{
			name: "invalid: host is missing",
			fields: fields{
				DBHost:          "",
				DBName:          "postgres",
				DBUser:          "postgres",
				DBPassword:      "postgres",
				TablePrompt:     "foo",
				TablePrediction: "bar",
			},
			wantErr: errors.New("host must be provided"),
		},
		{
			name: "invalid: dbname is missing",
			fields: fields{
				DBHost:          "localhost",
				DBName:          "",
				DBUser:          "postgres",
				DBPassword:      "postgres",
				TablePrompt:     "foo",
				TablePrediction: "bar",
			},
			wantErr: errors.New("dbname must be provided"),
		},
		{
			name: "invalid: user is missing",
			fields: fields{
				DBHost:          "localhost",
				DBName:          "postgres",
				DBUser:          "",
				DBPassword:      "postgres",
				TablePrompt:     "foo",
				TablePrediction: "bar",
			},
			wantErr: errors.New("user must be provided"),
		},
		{
			name: "invalid: table_prompt is missing",
			fields: fields{
				DBHost:          "localhost",
				DBName:          "postgres",
				DBUser:          "postgres",
				DBPassword:      "postgres",
				TablePrompt:     "",
				TablePrediction: "bar",
			},
			wantErr: errors.New("table_prompt must be provided"),
		},
		{
			name: "invalid: table_prediction is missing",
			fields: fields{
				DBHost:          "localhost",
				DBName:          "postgres",
				DBUser:          "postgres",
				DBPassword:      "postgres",
				TablePrompt:     "foo",
				TablePrediction: "",
			},
			wantErr: errors.New("table_prediction must be provided"),
		},
		{
			name: "invalid: ssl mode is wrong",
			fields: fields{
				DBHost:          "localhost",
				DBName:          "postgres",
				DBUser:          "postgres",
				DBPassword:      "postgres",
				TablePrompt:     "foo",
				TablePrediction: "bar",
				SSLMode:         "qux",
			},
			wantErr: errors.New("ssl mode qux is not supported"),
		},
	}

	t.Parallel()

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				cfg := Config{
					DBHost:          tt.fields.DBHost,
					DBName:          tt.fields.DBName,
					DBUser:          tt.fields.DBUser,
					DBPassword:      tt.fields.DBPassword,
					TablePrompt:     tt.fields.TablePrompt,
					TablePrediction: tt.fields.TablePrediction,
					SSLMode:         tt.fields.SSLMode,
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
					DBHost:          "mock",
					DBName:          "postgres",
					DBUser:          "postgres",
					DBPassword:      "foo",
					TablePrompt:     "bar",
					TablePrediction: "baz",
				},
			},
			want: &Client{
				c:                         mockDbClient{},
				tableWritePrompt:          "bar",
				tableWriteModelPrediction: "baz",
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

func Test_client_WritePrompt(t *testing.T) {
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
			fields: fields{mockDbClient{}},
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
			fields: fields{mockDbClient{}},
			args: args{
				ctx:    context.TODO(),
				prompt: "c4 diagram of four boxes",
				userID: "c40bad11-0822-4d84-9f61-44b9a97b0432",
			},
			wantErr: errors.New("request_id is required"),
		},
		{
			name:   "unhappy path: no prompt",
			fields: fields{mockDbClient{}},
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
				mockDbClient{
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

func Test_client_WriteModelResult(t *testing.T) {
	type fields struct {
		c dbClient
	}
	type args struct {
		ctx                           context.Context
		requestID, userID, prediction string
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
				c: mockDbClient{},
			},
			args: args{
				ctx:        context.TODO(),
				requestID:  "693a35ba-e42c-4168-8afc-5a7c359d1d05",
				userID:     "c40bad11-0822-4d84-9f61-44b9a97b0432",
				prediction: `{"nodes":[{"id":"0"},{"id":"1"}]}`,
			},
			wantErr: nil,
		},
		{
			name: "unhappy path: no request id",
			fields: fields{
				c: mockDbClient{},
			},
			args: args{
				ctx:        context.TODO(),
				userID:     "c40bad11-0822-4d84-9f61-44b9a97b0432",
				prediction: `{"nodes":[{"id":"0"},{"id":"1"}]}`,
			},
			wantErr: errors.New("request_id is required"),
		},
		{
			name: "unhappy path: no response",
			fields: fields{
				c: mockDbClient{},
			},
			args: args{
				ctx:       context.TODO(),
				requestID: "693a35ba-e42c-4168-8afc-5a7c359d1d05",
				userID:    "c40bad11-0822-4d84-9f61-44b9a97b0432",
			},
			wantErr: errors.New("response is required"),
		},
		{
			name: "unhappy path: no relation found",
			fields: fields{
				c: mockDbClient{
					err: errors.New(`pq: relation "bar" does not exist`),
				},
			},
			args: args{
				ctx:        context.TODO(),
				requestID:  "693a35ba-e42c-4168-8afc-5a7c359d1d05",
				userID:     "c40bad11-0822-4d84-9f61-44b9a97b0432",
				prediction: `{"nodes":[{"id":"0"},{"id":"1"}]}`,
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
				err := c.WriteModelResult(tt.args.ctx, tt.args.requestID, tt.args.userID, tt.args.prediction)
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
				c: mockDbClient{},
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
				c: mockDbClient{err: errors.New("error disconnecting")},
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

func Test_host(t *testing.T) {
	t.Parallel()

	t.Run(
		"with port", func(t *testing.T) {
			// GIVEN
			hosts := []string{
				"localhost",
				"127.0.0.1",
			}
			ports := []string{"5432", "15342", "1213"}

			for _, hostStr := range hosts {
				for _, port := range ports {
					// WHEN
					hostInput := hostStr + ":" + port
					gotHostPsqlStr := host(hostInput)

					// THEN
					wantPsqlStr := " host=" + hostStr + " port=" + port
					if gotHostPsqlStr != wantPsqlStr {
						t.Errorf("unexpected result for input host " + hostStr + ":" + port)
						return
					}
				}
			}
		},
	)

	t.Run(
		"without port", func(t *testing.T) {
			// GIVEN
			hosts := []string{
				"ep-fragrant-mouse-914820.us-east-2.aws.neon.tech",
			}

			// WHEN
			for _, hostStr := range hosts {
				// WHEN
				hostInput := hostStr
				gotHostPsqlStr := host(hostInput)

				// THEN
				wantPsqlStr := " host=" + hostStr
				if gotHostPsqlStr != wantPsqlStr {
					t.Errorf("unexpected result for input host " + hostStr)
					return
				}
			}
		},
	)
}
