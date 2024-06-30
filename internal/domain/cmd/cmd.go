package cmd

import (
	"context"
	"errors"
	"github.com/burenotti/redis_impl/internal/domain"
	"time"
)

const (
	GET     = "GET"
	SET     = "SET"
	PING    = "PING"
	ECHO    = "ECHO"
	MULTI   = "MULTI"
	EXEC    = "EXEC"
	DISCARD = "DISCARD"
	WATCH   = "WATCH"
	UNWATCH = "UNWATCH"
)

type Storage interface {
	Set(ctx context.Context, key string, value interface{}, expiresAt *time.Time) error
	Get(ctx context.Context, key string) (domain.Value, error)
	Del(ctx context.Context, key string) error
	StartTx(ctx context.Context) error
	RunTx(ctx context.Context) (*Result, error)
	DiscardTx(ctx context.Context) error
	Watch(ctx context.Context, keys ...string) error
	Unwatch(ctx context.Context) error
}

var (
	ErrInvalidOpt = errors.New("invalid option")
)

type Result struct {
	Values []interface{}
}

func NewResult(values ...interface{}) *Result {
	return &Result{Values: values}
}

func EmptyResult() *Result {
	return &Result{Values: []interface{}(nil)}
}

type Command interface {
	Name() string
	Execute(ctx context.Context, storage Storage) (*Result, error)
	IsModifying() bool
	IsTx() bool
}

type baseCommand struct {
}

func (b *baseCommand) IsModifying() bool {
	return false
}

func (b *baseCommand) IsTx() bool {
	return false
}

type modifyingCommand struct {
	baseCommand
}

func (m *modifyingCommand) IsModifying() bool {
	return true
}

type txCommand struct {
	baseCommand
}

func (m *txCommand) IsTx() bool {
	return true
}
