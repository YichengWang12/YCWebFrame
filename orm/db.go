package orm

import (
	"WebFrame/orm/internal/valuer"
	"WebFrame/orm/model"
	"database/sql"
)

type DB struct {
	r          model.Registry
	db         *sql.DB
	valCreator valuer.Creator
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
		r:          model.NewRegistry(),
		db:         db,
		valCreator: valuer.NewUnsafeValue,
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil

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

// MustNewDB provides a way to create a DB and panic if it fails.
//func MustNewDB(opts ...DBOption) *DB {
//	db := NewDB(opts...)
//	return db
//}
