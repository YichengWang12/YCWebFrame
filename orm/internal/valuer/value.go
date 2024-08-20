package valuer

import (
	"WebFrame/orm/model"
	"database/sql"
)

// Value is an internal abstraction of a struct instance
type Value interface {
	// Field returns the value of the field with the given name
	Field(name string) (any, error)
	// SetColumns sets the columns of the struct instance
	SetColumns(rows *sql.Rows) error
}

type Creator func(val interface{}, meta *model.Model) Value
