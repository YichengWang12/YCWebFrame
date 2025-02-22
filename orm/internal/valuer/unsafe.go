package valuer

import (
	"WebFrame/orm/internal/errs"
	"WebFrame/orm/model"
	"database/sql"
	"reflect"
	"unsafe"
)

type unsafeValue struct {
	addr unsafe.Pointer
	meta *model.Model
}

var _ Creator = NewUnsafeValue

func NewUnsafeValue(val interface{}, meta *model.Model) Value {
	return unsafeValue{
		addr: unsafe.Pointer(reflect.ValueOf(val).Pointer()),
		meta: meta,
	}
}

func (u unsafeValue) Field(name string) (any, error) {
	cm, ok := u.meta.FieldMap[name]
	if !ok {
		return nil, errs.NewErrUnknownField(name)
	}
	ptr := unsafe.Pointer(uintptr(u.addr) + cm.Offset)
	val := reflect.NewAt(cm.Type, ptr)
	return val.Elem().Interface(), nil
}

func (u unsafeValue) SetColumns(rows *sql.Rows) error {
	cs, err := rows.Columns()
	if err != nil {
		return err
	}
	if len(cs) > len(u.meta.ColumnMap) {
		return errs.ErrTooManyReturnedColumns
	}
	colValues := make([]interface{}, len(cs))
	for i, c := range cs {
		cm, ok := u.meta.ColumnMap[c]
		if !ok {
			return errs.NewErrUnknownColumn(c)
		}
		ptr := unsafe.Pointer(uintptr(u.addr) + cm.Offset)
		val := reflect.NewAt(cm.Type, ptr)
		colValues[i] = val.Interface()
	}
	return rows.Scan(colValues...)
}
