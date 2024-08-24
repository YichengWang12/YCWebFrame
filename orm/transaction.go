package orm

import (
	"context"
	"database/sql"
)

var _ session = &Tx{}
var _ session = &DB{}

type session interface {
	getCore() core
	queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	execContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type Tx struct {
	tx   *sql.Tx
	db   *DB
	done bool
}

func (t *Tx) getCore() core {
	return t.db.core
}

func (t *Tx) queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return t.tx.QueryContext(ctx, query, args...)
}

func (t *Tx) execContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}

func (t *Tx) Commit() error {
	t.done = true
	return t.tx.Commit()
}

func (t *Tx) Rollback() error {
	t.done = true
	return t.tx.Rollback()
}
