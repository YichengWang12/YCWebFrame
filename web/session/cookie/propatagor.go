package cookie

import (
	"net/http"
)

type PropagatorOption func(propagator *Propagator)

type Propagator struct {
	cookieName string
	cookieOpt  func(cookie *http.Cookie)
}

func WithCookieOption(opt func(cookie *http.Cookie)) PropagatorOption {
	return func(propagator *Propagator) {
		propagator.cookieOpt = opt
	}
}

func NewPropagator(cookieName string, opts ...PropagatorOption) *Propagator {
	p := &Propagator{
		cookieName: cookieName,
		cookieOpt: func(cookie *http.Cookie) {
		},
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *Propagator) Inject(id string, writer http.ResponseWriter) error {
	cookie := &http.Cookie{
		Name:  p.cookieName,
		Value: id,
	}
	p.cookieOpt(cookie)
	http.SetCookie(writer, cookie)
	return nil
}

func (p *Propagator) Extract(req *http.Request) (string, error) {
	cookie, err := req.Cookie(p.cookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func (p *Propagator) Remove(writer http.ResponseWriter) error {
	cookie := &http.Cookie{Name: p.cookieName, MaxAge: -1}
	p.cookieOpt(cookie)
	http.SetCookie(writer, cookie)
	return nil
}
