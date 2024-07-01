package memory

import (
	"context"
	"github.com/burenotti/redis_impl/internal/domain/cmd"
	"time"
)

type ValueType int

type Entry struct {
	key       string
	value     interface{}
	revision  uint64
	expiresAt *time.Time
}

func (e *Entry) Key() string {
	return e.key
}

func (e *Entry) Value() interface{} {
	return e.value
}

func (e *Entry) ExpiresAt() *time.Time {
	return e.expiresAt
}

func (e *Entry) Revision() uint64 {
	return e.revision
}

type Storage struct {
	kv   map[string]*Entry
	lock chan struct{}
}

func New() *Storage {
	return &Storage{
		kv:   make(map[string]*Entry),
		lock: make(chan struct{}, 1),
	}
}

func (s *Storage) Lock(ctx context.Context) error {
	select {
	case s.lock <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Storage) Unlock(ctx context.Context) error {
	select {
	case <-s.lock:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Storage) Set(ctx context.Context, key string, value interface{}, expiresAt *time.Time) error {
	prevRev := s.revision(key)
	s.kv[key] = &Entry{
		key:       key,
		value:     value,
		revision:  prevRev + 1,
		expiresAt: expiresAt,
	}
	return nil
}

func (s *Storage) Get(ctx context.Context, key string) (cmd.Value, error) {
	e, ok := s.kv[key]
	if !ok {
		return nil, cmd.ErrKeyNotFound
	}
	return e, nil
}

func (s *Storage) Del(ctx context.Context, key string) error {
	if _, ok := s.kv[key]; !ok {
		return cmd.ErrKeyNotFound
	}
	delete(s.kv, key)
	return nil
}

func (s *Storage) revision(key string) uint64 {
	if val, ok := s.kv[key]; ok {
		return val.revision
	}
	return 0
}
