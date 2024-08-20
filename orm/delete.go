package orm

import (
	"WebFrame/orm/internal/errs"
	"reflect"
)

type Deleter[T any] struct {
	builder
	table string
	where []Predicate

	sess session
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
	d.model, err = d.r.Get(&t)
	if err != nil {
		return nil, err
	}
	d.sb.WriteString("DELETE FROM ")
	if d.table == "" {
		var t T

		d.quote(reflect.TypeOf(t).Name())
	} else {
		d.sb.WriteString(d.table)
	}
	if len(d.where) > 0 {
		d.sb.WriteString(" WHERE ")
		p := d.where[0]
		for i := 1; i < len(d.where); i++ {
			p = p.And(d.where[i])
		}
		if er := d.buildExpression(p); er != nil {
			return nil, er
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
		fd, ok := d.model.FieldMap[exp.name]
		if !ok {
			return errs.NewErrUnknownField(exp.name)
		}
		d.quote(fd.ColName)
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

func NewDeleter[T any](sess session) *Deleter[T] {
	c := sess.getCore()
	return &Deleter[T]{
		sess: sess,
		builder: builder{
			core: c,

			dialect: c.dialect,
			quoter:  c.dialect.quoter(),
		},
	}
}
