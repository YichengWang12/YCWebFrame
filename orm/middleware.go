package orm

import (
	"WebFrame/orm/model"
	"context"
)

type QueryContext struct {
	Type    string
	Builder QueryBuilder
	Model   *model.Model
}

//func (qc *QueryContext) Query() (*Query, error) {
//	if qc.q != nil {
//		return qc.q, nil
//	}
//	var err error
//	qc.q, err = qc.Builder.Build()
//	return qc.q, err
//
//}

type QueryResult struct {
	// result is different types in different queries
	// in Selector.Get, it will be a single result
	// in Selector.GetMulti, it will be a slice
	// in other cases, it will be a Result type
	Res any
	Err error
}

type Middleware func(next HandleFunc) HandleFunc

type HandleFunc func(ctx context.Context, qc *QueryContext) *QueryResult
