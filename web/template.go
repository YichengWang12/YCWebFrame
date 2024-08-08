package web

import "context"

type TemplateEngine interface {
	// Render 渲染页面
	Render(ctx context.Context, tplName string, data any) ([]byte, error)
}
