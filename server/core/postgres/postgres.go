package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/kislerdm/diagramastext/server/core/contract"
	_ "github.com/lib/pq"
)

// Config configuration of the postgres client.
type Config struct {
	DBHost          string `json:"db_host"`
	DBName          string `json:"db_name"`
	DBUser          string `json:"db_user"`
	DBPassword      string `json:"db_password"`
	TablePrompt     string `json:"table_prompt,omitempty"`
	TablePrediction string `json:"table_prediction,omitempty"`
}

func (cfg Config) Validate() error {
	if cfg.DBHost == "" {
		return errors.New("host must be provided")
	}
	if cfg.DBName == "" {
		return errors.New("dbname must be provided")
	}
	if cfg.DBUser == "" {
		return errors.New("user must be provided")
	}
	if cfg.TablePrompt == "" {
		return errors.New("table to store prompt must be provided")
	}
	if cfg.TablePrediction == "" {
		return errors.New("table to store prediction must be provided")
	}
	return nil
}

// NewClient initiates the postgres pgClient.
func NewClient(ctx context.Context, cfg Config) (
	contract.ClientStorage, error,
) {
	if cfg.DBHost == "mock" {
		return pgClient{
			c:                         mockDbClient{},
			tableWritePrompt:          cfg.TablePrompt,
			tableWriteModelPrediction: cfg.TablePrediction,
		}, nil
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	connStr := "user=" + cfg.DBUser +
		" dbname=" + cfg.DBName +
		" host=" + cfg.DBHost +
		" sslmode=verify-full"

	if cfg.DBPassword != "" {
		connStr += " password=" + cfg.DBPassword
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return &pgClient{
		c:                         db,
		tableWritePrompt:          cfg.TablePrompt,
		tableWriteModelPrediction: cfg.TablePrediction,
	}, nil
}

type pgClient struct {
	c                         dbClient
	tableWritePrompt          string
	tableWriteModelPrediction string
}

func (c pgClient) Close(ctx context.Context) error {
	return c.c.Close()
}

func (c pgClient) WritePrompt(ctx context.Context, requestID, prompt, userID string) error {
	if requestID == "" {
		return errors.New("request_id is required")
	}
	if prompt == "" {
		return errors.New("prompt is required")
	}
	_, err := c.c.ExecContext(
		ctx, `INSERT INTO `+c.tableWritePrompt+
			` (request_id, user_id, prompt, timestamp) VALUES ($1, $2, $3, $4)`,
		requestID,
		userID,
		prompt,
		time.Now().UTC(),
	)
	return err
}

func (c pgClient) WriteModelPrediction(ctx context.Context, requestID, result, userID string) error {
	if requestID == "" {
		return errors.New("request_id is required")
	}
	if result == "" {
		return errors.New("response is required")
	}
	_, err := c.c.ExecContext(
		ctx, `INSERT INTO `+c.tableWriteModelPrediction+
			` (request_id, user_id, response, timestamp) VALUES ($1, $2, $3, $4)`,
		requestID,
		userID,
		result,
		time.Now().UTC(),
	)
	return err
}

type mockDbClient struct {
	err error
}

func (m mockDbClient) Close() error {
	return m.err
}

func (m mockDbClient) PingContext(_ context.Context) error {
	return m.err
}

func (m mockDbClient) ExecContext(_ context.Context, _ string, _ ...any) (sql.Result, error) {
	return nil, m.err
}

type dbClient interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	PingContext(ctx context.Context) error
	Close() error
}
