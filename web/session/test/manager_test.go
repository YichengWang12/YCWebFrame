package test

import (
	"WebFrame/web"
	"WebFrame/web/session"
	"WebFrame/web/session/cookie"
	"WebFrame/web/session/memory"
	"github.com/google/uuid"
	"log"
	"net/http"
	"testing"
	"time"
)

func TestSession(t *testing.T) {
	s := web.NewHTTPServer()
	m := session.Manager{
		Store: memory.NewStore(time.Minute * 30),
		Propagator: cookie.NewPropagator("sessid", cookie.WithCookieOption(func(c *http.Cookie) {
			c.HttpOnly = true
		})),
		SessCtxKey: "_sess",
	}

	s.Post("/login", func(ctx *web.Context) {
		//login check logic here
		id := uuid.New()
		// init the session for the ctx
		sess, err := m.InitSession(ctx, id.String())
		if err != nil {
			ctx.RespStatusCode = http.StatusInternalServerError
			return
		}

		// set the session value
		err = sess.Set(ctx.Req.Context(), "mykey", "some value")
		if err != nil {
			ctx.RespStatusCode = http.StatusInternalServerError
			return
		}
		ctx.RespData = []byte("login success")
	})

	s.Get("/resource", func(ctx *web.Context) {
		// get the session from ctx
		sess, err := m.GetSession(ctx)
		if err != nil {
			ctx.RespStatusCode = http.StatusInternalServerError
			return
		}
		// get the value from session
		val, err := sess.Get(ctx.Req.Context(), "mykey")
		ctx.RespData = []byte(val)
	})

	s.Post("/logout", func(ctx *web.Context) {
		// remove the session
		_ = m.RemoveSession(ctx)
	})

	s.Use(func(next web.HandleFunc) web.HandleFunc {
		return func(ctx *web.Context) {
			// do some check
			if ctx.Req.URL.Path != "/login" {
				sess, err := m.GetSession(ctx)
				if err != nil {
					ctx.RespStatusCode = http.StatusUnauthorized
					ctx.RespData = []byte("unauthorized")
					return
				}
				ctx.UserValues["sess"] = sess
				_ = m.Refresh(ctx.Req.Context(), sess.ID())
			}
			next(ctx)
		}
	})

	err := s.Start(":8081")
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	} else {
		log.Println("Server started on port 8082")
	}
}
