package orm

import (
	"WebFrame/orm/internal/errs"
	"reflect"
	"unicode"
)

type model struct {
	tableName string
	fieldMap  map[string]*field
}

type field struct {
	colName string
}

func parseModel(val any) (*model, error) {
	typ := reflect.TypeOf(val)
	if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Struct {
		return nil, errs.ErrPointerOnly
	}
	typ = typ.Elem()
	// get the number of fields
	numField := typ.NumField()
	fds := make(map[string]*field, numField)
	for i := 0; i < numField; i++ {
		fdType := typ.Field(i)
		fds[fdType.Name] = &field{
			colName: underscoreName(fdType.Name),
		}
	}

	return &model{
		tableName: underscoreName(typ.Name()),
		fieldMap:  fds,
	}, nil
}

// undersocreName converts camel case to snake case
func underscoreName(tableName string) string {
	var buf []byte
	for i, v := range tableName {
		if unicode.IsUpper(v) {
			if i != 0 {
				buf = append(buf, '_')
			}
			buf = append(buf, byte(unicode.ToLower(v)))
		} else {
			buf = append(buf, byte(v))
		}
	}
	return string(buf)
}
