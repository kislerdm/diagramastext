package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
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

func (c Client) GetDailySuccessfulResultsTimestampsByUserID(ctx context.Context, userID string) ([]time.Time, error) {
	rows, err := c.c.Query(
		ctx, `SELECT timestamp FROM `+c.tableWriteSuccessFlag+
			` WHERE timestamp::date = current_date AND user_id = $1`, userID,
	)
	if err != nil {
		return nil, err
	}

	var o []time.Time
	var ts time.Time
	for rows.Next() {
		if err := rows.Scan(&ts); err != nil {
			return nil, err
		}
		o = append(o, ts)
	}
	rows.Close()
	return o, nil
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

func (c Client) CreateUser(ctx context.Context, id, email, fingerprint string, isActive bool) error {
	if id == "" {
		return errors.New("id is required")
	}
	_, err := c.c.Exec(
		ctx, "INSERT INTO "+c.tableUsers+" (user_id,email,web_fingerprint,is_active) VALUES ($1,$2,$3,$4)",
		id, email, fingerprint, isActive,
	)
	return err
}

func (c Client) ReadUser(ctx context.Context, id string) (
	found, isActive, emailVerified bool, email, fingerprint string, err error,
) {
	if id == "" {
		err = errors.New("id is required")
		return
	}
	rows, err := c.c.Query(
		ctx, `SELECT 
	is_active
	,email
    ,email_verified
	,web_fingerprint
FROM `+c.tableUsers+` WHERE user_id = $1`, id,
	)
	if err != nil {
		return
	}
	if rows.Next() {
		if err := rows.Scan(&isActive, &email, &emailVerified, &fingerprint); err != nil {
			return false, false, false, "", "", err
		}
		rows.Close()

		found = true
		return
	}
	return false, false, false, "", "", nil
}

func (c Client) LookupUserByEmail(ctx context.Context, email string) (id string, isActive bool, err error) {
	if email == "" {
		err = errors.New("email is required")
		return
	}
	rows, err := c.c.Query(
		ctx, `SELECT user_id, is_active FROM `+c.tableUsers+
			// The last registered user with the given email will be selected
			// FIXME: shall this behaviour be sustained?
			// FIXME: consider alternatives to ORDER BY for the sake of performance
			` WHERE email = $1 ORDER BY created_at LIMIT 1`, email,
	)
	if err != nil {
		return
	}
	if rows.Next() {
		if err = rows.Scan(&id, &isActive); err != nil {
			return
		}
		rows.Close()
	}
	return
}

func (c Client) LookupUserByFingerprint(ctx context.Context, fingerprint string) (id string, isActive bool, err error) {
	if fingerprint == "" {
		err = errors.New("fingerprint is required")
		return
	}
	rows, err := c.c.Query(
		ctx, `SELECT user_id, is_active FROM `+c.tableUsers+
			// The last registered user with the given fingerprint will be selected
			// FIXME: shall this behaviour be sustained?
			// FIXME: consider alternatives to ORDER BY for the sake of performance
			` WHERE fingerprint = $1 ORDER BY created_at LIMIT 1`, fingerprint,
	)
	if err != nil {
		return
	}
	if rows.Next() {
		if err = rows.Scan(&id, &isActive); err != nil {
			return
		}
		rows.Close()
	}
	return
}

func (c Client) UpdateUserSetEmailVerified(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id is required")
	}
	_, err := c.c.Exec(
		ctx, "UPDATE "+c.tableUsers+
			" SET email_verified = TRUE, is_active = TRUE WHERE user_id = %1", id,
	)
	return err
}
