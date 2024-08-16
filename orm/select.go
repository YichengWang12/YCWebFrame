package orm

import (
	"WebFrame/orm/internal/errs"
	model2 "WebFrame/orm/model"
	"context"
	"reflect"
	"strings"
)

type Selector[T any] struct {
	sb    strings.Builder
	args  []any
	where []Predicate
	table string
	model *model2.Model

	db *DB
}

func (s *Selector[T]) From(tbl string) *Selector[T] {
	s.table = tbl
	return s
}

func (s *Selector[T]) Build() (*Query, error) {
	var (
		t   T
		err error
	)
	s.model, err = s.db.r.Get(&t)
	if err != nil {
		return nil, err
	}
	s.sb.WriteString("SELECT * FROM ")
	if s.table == "" {
		var t T
		s.sb.WriteByte('`')
		s.sb.WriteString(reflect.TypeOf(t).Name())
		s.sb.WriteByte('`')
	} else {
		s.sb.WriteString(s.table)
	}
	// part where
	if len(s.where) > 0 {
		s.sb.WriteString(" WHERE ")
		p := s.where[0]
		for i := 1; i < len(s.where); i++ {
			p = p.And(s.where[i])
		}
		if err := s.buildExpression(p); err != nil {
			return nil, err
		}
	}
	s.sb.WriteByte(';')
	return &Query{
		SQL:  s.sb.String(),
		Args: s.args,
	}, nil

}

func (s *Selector[T]) buildExpression(e Expression) error {
	if e == nil {
		return nil
	}
	switch exp := e.(type) {
	case Column:
		fd, ok := s.model.FieldMap[exp.name]
		if !ok {
			return errs.NewErrUnknownField(exp.name)
		}
		s.sb.WriteByte('`')
		s.sb.WriteString(fd.ColName)
		s.sb.WriteByte('`')
	case value:
		s.sb.WriteByte('?')
		s.args = append(s.args, exp.val)
	case Predicate:
		_, lp := exp.left.(Predicate)
		if lp {
			s.sb.WriteByte('(')
		}
		if err := s.buildExpression(exp.left); err != nil {
			return err
		}
		if lp {
			s.sb.WriteByte(')')
		}
		s.sb.WriteByte(' ')
		s.sb.WriteString(exp.op.String())
		s.sb.WriteByte(' ')
		_, rp := exp.right.(Predicate)
		if rp {
			s.sb.WriteByte('(')
		}
		if err := s.buildExpression(exp.right); err != nil {
			return err
		}
		if rp {
			s.sb.WriteByte(')')
		}
	default:
		return errs.NewErrUnsupportedExpressionType(exp)
	}
	return nil
}

func (s *Selector[T]) Where(ps ...Predicate) *Selector[T] {
	s.where = ps
	return s
}

func (s *Selector[T]) Get(ctx context.Context) (*T, error) {
	q, err := s.Build()
	if err != nil {
		return nil, err
	}
	//s.db is our defined DB
	//s.db.db is the *sql.DB
	rows, err := s.db.db.QueryContext(ctx, q.SQL, q.Args...)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, errs.ErrNoRows
	}
	// construct
	tp := new(T)
	meta, err := s.db.r.Get(tp)
	if err != nil {
		return nil, err
	}
	val := s.db.valCreator(tp, meta)
	err = val.SetColumns(rows)
	return tp, err
}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	q, err := s.Build()
	if err != nil {
		return nil, err
	}
	rows, err := s.db.db.QueryContext(ctx, q.SQL, q.Args...)
	if err != nil {
		return nil, err
	}
	tps := make([]*T, 2)
	count := 0
	for rows.Next() {
		// construct
		tp := new(T)
		meta, er := s.db.r.Get(tp)
		if er != nil {
			return nil, er
		}
		val := s.db.valCreator(tp, meta)
		er = val.SetColumns(rows)
		if er != nil && count <= cap(tps) {
			tps[count] = tp
			count++
		}
	}
	return tps, nil
}

func NewSelector[T any](db *DB) *Selector[T] {
	return &Selector[T]{
		db: db,
	}
}
