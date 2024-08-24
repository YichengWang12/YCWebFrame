package orm

import (
	"WebFrame/orm/internal/errs"
	"WebFrame/orm/model"
	"context"
	"fmt"
)

type UpsertBuilder[T any] struct {
	i               *Inserter[T]
	conflictColumns []string
}

type Upsert struct {
	conflictColumns []string
	assigns         []Assignable
}

func (u *UpsertBuilder[T]) ConflictColumns(cols ...string) *UpsertBuilder[T] {
	u.conflictColumns = cols
	return u
}

func (u *UpsertBuilder[T]) Update(assigns ...Assignable) *Inserter[T] {
	u.i.upsert = &Upsert{
		conflictColumns: u.conflictColumns,
		assigns:         assigns,
	}
	return u.i
}

type Inserter[T any] struct {
	builder
	values  []*T
	columns []string

	upsert *Upsert
	sess   session
	core
}

func NewInserter[T any](sess session) *Inserter[T] {
	c := sess.getCore()
	return &Inserter[T]{
		core: c,
		sess: sess,
		builder: builder{
			dialect: c.dialect,
			quoter:  c.dialect.quoter(),
		},
	}
}

func (i *Inserter[T]) OnDuplicateKey() *UpsertBuilder[T] {
	return &UpsertBuilder[T]{
		i: i,
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
	fmt.Println("i.values", i.values)
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

	if i.upsert != nil {
		err = i.core.dialect.buildUpsert(&i.builder, i.upsert)
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

func (i *Inserter[T]) Exec(ctx context.Context) Result {
	return exec(ctx, i.sess, i.core, &QueryContext{
		Builder: i,
		Type:    "INSERT",
	})
}
