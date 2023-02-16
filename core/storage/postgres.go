package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/kislerdm/diagramastext/core"
	_ "github.com/lib/pq"
)

type client struct {
	c dbClient
}

func (c client) Close(ctx context.Context) error {
	return c.c.Close()
}

const tableWritePrompt = "user_prompt"

func (c client) WritePrompt(ctx context.Context, v core.UserInput) error {
	if v.UserID == "" {
		return errors.New("user_id is required")
	}
	if v.RequestID == "" {
		return errors.New("request_id is required")
	}
	if v.Prompt == "" {
		return errors.New("prompt is required")
	}
	if v.Timestamp.Second() <= 0 {
		v.Timestamp = time.Now().UTC()
	}

	_, err := c.c.ExecContext(
		ctx, `INSERT INTO `+tableWritePrompt+
			` (request_id, user_id, prompt, timestamp) VALUES ($1, $2, $3, $4)`,
		v.RequestID,
		v.UserID,
		v.Prompt,
		v.Timestamp,
	)
	return err
}

const tableWriteModelPrediction = "openai_response"

func (c client) WriteModelPrediction(ctx context.Context, v core.ModelOutput) error {
	if v.UserID == "" {
		return errors.New("user_id is required")
	}
	if v.RequestID == "" {
		return errors.New("request_id is required")
	}
	if v.Response == "" {
		return errors.New("response is required")
	}
	if v.Timestamp.Second() <= 0 {
		v.Timestamp = time.Now().UTC()
	}
	_, err := c.c.ExecContext(
		ctx, `INSERT INTO `+tableWriteModelPrediction+
			` (request_id, user_id, response, timestamp) VALUES ($1, $2, $3, $4)`,
		v.RequestID,
		v.UserID,
		v.Response,
		v.Timestamp,
	)
	return err
}

type mockDbClient struct {
	err error
}

func (m mockDbClient) Close() error {
	return m.err
}

func (m mockDbClient) PingContext(ctx context.Context) error {
	return m.err
}

func (m mockDbClient) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return nil, m.err
}

// NewClient initiates the postgres client.
func NewClient(ctx context.Context, host, dbname, user, password string) (core.ClientStorage, error) {
	if host == "" {
		return nil, errors.New("host must be provided")
	}

	if dbname == "" {
		return nil, errors.New("dbname must be provided")
	}

	if user == "" {
		return nil, errors.New("user must be provided")
	}

	connStr := "user=" + user +
		" dbname=" + dbname +
		" host=" + host +
		" sslmode=verify-full"

	if password != "" {
		connStr += " password=" + password
	}

	var db dbClient

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if host == "mock" {
		db = mockDbClient{}
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return client{c: db}, nil
}

type dbClient interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	PingContext(ctx context.Context) error
	Close() error
}
