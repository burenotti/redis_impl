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
