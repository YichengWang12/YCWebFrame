package errhdl

import "WebFrame/web"

type MiddlewareBuilder struct {
	resp map[int][]byte
}

func NewMiddlewareBuilder() *MiddlewareBuilder {
	return &MiddlewareBuilder{resp: make(map[int][]byte, 64)}
}

func (b *MiddlewareBuilder) RegisterError(code int, resp []byte) *MiddlewareBuilder {
	b.resp[code] = resp
	return b
}

func (b *MiddlewareBuilder) Build() web.Middleware {
	return func(next web.HandleFunc) web.HandleFunc {
		return func(ctx *web.Context) {
			next(ctx)
			resp, ok := b.resp[ctx.RespStatusCode]
			if ok {
				ctx.RespData = resp
			}
		}
	}
}
