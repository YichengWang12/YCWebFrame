package orm

import (
	"WebFrame/orm/internal/errs"
	"WebFrame/orm/model"
	"context"
	"database/sql"
)

type UpdateBuilder[T any] struct {
	i               *Inserter[T]
	conflictColumns []string
}

type Update struct {
	conflictColumns []string
	assigns         []Assignable
}

func (u *UpdateBuilder[T]) ConflictColumns(cols ...string) *UpdateBuilder[T] {
	u.conflictColumns = cols
	return u
}

func (u *UpdateBuilder[T]) Update(assigns ...Assignable) *Inserter[T] {
	u.i.onDuplicate = &Update{
		conflictColumns: u.conflictColumns,
		assigns:         assigns,
	}
	return u.i
}

type Inserter[T any] struct {
	builder
	values  []*T
	columns []string

	onDuplicate *Update
	sess        session
}

func (i *Inserter[T]) OnDuplicateKey() *UpdateBuilder[T] {
	return &UpdateBuilder[T]{
		i: i,
	}

}

func NewInserter[T any](sess session) *Inserter[T] {
	c := sess.getCore()
	return &Inserter[T]{
		sess: sess,
		builder: builder{
			core:    c,
			dialect: c.dialect,
			quoter:  c.dialect.quoter(),
		},
	}
}

// Values select the values to be inserted
func (i *Inserter[T]) Values(vals ...*T) *Inserter[T] {
	i.values = vals
	return i
}

// Columns select the columns to be inserted
func (i *Inserter[T]) Columns(cols ...string) *Inserter[T] {
	i.columns = cols
	return i
}

func (i *Inserter[T]) Build() (*Query, error) {
	if len(i.values) == 0 {
		return nil, errs.ErrInsertZeroRow
	}
	m, err := i.r.Get(i.values[0])
	if err != nil {
		return nil, err
	}
	i.model = m
	i.sb.WriteString("INSERT INTO ")
	i.quote(m.TableName)
	i.sb.WriteString("(")
	fields := m.Fields
	if len(i.columns) != 0 {
		fields = make([]*model.Field, 0, len(i.columns))
		for _, c := range i.columns {
			field, ok := m.FieldMap[c]
			if !ok {
				return nil, errs.NewErrUnknownField(c)
			}
			fields = append(fields, field)
		}
	}

	i.args = make([]any, 0, len(fields)*(len(i.values)+1))
	for idx, fd := range fields {
		if idx > 0 {
			i.sb.WriteString(", ")
		}
		i.quote(fd.ColName)
	}
	i.sb.WriteString(") VALUES")
	for vIdx, val := range i.values {
		if vIdx > 0 {
			i.sb.WriteString(",")
		}
		refVal := i.valCreator(val, i.model)

		i.sb.WriteByte('(')
		for fIdx, field := range fields {
			if fIdx > 0 {
				i.sb.WriteByte(',')
			}
			i.sb.WriteByte('?')
			fdVal, err := refVal.Field(field.GoName)
			if err != nil {
				return nil, err
			}
			i.addArgs(fdVal)
		}
		i.sb.WriteByte(')')

	}

	if i.onDuplicate != nil {
		err = i.dialect.buildUpdate(&i.builder, i.onDuplicate)
		if err != nil {
			return nil, err
		}
	}
	i.sb.WriteByte(';')
	return &Query{
		SQL:  i.sb.String(),
		Args: i.args,
	}, nil

}

//func (i *Inserter[T]) addArgs(args ...any) {
//	i.args = append(i.args, args...)
//}

func (i *Inserter[T]) Exec(ctx context.Context) sql.Result {
	q, err := i.Build()
	if err != nil {
		return Result{err: err}
	}
	res, err := i.sess.execContext(ctx, q.SQL, q.Args)
	return Result{err: err, res: res}
}
