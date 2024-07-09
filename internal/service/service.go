package service

import (
	"context"
	"github.com/burenotti/redis_impl/internal/domain/cmd"
	"time"
)

type Storage interface {
	Set(ctx context.Context, key string, value interface{}, expiresAt *time.Time) (cmd.Entry, error)
	Get(ctx context.Context, key string) (cmd.Entry, error)
	Del(ctx context.Context, key string) (cmd.Entry, error)
}

type RedisService struct {
	lock      chan struct{}
	done      chan struct{}
	storage   Storage
	wal       chan []cmd.Command
	listeners map[string]chan []cmd.Command
}

func NewService(storage Storage, walSize int) *RedisService {
	s := &RedisService{
		storage: storage,
		wal:     make(chan []cmd.Command, walSize),
		lock:    make(chan struct{}, 1),
		done:    make(chan struct{}),
	}
	s.Run()
	return s
}

func (s *RedisService) Lock(ctx context.Context) error {
	select {
	case s.lock <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *RedisService) Unlock() {
	<-s.lock
}

func (s *RedisService) Storage() Storage {
	return s.storage
}

type atomicFunc func(context.Context) error

func (s *RedisService) Atomic(ctx context.Context, f atomicFunc) error {
	atomicCtx, cancel := context.WithCancel(ctx)

	defer cancel()

	if err := s.Lock(atomicCtx); err != nil {
		return err
	}

	defer s.Unlock()

	return f(atomicCtx)
}

func (s *RedisService) WalAppend(ctx context.Context, commands ...cmd.Command) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case s.wal <- commands:
		return nil
	}
}

func (s *RedisService) AddWalListener(name string, ch chan []cmd.Command) {
	s.listeners[name] = ch
}

func (s *RedisService) RemoveWalListener(name string) bool {
	_, ok := s.listeners[name]
	delete(s.listeners, name)
	return ok
}

func (s *RedisService) Run() {
	go func() {
		for {
			select {
			case <-s.done:
				return
			case c, ok := <-s.wal:
				for _, lis := range s.listeners {
					if !ok {
						close(lis)
					}
					select {
					case <-s.done:
						return
					case lis <- c:
					}
				}
			}
		}
	}()
}

func (s *RedisService) Stop() {
	close(s.done)
}
