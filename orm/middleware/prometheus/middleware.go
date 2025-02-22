package prometheus

import (
	"WebFrame/orm"
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type MiddlewareBuilder struct {
	Name        string
	Subsystem   string
	ConstLabels map[string]string
	Help        string
}

func (m MiddlewareBuilder) Build() orm.Middleware {
	summaryVec := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:        m.Name,
		Subsystem:   m.Subsystem,
		ConstLabels: m.ConstLabels,
		Help:        m.Help,
	}, []string{"type", "table"})

	return func(next orm.HandleFunc) orm.HandleFunc {
		return func(ctx context.Context, qc *orm.QueryContext) *orm.QueryResult {
			startTime := time.Now()
			defer func() {
				endTime := time.Now()
				typ := "unknown"
				// 原生查询才会走到这里
				tblName := "unknown"
				if qc.Model != nil {
					typ = qc.Model.TableName
					tblName = qc.Model.TableName
				}
				summaryVec.WithLabelValues(typ, tblName).
					Observe(float64(endTime.Sub(startTime).Milliseconds()))
			}()
			return next(ctx, qc)
		}
	}
}
