package cmd

import (
	"context"
)

func Ping() Command {
	return &ping{}
}

type ping struct {
	baseCommand
}

func (p *ping) Execute(ctx context.Context, storage Storage) (*Result, error) {
	return NewResult("PONG"), nil
}

func (p *ping) Name() string {
	return PING
}

type hello struct {
	baseCommand
}

func Hello() Command {
	return &hello{}
}

func (h *hello) Name() string {
	return HELLO
}
func (h *hello) Execute(ctx context.Context, storage Storage) (*Result, error) {
	return NewResult([]interface{}{
		"server", "redis",
		"version", "3.0.0",
		"proto", "2",
		"id", "1",
		"mode", "standalone",
		"role", "master",
		"modules", []interface{}{},
	}), nil
}
