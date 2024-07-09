package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/burenotti/redis_impl/internal/domain/cmd"
)

const (
	defaultUnlockTimeout = 5 * time.Second
)

var (
	ErrNestedMulti         = fmt.Errorf("nested MULTI calls is not supported")
	ErrDiscardWithoutMulti = fmt.Errorf("discard without multi")
	ErrExecWithoutMulti    = fmt.Errorf("exec without multi")
)

func NewClient(service *RedisService) *Client {
	return &Client{
		service:        service,
		queuedCommands: nil,
		watches:        make(map[string]uint64),
		inProgress:     false,
	}
}

type Client struct {
	service        *RedisService
	queuedCommands []cmd.Command
	watches        map[string]uint64
	inProgress     bool
}

func (c *Client) Storage() cmd.Storage {
	return c.service.Storage()
}

func (c *Client) Run(ctx context.Context, command cmd.Command) (res *cmd.Result, err error) {
	if c.inProgress && !command.IsTx() {
		c.queuedCommands = append(c.queuedCommands, command)
		return cmd.NewResult("QUEUED"), nil
	}

	err = c.service.Atomic(ctx, func(ctx context.Context) error {
		res, err = command.Execute(ctx, c)
		if err == nil && command.IsModifying() {
			err = c.service.WalAppend(ctx, command)
		}
		return err
	})
	if err != nil {
		return cmd.NewResult(err), err
	}
	return res, nil
}

func (c *Client) StartTx(_ context.Context) error {
	if c.inProgress {
		return ErrNestedMulti
	}

	c.inProgress = true
	return nil
}

func (c *Client) ExecTx(ctx context.Context) (*cmd.Result, error) {
	if !c.inProgress {
		return cmd.EmptyResult(), ErrExecWithoutMulti
	}

	result := cmd.EmptyResult()

	for _, command := range c.queuedCommands {
		res, err := command.Execute(ctx, c)
		if err != nil {
			return res, err
		}

		result.Values = append(result.Values, res.Values)
		if err = c.service.WalAppend(ctx, command); err != nil {
			return res, err
		}
	}

	c.queuedCommands = c.queuedCommands[:0]
	c.inProgress = false
	_ = c.Unwatch(ctx)
	return result, nil
}

func (c *Client) DiscardTx(_ context.Context) error {
	if !c.inProgress {
		return ErrDiscardWithoutMulti
	}
	clear(c.queuedCommands)
	c.inProgress = false
	return nil
}

func (c *Client) Unwatch(_ context.Context) error {
	clear(c.watches)
	return nil
}

func (c *Client) Watch(ctx context.Context, keys ...string) error {
	return c.service.Atomic(ctx, func(ctx context.Context) error {
		for _, key := range keys {
			entry, err := c.service.Storage().Get(ctx, key)
			if errors.Is(err, cmd.ErrKeyNotFound) {
				c.watches[key] = 0
			}
			if err != nil {
				return err
			}
			c.watches[key] = entry.Revision()
		}

		return nil
	})
}
