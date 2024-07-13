package cmd

import (
	"context"
	"errors"
)

const (
	GET      = "GET"
	SET      = "SET"
	PING     = "PING"
	ECHO     = "ECHO"
	MULTI    = "MULTI"
	EXEC     = "EXEC"
	DISCARD  = "DISCARD"
	WATCH    = "WATCH"
	UNWATCH  = "UNWATCH"
	HELLO    = "HELLO"
	REPLCONF = "REPLCONF"
)

func NilString() []byte {
	return []byte(nil)
}

func NilArray() []interface{} {
	return []interface{}(nil)
}

var ErrInvalidOpt = errors.New("invalid option")

type Result struct {
	Values []interface{}
}

func NewResult(values ...interface{}) *Result {
	return &Result{Values: values}
}

func EmptyResult() *Result {
	return &Result{Values: []interface{}(nil)}
}

func OkResult() *Result {
	return NewResult("OK")
}

type Command interface {
	Name() string
	Execute(ctx context.Context, storage Client) (*Result, error)
	IsModifying() bool
	IsTx() bool
	Args() []interface{}
}

type baseCommand struct{}

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
