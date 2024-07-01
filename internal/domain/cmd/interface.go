//go:generate mockgen -source interface.go -destination mock_test.go -package cmd_test

package cmd

import (
	"context"
	"errors"
	"time"
)

var (
	ErrKeyNotFound = errors.New("key not found")
	ErrKeyExists   = errors.New("key already exists")
)

type Storage interface {
	Set(ctx context.Context, key string, value interface{}, expiresAt *time.Time) error
	Get(ctx context.Context, key string) (Value, error)
	Del(ctx context.Context, key string) error
	StartTx(ctx context.Context) error
	RunTx(ctx context.Context) (*Result, error)
	DiscardTx(ctx context.Context) error
	Watch(ctx context.Context, keys ...string) error
	Unwatch(ctx context.Context) error
}

type Value interface {
	Value() interface{}
	ExpiresAt() *time.Time
	Revision() uint64
}
