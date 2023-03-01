package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/kislerdm/diagramastext/server/errors"

	_ "github.com/lib/pq"
)

type pgClient struct {
	c dbClient
}

func (c pgClient) Close(ctx context.Context) error {
	return c.c.Close()
}

const tableWritePrompt = "user_prompt"

func (c pgClient) WritePrompt(ctx context.Context, v UserInput) error {
	if v.UserID == "" {
		return errors.Error{
			Service: errors.ServiceStorage,
			Stage:   errors.StageSerialization,
			Message: "user_id is required",
		}
	}
	if v.RequestID == "" {
		return errors.Error{
			Service: errors.ServiceStorage,
			Stage:   errors.StageSerialization,
			Message: "request_id is required",
		}
	}
	if v.Prompt == "" {
		return errors.Error{
			Service: errors.ServiceStorage,
			Stage:   errors.StageSerialization,
			Message: "prompt is required",
		}
	}
	if v.Timestamp.Second() <= 0 {
		v.Timestamp = time.Now().UTC()
	}

	if _, err := c.c.ExecContext(
		ctx, `INSERT INTO `+tableWritePrompt+
			` (request_id, user_id, prompt, timestamp) VALUES ($1, $2, $3, $4)`,
		v.RequestID,
		v.UserID,
		v.Prompt,
		v.Timestamp,
	); err != nil {
		return errors.Error{
			Service: errors.ServiceStorage,
			Stage:   errors.StageSerialization,
			Message: err.Error(),
		}
	}

	return nil
}

const tableWriteModelPrediction = "openai_response"

func (c pgClient) WriteModelPrediction(ctx context.Context, v ModelOutput) error {
	if v.UserID == "" {
		return errors.Error{
			Service: errors.ServiceStorage,
			Stage:   errors.StageSerialization,
			Message: "user_id is required",
		}
	}
	if v.RequestID == "" {
		return errors.Error{
			Service: errors.ServiceStorage,
			Stage:   errors.StageSerialization,
			Message: "request_id is required",
		}
	}
	if v.Response == "" {
		return errors.Error{
			Service: errors.ServiceStorage,
			Stage:   errors.StageSerialization,
			Message: "response is required",
		}
	}
	if v.Timestamp.Second() <= 0 {
		v.Timestamp = time.Now().UTC()
	}
	if _, err := c.c.ExecContext(
		ctx, `INSERT INTO `+tableWriteModelPrediction+
			` (request_id, user_id, response, timestamp) VALUES ($1, $2, $3, $4)`,
		v.RequestID,
		v.UserID,
		v.Response,
		v.Timestamp,
	); err != nil {
		return errors.Error{
			Service: errors.ServiceStorage,
			Stage:   errors.StageSerialization,
			Message: err.Error(),
		}
	}
	
	return nil
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

// NewPgClient initiates the postgres pgClient.
func NewPgClient(ctx context.Context, host, dbname, user, password string) (Client, error) {
	if host == "" {
		return nil, errors.Error{
			Service: errors.ServiceStorage,
			Stage:   errors.StageInit,
			Message: "host must be provided",
		}
	}

	if dbname == "" {
		return nil, errors.Error{
			Service: errors.ServiceStorage,
			Stage:   errors.StageInit,
			Message: "dbname must be provided",
		}
	}

	if user == "" {
		return nil, errors.Error{
			Service: errors.ServiceStorage,
			Stage:   errors.StageInit,
			Message: "user must be provided",
		}
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
		return nil, errors.Error{
			Service: errors.ServiceStorage,
			Stage:   errors.StageInit,
			Message: err.Error(),
		}
	}

	if host == "mock" {
		db = mockDbClient{}
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, errors.Error{
			Service: errors.ServiceStorage,
			Stage:   errors.StageConnection,
			Message: err.Error(),
		}
	}

	return pgClient{c: db}, nil
}

type dbClient interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	PingContext(ctx context.Context) error
	Close() error
}
