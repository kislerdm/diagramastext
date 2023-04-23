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

type mockDbClient struct {
	err   error
	query string
	tx    pgx.Tx
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

func (m *mockDbClient) Begin(_ context.Context) (pgx.Tx, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.tx, nil
}

type dbClient interface {
	Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, query string, args ...any) (pgx.Rows, error)
	Close(ctx context.Context) error
	Begin(ctx context.Context) (pgx.Tx, error)
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
	if m.err != nil {
		return m.err
	}

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
		case *time.Time:
			*dest[i].(*time.Time) = el.(time.Time)
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

type mockRow struct {
	v   []any
	err error
}

func (m mockRow) Scan(dest ...any) error {
	if m.err != nil {
		return m.err
	}
	for i, el := range m.v {
		switch dest[i].(type) {
		case *string:
			*dest[i].(*string) = el.(string)
		case *bool:
			*dest[i].(*bool) = el.(bool)
		case *int:
			*dest[i].(*int) = el.(int)
		case *time.Time:
			*dest[i].(*time.Time) = el.(time.Time)
		}
	}
	return nil
}

type mockTx struct {
	client dbClient
	row    mockRow
	err    error
}

func (m mockTx) Begin(_ context.Context) (pgx.Tx, error) {
	return m, m.err
}

func (m mockTx) Commit(_ context.Context) error {
	return m.err
}

func (m mockTx) Rollback(_ context.Context) error {
	return m.err
}

func (m mockTx) CopyFrom(
	_ context.Context, _ pgx.Identifier, _ []string, _ pgx.CopyFromSource,
) (int64, error) {
	return 0, m.err
}

func (m mockTx) SendBatch(_ context.Context, b *pgx.Batch) pgx.BatchResults {
	return nil
}

func (m mockTx) LargeObjects() pgx.LargeObjects {
	return pgx.LargeObjects{}
}

func (m mockTx) Prepare(_ context.Context, _, _ string) (*pgconn.StatementDescription, error) {
	return nil, m.err
}

func (m mockTx) Exec(ctx context.Context, sql string, args ...any) (commandTag pgconn.CommandTag, err error) {
	return m.client.Exec(ctx, sql, args)
}

func (m mockTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return m.client.Query(ctx, sql, args)
}

func (m mockTx) QueryRow(_ context.Context, _ string, _ ...any) pgx.Row {
	return m.row
}

func (m mockTx) Conn() *pgx.Conn {
	return m.client.(*pgx.Conn)
}
