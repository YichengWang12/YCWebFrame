package orm

import (
	"WebFrame/orm/internal/valuer"
	"WebFrame/orm/model"
	"context"
	"database/sql"
)

type DB struct {
	db *sql.DB

	core
}

type DBOption func(*DB)

func Open(driver string, dsn string, opts ...DBOption) (*DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	return OpenDB(db, opts...)

}

func OpenDB(db *sql.DB, opts ...DBOption) (*DB, error) {
	res := &DB{
		core: core{
			dialect:    MySQL,
			r:          model.NewRegistry(),
			valCreator: valuer.NewUnsafeValue,
		},
		db: db,
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil

}

func DBWithDialect(d Dialect) DBOption {
	return func(db *DB) {
		db.dialect = d
	}
}

func DBWithRegistry(r model.Registry) DBOption {
	return func(db *DB) {
		db.r = r
	}
}

func DBUseReflectValuer() DBOption {
	return func(db *DB) {
		db.valCreator = valuer.NewReflectValue
	}
}

func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Tx{tx: tx}, nil
}

func (db *DB) queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return db.db.QueryContext(ctx, query, args...)
}

func (db *DB) execContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return db.db.ExecContext(ctx, query, args...)
}
func (db *DB) getCore() core {
	return db.core
}

// MustNewDB provides a way to create a DB and panic if it fails.
//func MustNewDB(opts ...DBOption) *DB {
//	db := NewDB(opts...)
//	return db
//}
