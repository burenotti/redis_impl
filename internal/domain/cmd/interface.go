//go:generate mockgen -destination mock_test.go -package cmd_test . Client,Storage

package cmd

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrKeyNotFound = errors.New("key not found")
	ErrKeyExists   = errors.New("key already exists")
	ErrExpired     = fmt.Errorf("%w: expired", ErrKeyNotFound)
)

type Storage interface {
	Set(ctx context.Context, key string, value interface{}, expiresAt *time.Time) (Entry, error)
	Get(ctx context.Context, key string) (Entry, error)
	Del(ctx context.Context, key string) (Entry, error)
}

type Client interface {
	StartTx(ctx context.Context) error
	ExecTx(ctx context.Context) (*Result, error)
	DiscardTx(ctx context.Context) error
	Watch(ctx context.Context, keys ...string) error
	Unwatch(ctx context.Context) error
	Storage() Storage
}

type Entry interface {
	Value() interface{}
	ExpiresAt() *time.Time
	Revision() uint64
}
