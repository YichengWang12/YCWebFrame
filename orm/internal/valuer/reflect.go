package valuer

import (
	"WebFrame/orm/internal/errs"
	"WebFrame/orm/model"
	"database/sql"
	"reflect"
)

// reflectValue is a Value implementation based on reflection
type reflectValue struct {
	val  reflect.Value
	meta *model.Model
}

var _ Creator = NewReflectValue

// NewReflectValue returns a wrapped Value based on reflection
// The input val must be a pointer to a struct instance, not any other type
func NewReflectValue(val interface{}, meta *model.Model) Value {
	return reflectValue{
		val:  reflect.ValueOf(val).Elem(), // reflect.ValueOf(val) returns a pointer, so Elem() is needed to get the struct instance
		meta: meta,
	}
}

func (r reflectValue) Field(name string) (any, error) {
	res := r.val.FieldByName(name)
	if res == (reflect.Value{}) {
		return nil, errs.NewErrUnknownField(name)
	}
	return res.Interface(), nil

}

func (r reflectValue) SetColumns(rows *sql.Rows) error {
	cs, err := rows.Columns()
	if err != nil {
		return err
	}
	if len(cs) > len(r.meta.FieldMap) {
		return errs.ErrTooManyReturnedColumns
	}
	colValues := make([]interface{}, len(cs))
	colElmValues := make([]reflect.Value, len(cs))
	for i, c := range cs {
		cm, ok := r.meta.ColumnMap[c]
		if !ok {
			return errs.NewErrUnknownColumn(c)
		}
		val := reflect.New(cm.Type)
		colValues[i] = val.Interface()
		colElmValues[i] = val.Elem()
	}
	if err = rows.Scan(colValues...); err != nil {
		return err
	}
	for i, c := range cs {
		cm := r.meta.ColumnMap[c]
		fd := r.val.FieldByName(cm.GoName)
		fd.Set(colElmValues[i])
	}
	return nil
}
