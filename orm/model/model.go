package model

import (
	"reflect"
)

type ModelOpt func(model *Model) error

type Model struct {
	TableName string
	FieldMap  map[string]*Field
	ColumnMap map[string]*Field
}

type Field struct {
	ColName string
	GoName  string
	Type    reflect.Type

	Offset uintptr
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
