package orm

type DB struct {
	r *registry
}

type DBOption func(*DB)

func NewDB(opts ...DBOption) (*DB, error) {
	res := &DB{
		r: &registry{},
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

// MustNewDB provides a way to create a DB and panic if it fails.
//func MustNewDB(opts ...DBOption) *DB {
//	db := NewDB(opts...)
//	return db
//}
