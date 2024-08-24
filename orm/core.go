package orm

import (
	"WebFrame/orm/internal/errs"
	"WebFrame/orm/internal/valuer"
	"WebFrame/orm/model"
	"context"
	"database/sql"
)

type core struct {
	r          model.Registry
	dialect    Dialect
	valCreator valuer.Creator
	ms         []Middleware
}

func getHandler[T any](ctx context.Context,
	sess session,
	c core,
	qc *QueryContext) *QueryResult {
	q, err := qc.Builder.Build()
	if err != nil {
		return &QueryResult{
			Err: err,
		}
	}
	rows, err := sess.queryContext(ctx, q.SQL, q.Args...)
	if err != nil {
		return &QueryResult{
			Err: err,
		}
	}

	if !rows.Next() {
		return &QueryResult{
			Err: errs.ErrNoRows,
		}
	}

	tp := new(T)
	meta, err := c.r.Get(tp)
	if err != nil {
		return &QueryResult{
			Err: err,
		}
	}
	val := c.valCreator(tp, meta)
	err = val.SetColumns(rows)
	return &QueryResult{
		Res: tp,
		Err: err,
	}
}

func get[T any](ctx context.Context, c core, sess session, qc *QueryContext) *QueryResult {
	var handler HandleFunc = func(ctx context.Context, qc *QueryContext) *QueryResult {
		return getHandler[T](ctx, sess, c, qc)
	}
	ms := c.ms
	for i := len(ms) - 1; i >= 0; i-- {
		handler = ms[i](handler)
	}
	return handler(ctx, qc)
}

func exec(ctx context.Context, sess session, c core, qc *QueryContext) Result {
	var handler HandleFunc = func(ctx context.Context, qc *QueryContext) *QueryResult {
		q, err := qc.Builder.Build()
		if err != nil {
			return &QueryResult{
				Err: err,
			}
		}
		res, err := sess.execContext(ctx, q.SQL, q.Args...)
		return &QueryResult{Err: err, Res: res}
	}
	ms := c.ms
	for i := len(ms) - 1; i >= 0; i-- {
		handler = ms[i](handler)
	}
	qr := handler(ctx, qc)
	var res sql.Result
	if qr.Res != nil {
		res = qr.Res.(sql.Result)
	}
	return Result{err: qr.Err, res: res}
}
