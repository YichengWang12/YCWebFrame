package orm

import (
	"WebFrame/orm/internal/errs"
	"reflect"
	"strings"
)

type Deleter[T any] struct {
	sb    strings.Builder
	table string
	where []Predicate
	args  []any
	model *Model

	db *DB
}

func (d *Deleter[T]) From(tbl string) *Deleter[T] {
	d.table = tbl
	return d
}

func (d *Deleter[T]) Build() (*Query, error) {
	var (
		t   T
		err error
	)
	d.model, err = d.db.r.Get(&t)
	if err != nil {
		return nil, err
	}
	d.sb.WriteString("DELETE FROM ")
	if d.table == "" {
		var t T
		d.sb.WriteByte('`')
		d.sb.WriteString(reflect.TypeOf(t).Name())
		d.sb.WriteByte('`')
	} else {
		d.sb.WriteString(d.table)
	}
	if len(d.where) > 0 {
		d.sb.WriteString(" WHERE ")
		p := d.where[0]
		for i := 1; i < len(d.where); i++ {
			p = p.And(d.where[i])
		}
		if err := d.buildExpression(p); err != nil {
			return nil, err
		}

	}
	d.sb.WriteByte(';')
	return &Query{
		SQL:  d.sb.String(),
		Args: d.args,
	}, nil
}

func (d *Deleter[T]) buildExpression(e Expression) error {
	if e == nil {
		return nil
	}
	switch exp := e.(type) {
	case Column:
		fd, ok := d.model.fieldMap[exp.name]
		if !ok {
			return errs.NewErrUnknownField(exp.name)
		}
		d.sb.WriteByte('`')
		d.sb.WriteString(fd.colName)
		d.sb.WriteByte('`')
	case value:
		d.sb.WriteByte('?')
		d.args = append(d.args, exp.val)
	case Predicate:
		_, lp := exp.left.(Predicate)
		if lp {
			d.sb.WriteByte('(')
		}
		if err := d.buildExpression(exp.left); err != nil {
			return err
		}
		if lp {
			d.sb.WriteByte(')')
		}
		d.sb.WriteString(" ")
		d.sb.WriteString(exp.op.String())
		d.sb.WriteString(" ")
		_, rp := exp.right.(Predicate)
		if rp {
			d.sb.WriteByte('(')
		}
		if err := d.buildExpression(exp.right); err != nil {
			return err
		}
		if rp {
			d.sb.WriteByte(')')
		}
	default:
		return errs.NewErrUnsupportedExpressionType(exp)
	}
	return nil

}

func (d *Deleter[T]) Where(preds ...Predicate) *Deleter[T] {
	d.where = preds
	return d
}

func NewDeleter[T any](db *DB) *Deleter[T] {
	return &Deleter[T]{
		db: db,
	}
}
