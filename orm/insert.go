package orm

import (
	"WebFrame/orm/internal/errs"
	"WebFrame/orm/model"
	"reflect"
	"strings"
)

type onDuplicateKeyBuilder[T any] struct {
	i *Inserter[T]
}

type onDuplicateKey struct {
	assigns []Assignable
}

type Inserter[T any] struct {
	values  []*T
	db      *DB
	columns []string
	sb      strings.Builder
	args    []any
	model   *model.Model

	onDuplicateKey *onDuplicateKey
}

func (o *onDuplicateKeyBuilder[T]) Update(assigns ...Assignable) *Inserter[T] {
	o.i.onDuplicateKey = &onDuplicateKey{
		assigns: assigns,
	}
	return o.i
}

func (i *Inserter[T]) OnDuplicateKey() *onDuplicateKeyBuilder[T] {
	return &onDuplicateKeyBuilder[T]{
		i: i,
	}

}

func NewInserter[T any](db *DB) *Inserter[T] {
	return &Inserter[T]{
		db: db,
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
	m, err := i.db.r.Get(i.values[0])
	if err != nil {
		return nil, err
	}
	i.model = m
	i.sb.WriteString("INSERT INTO ")
	i.sb.WriteByte('`')
	i.sb.WriteString(m.TableName)
	i.sb.WriteString("`(")
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
		i.sb.WriteString("`")
		i.sb.WriteString(fd.ColName)
		i.sb.WriteString("`")
	}
	i.sb.WriteString(") VALUES")
	for vIdx, val := range i.values {
		if vIdx > 0 {
			i.sb.WriteString(",")
		}
		refVal := reflect.ValueOf(val).Elem()
		i.sb.WriteByte('(')
		for fIdx, field := range fields {
			if fIdx > 0 {
				i.sb.WriteByte(',')
			}
			i.sb.WriteByte('?')
			fdVal := refVal.Field(field.Index)
			i.addArgs(fdVal.Interface())
		}
		i.sb.WriteByte(')')

	}

	if i.onDuplicateKey != nil {
		i.sb.WriteString(" ON DUPLICATE KEY UPDATE ")
		for idx, assign := range i.onDuplicateKey.assigns {
			if idx > 0 {
				i.sb.WriteByte(',')
			}
			if err = i.buildAssignment(assign); err != nil {
				return nil, err
			}
		}
	}
	i.sb.WriteByte(';')
	return &Query{
		SQL:  i.sb.String(),
		Args: i.args,
	}, nil

}

func (i *Inserter[T]) buildAssignment(a Assignable) error {
	switch assign := a.(type) {
	case Column:
		i.sb.WriteByte('`')
		fd, ok := i.model.FieldMap[assign.name]
		if !ok {
			return errs.NewErrUnknownField(assign.name)
		}
		i.sb.WriteString(fd.ColName)
		i.sb.WriteByte('`')
		i.sb.WriteString("=VALUES(")
		i.sb.WriteByte('`')
		i.sb.WriteString(fd.ColName)
		i.sb.WriteByte('`')
		i.sb.WriteByte(')')
	case Assignment:
		i.sb.WriteByte('`')
		fd, ok := i.model.FieldMap[assign.col]
		if !ok {
			return errs.NewErrUnknownField(assign.col)
		}
		i.sb.WriteString(fd.ColName)
		i.sb.WriteByte('`')
		i.sb.WriteString("= ?")
		i.addArgs(assign.val)
	default:
		return errs.NewErrUnsupportedAssignableType(a)
	}
	return nil
}

func (i *Inserter[T]) addArgs(args ...any) {
	i.args = append(i.args, args...)
}
