package valuer

import (
	"WebFrame/orm/model"
	"database/sql"
)

// Value is an internal abstraction of a struct instance
type Value interface {
	// SetColumns sets the columns of the struct instance
	SetColumns(rows *sql.Rows) error
}

type Creator func(val interface{}, meta *model.Model) Value
