package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/kislerdm/diagramastext/core/storage"
	_ "github.com/lib/pq"
)

type client struct {
	c *sql.DB
}

func (c client) WritePrompt(ctx context.Context, v storage.UserInput) error {
	const table = "user_prompt"
	_, err := c.c.QueryContext(
		ctx, "INSERT INTO "+table+" (request_id, user_id, prompt) VALUES ('?', '?', '?')", v.RequestID,
		v.UserID,
		v.Prompt,
	)
	return err
}

func (c client) WriteModelPrediction(ctx context.Context, v storage.ModelOutput) error {
	const table = "openai_response"
	_, err := c.c.QueryContext(
		ctx, "INSERT INTO "+table+" (request_id, user_id, respose) VALUES ('?', '?', '?')", v.RequestID,
		v.UserID,
		v.Response,
	)
	return err
}

// NewClient initiates the postgres client.
func NewClient(host, dbname, user, password string) (storage.Client, error) {
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

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	return client{c: db}, nil
}
