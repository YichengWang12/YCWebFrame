package orm

import (
	"WebFrame/orm/internal/errs"
	"unicode"
)

type ModelOpt func(model *Model) error

type Model struct {
	tableName string
	fieldMap  map[string]*field
}

type field struct {
	colName string
}

func ModelWithTableName(tableName string) ModelOpt {
	return func(model *Model) error {
		model.tableName = tableName
		return nil
	}
}

func ModelWithColumnName(field string, columnName string) ModelOpt {
	return func(model *Model) error {
		fd, ok := model.fieldMap[field]
		if !ok {
			return errs.NewErrUnknownField(field)
		}
		fd.colName = columnName
		return nil
	}
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

// We put all the keys of the tags we support here
// to make it easier for users to find and for us to maintain
const (
	tagKeyColumn = "column"
)

// TableName is an interface that users can implement to return a custom table name
type TableName interface {
	TableName() string
}
