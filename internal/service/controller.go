package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/burenotti/redis_impl/internal/domain"
	"github.com/burenotti/redis_impl/internal/domain/cmd"
	"github.com/burenotti/redis_impl/internal/storage/memory"
	"time"
)

var (
	ErrNestedMulti         = fmt.Errorf("nested MULTI calls is not supported")
	ErrDiscardWithoutMulti = fmt.Errorf("discard without multi")
	ErrExecWithoutMulti    = fmt.Errorf("exec without multi")
)

func New(store *memory.Storage) *Controller {
	return &Controller{
		storage:        store,
		queuedCommands: nil,
		watches:        make(map[string]uint64),
		inProgress:     false,
	}
}

type Controller struct {
	storage        *memory.Storage
	queuedCommands []cmd.Command
	watches        map[string]uint64
	inProgress     bool
}

func (s *Controller) Set(ctx context.Context, key string, value interface{}, expiresAt *time.Time) error {
	return s.storage.Set(ctx, key, value, expiresAt)
}

func (s *Controller) Get(ctx context.Context, key string) (domain.Value, error) {
	return s.storage.Get(ctx, key)
}

func (s *Controller) Del(ctx context.Context, key string) error {
	return s.storage.Del(ctx, key)
}

func (s *Controller) Run(ctx context.Context, c cmd.Command) (res *cmd.Result, err error) {

	if s.inProgress && !c.IsTx() {
		s.queuedCommands = append(s.queuedCommands, c)
		return cmd.NewResult("QUEUED"), nil
	}

	err = atomic(ctx, s, func(ctx context.Context) error {
		res, err = c.Execute(ctx, s)
		return err
	})
	if err != nil {
		return cmd.NewResult(err), err
	}
	return res, nil
}

func (s *Controller) StartTx(ctx context.Context) error {
	if s.inProgress {
		return ErrNestedMulti
	}

	s.inProgress = true
	return nil
}

func (s *Controller) RunTx(ctx context.Context) (*cmd.Result, error) {

	if !s.inProgress {
		return cmd.EmptyResult(), ErrExecWithoutMulti
	}

	result := cmd.EmptyResult()

	for _, command := range s.queuedCommands {
		res, err := command.Execute(ctx, s)
		if err != nil {
			return res, err
		}
		result.Values = append(result.Values, res.Values)
	}

	s.inProgress = false
	_ = s.Unwatch(ctx)
	return result, nil
}

func (s *Controller) DiscardTx(ctx context.Context) error {
	if !s.inProgress {
		return ErrDiscardWithoutMulti
	}
	clear(s.queuedCommands)
	s.inProgress = false
	return nil
}

func (s *Controller) Unwatch(ctx context.Context) error {
	clear(s.watches)
	return nil
}

func (s *Controller) Watch(ctx context.Context, keys ...string) error {
	return atomic(ctx, s, func(ctx context.Context) error {
		for _, key := range keys {
			entry, err := s.storage.Get(ctx, key)
			if errors.Is(err, domain.ErrKeyNotFound) {
				s.watches[key] = 0
			}
			if err != nil {
				return err
			}
			s.watches[key] = entry.Revision()
		}

		return nil
	})
}

type atomicFunc func(context.Context) error

func atomic(ctx context.Context, c *Controller, f atomicFunc) error {
	atomicCtx, cancel := context.WithCancel(ctx)

	defer cancel()

	if err := c.storage.Lock(atomicCtx); err != nil {
		return err
	}

	defer func() {
		unlockCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := c.storage.Unlock(unlockCtx); err != nil {
			panic("failed to unlock controller")
		}
	}()

	return f(atomicCtx)
}
