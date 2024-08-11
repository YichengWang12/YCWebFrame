package session

import (
	"context"
	"net/http"
)

type Store interface {
	Generate(ctx context.Context, id string) (Session, error)
	Refresh(ctx context.Context, id string) error
	Remove(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (Session, error)
}

type Session interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, val string) error
	ID() string
}

type Propagator interface {

	// Inject session id into the response
	// Inject must be idempotent
	Inject(id string, writer http.ResponseWriter) error

	// Extract session id from http.Request
	// For example, extract the session id from the cookie
	Extract(req *http.Request) (string, error)

	// Remove session id from http.ResponseWriter
	// For example, delete the corresponding cookie
	Remove(writer http.ResponseWriter) error
}
