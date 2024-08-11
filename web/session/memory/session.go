package memory

import (
	"WebFrame/web/session"
	"context"
	"errors"
	cache "github.com/patrickmn/go-cache"
	"time"
)

type Store struct {
	c          *cache.Cache
	expiration time.Duration
}

func NewStore(expiration time.Duration) *Store {
	s := &Store{
		c: cache.New(expiration, time.Second),
	}
	return s
}

func (s *Store) Generate(ctx context.Context, id string) (session.Session, error) {
	sess := &memorySession{
		id:   id,
		data: make(map[string]string),
	}
	s.c.Set(sess.ID(), sess, s.expiration)
	return sess, nil
}

func (s *Store) Refresh(ctx context.Context, id string) error {
	sess, err := s.Get(ctx, id)
	if err != nil {
		return nil
	}
	s.c.Set(sess.ID(), sess, s.expiration)
	return nil
}

func (s *Store) Remove(ctx context.Context, id string) error {
	s.c.Delete(id)
	return nil
}

func (s *Store) Get(ctx context.Context, id string) (session.Session, error) {
	sess, ok := s.c.Get(id)
	if !ok {
		return nil, errors.New("session not found")
	}
	return sess.(*memorySession), nil
}

type memorySession struct {
	id         string
	data       map[string]string
	expiration time.Duration
}

func (m *memorySession) Get(ctx context.Context, key string) (string, error) {
	val, ok := m.data[key]
	if !ok {
		return "", errors.New("cannot find the key in the session")
	}
	return val, nil
}

func (m *memorySession) Set(ctx context.Context, key string, val string) error {
	m.data[key] = val
	return nil
}

func (m *memorySession) ID() string {
	return m.id
}
