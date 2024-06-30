package cmd

import (
	"context"
)

func Multi() Command {
	return &multi{}
}

type multi struct {
	txCommand
}

func (m *multi) Name() string {
	return MULTI
}

func (m *multi) Execute(ctx context.Context, storage Storage) (*Result, error) {
	if err := storage.StartTx(ctx); err != nil {
		return NewResult(err), err
	}
	return NewResult("OK"), nil
}

type exec struct {
	txCommand
}

func Exec() Command {
	return &exec{}
}

func (e *exec) Name() string {
	return EXEC
}

func (e *exec) Execute(ctx context.Context, storage Storage) (*Result, error) {
	if res, err := storage.RunTx(ctx); err == nil {
		return res, nil
	} else {
		return NewResult(err), err
	}
}

type discard struct {
	txCommand
}

func Discard() Command {
	return &discard{}
}

func (d *discard) Name() string {
	return DISCARD
}

func (d *discard) Execute(ctx context.Context, storage Storage) (*Result, error) {
	if err := storage.DiscardTx(ctx); err != nil {
		return NewResult(err), err
	}
	return NewResult("OK"), nil
}

type watch struct {
	txCommand
	keys []string
}

func Watch(keys ...string) Command {
	return &watch{
		keys: keys,
	}
}

func (w *watch) Name() string {
	return WATCH
}

func (w *watch) Execute(ctx context.Context, storage Storage) (*Result, error) {
	if err := storage.Watch(ctx, w.keys...); err != nil {
		return NewResult(err), err
	}
	return NewResult("OK"), nil
}

type unwatch struct {
	txCommand
}

func Unwatch() Command {
	return &unwatch{}
}

func (u *unwatch) Name() string {
	return UNWATCH
}

func (u *unwatch) Execute(ctx context.Context, storage Storage) (*Result, error) {
	if err := storage.Unwatch(ctx); err != nil {
		return NewResult(err), err
	}
	return NewResult("OK"), nil
}
