package postgres

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Config configuration of the postgres Client.
type Config struct {
	DBHost             string `json:"db_host"`
	DBName             string `json:"db_name"`
	DBUser             string `json:"db_user"`
	DBPassword         string `json:"db_password"`
	TablePrompt        string `json:"table_prompt,omitempty"`
	TablePrediction    string `json:"table_prediction,omitempty"`
	TableSuccessStatus string `json:"table_success_status,omitempty"`
	TableUsers         string `json:"table_users,omitempty"`
	TableTokens        string `json:"table_tokens,omitempty"`
	SSLMode            string `json:"ssl_mode"`
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
		return errors.New("table_prompt must be provided")
	}
	if cfg.TablePrediction == "" {
		return errors.New("table_prediction must be provided")
	}
	if cfg.TableSuccessStatus == "" {
		return errors.New("table_success_status must be provided")
	}
	if cfg.TableUsers == "" {
		return errors.New("table_users must be provided")
	}
	if cfg.TableTokens == "" {
		return errors.New("table_tokens must be provided")
	}
	return validateSSLMode(cfg.SSLMode)
}

func (cfg Config) ConnectionString() string {
	writeStrings := func(buf *strings.Builder, s ...string) {
		for _, el := range s {
			_, _ = buf.WriteString(el)
		}
	}
	var buf strings.Builder
	writeStrings(&buf, "postgres://", cfg.DBUser, ":", cfg.DBPassword, "@", cfg.DBHost, "/", cfg.DBName)
	if cfg.SSLMode != "" {
		writeStrings(&buf, "?sslmode=", cfg.SSLMode)
	}
	return buf.String()
}

func validateSSLMode(mode string) error {
	switch mode {
	case "verify-full", "disable", "":
		return nil
	default:
		return errors.New("ssl mode " + mode + " is not supported")
	}
}

// NewPostgresClient initiates the postgres Client.
func NewPostgresClient(ctx context.Context, cfg Config) (
	*Client, error,
) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	var db dbClient

	switch cfg.DBHost == "mock" {
	case true:
		db = &mockDbClient{}
	default:
		var err error
		db, err = pgx.Connect(ctx, cfg.ConnectionString())
		if err != nil {
			return nil, err
		}
	}

	return &Client{
		c:                         db,
		tableWritePrompt:          cfg.TablePrompt,
		tableWriteModelPrediction: cfg.TablePrediction,
		tableWriteSuccessFlag:     cfg.TableSuccessStatus,
		tableUsers:                cfg.TableUsers,
		tableTokens:               cfg.TableTokens,
	}, nil
}

type Client struct {
	c                         dbClient
	tableWritePrompt          string
	tableWriteModelPrediction string
	tableWriteSuccessFlag     string
	tableUsers                string
	tableTokens               string
}

func (c Client) GetActiveUserIDByActiveTokenID(ctx context.Context, id string) (string, error) {
	rows, err := c.c.Query(
		ctx, `SELECT u.user_id 
FROM `+c.tableUsers+` AS u 
INNER JOIN `+c.tableTokens+` AS t USING (user_id) 
WHERE t.token = $1 AND t.is_active AND u.is_active`, id,
	)
	if err != nil {
		return "", err
	}

	var userID string
	defer rows.Close()
	if rows.Next() {
		_ = rows.Scan(&userID)
	}
	return userID, nil
}

func (c Client) Close(ctx context.Context) error {
	return c.c.Close(ctx)
}

func (c Client) WriteInputPrompt(ctx context.Context, requestID, userID, prompt string) error {
	if requestID == "" {
		return errors.New("request_id is required")
	}
	if prompt == "" {
		return errors.New("prompt is required")
	}
	_, err := c.c.Exec(
		ctx, `INSERT INTO `+c.tableWritePrompt+
			` (request_id, user_id, prompt, timestamp) VALUES ($1, $2, $3, $4)`,
		requestID,
		userID,
		prompt,
		time.Now().UTC(),
	)
	return err
}

func (c Client) WriteModelResult(
	ctx context.Context, requestID, userID, predictionRaw, prediction, model string,
	usageTokensPrompt, usageTokensCompletions uint16,
) error {
	if requestID == "" {
		return errors.New("request_id is required")
	}
	if userID == "" {
		return errors.New("user_id is required")
	}
	if predictionRaw == "" {
		return errors.New("raw response is required")
	}
	if prediction == "" {
		return errors.New("response is required")
	}
	if model == "" {
		return errors.New("model is required")
	}
	_, err := c.c.Exec(
		ctx, `INSERT INTO `+c.tableWriteModelPrediction+
			` (
	 request_id
   , user_id
   , response
   , timestamp
   , model_id
   , prompt_tokens
   , completion_tokens
   , response_raw
) VALUES (
		  $1
		, $2
		, $3
		, $4
		, $5
		, $6
		, $7
		, $8
)`,
		requestID,
		userID,
		prediction,
		time.Now().UTC(),
		model,
		usageTokensPrompt,
		usageTokensCompletions,
		predictionRaw,
	)
	return err
}

func (c Client) WriteSuccessFlag(ctx context.Context, requestID, userID, token string) error {
	if requestID == "" {
		return errors.New("request_id is required")
	}
	if userID == "" {
		return errors.New("user_id is required")
	}

	if token != "" {
		_, err := c.c.Exec(
			ctx, `INSERT INTO `+c.tableWriteSuccessFlag+
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
			requestID,
			userID,
			time.Now().UTC(),
			token,
		)
		return err
	}

	_, err := c.c.Exec(
		ctx, `INSERT INTO `+c.tableWriteSuccessFlag+
			` (
	request_id
  , user_id
  , timestamp
) VALUES (
		  $1
    	, $2
    	, $3
)`,
		requestID,
		userID,
		time.Now().UTC(),
	)

	return err
}

type mockDbClient struct {
	err   error
	query string
	v     pgx.Rows
}

func (m *mockDbClient) Query(_ context.Context, query string, _ ...any) (pgx.Rows, error) {
	m.query = query
	if m.err != nil {
		return nil, m.err
	}
	return m.v, nil
}

func (m *mockDbClient) Close(_ context.Context) error {
	return m.err
}

func (m *mockDbClient) Exec(_ context.Context, query string, _ ...any) (pgconn.CommandTag, error) {
	m.query = query
	if m.err != nil {
		return pgconn.CommandTag{}, m.err
	}
	return pgconn.NewCommandTag(strings.ToUpper(strings.Split(query, " ")[0])), nil
}

type dbClient interface {
	Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, query string, args ...any) (pgx.Rows, error)
	Close(ctx context.Context) error
}

type mockRows struct {
	tag    pgconn.CommandTag
	err    error
	v      [][]any
	rowCnt int
	s      *sync.RWMutex
}

func (m *mockRows) Close() {
	return
}

func (m *mockRows) Err() error {
	return m.err
}

func (m *mockRows) CommandTag() pgconn.CommandTag {
	return m.tag
}

func (m *mockRows) FieldDescriptions() []pgconn.FieldDescription {
	return nil
}

func (m *mockRows) Next() bool {
	m.s.Lock()
	var f bool
	if len(m.v) > m.rowCnt {
		f = true
	}
	m.s.Unlock()
	return f
}

func (m *mockRows) Scan(dest ...any) error {
	m.s.Lock()
	defer m.s.Unlock()
	if len(m.v[m.rowCnt]) != len(dest) {
		return errors.New(
			"number of field descriptions must equal number of destinations, got " +
				strconv.Itoa(len(m.v[m.rowCnt])) + " and " + strconv.Itoa(len(dest)),
		)
	}
	for i, el := range m.v[m.rowCnt] {
		switch dest[i].(type) {
		case *string:
			*dest[i].(*string) = el.(string)
		case *bool:
			*dest[i].(*bool) = el.(bool)
		case *int:
			*dest[i].(*int) = el.(int)
		}
	}
	m.rowCnt++
	return nil
}

func (m *mockRows) Values() ([]any, error) {
	m.s.Lock()
	defer m.s.Unlock()
	return m.v[m.rowCnt], m.Err()
}

func (m *mockRows) RawValues() [][]byte {
	return nil
}

func (m *mockRows) Conn() *pgx.Conn {
	return nil
}
