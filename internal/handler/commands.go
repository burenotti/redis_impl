package handler

import (
	"context"
	"github.com/burenotti/redis_impl/pkg/resp"
)

func (h *Handler) ping(ctx context.Context, cmd *Command) (resp.Value, error) {
	if len(cmd.Args) > 0 {
		return resp.Error("ping does not accept any arguments"), nil
	}

	return resp.SimpleString("PONG"), nil
}

func (h *Handler) echo(ctx context.Context, cmd *Command) (resp.Value, error) {
	if len(cmd.Args) > 1 {
		return resp.Array(cmd.Args), nil
	}
	return cmd.Args[0], nil
}

func (h *Handler) info(ctx context.Context, cmd *Command) (result resp.Value, err error) {
	return resp.BulkString([]byte("Уаааа, я эту хуйню целый день хуярил\n")), nil
}

func (h *Handler) get(ctx context.Context, cmd *Command) (result resp.Value, err error) {
	if len(cmd.Args) != 1 {
		return resp.Error("command accepts exactly one argument"), nil
	}
	key, ok := cmd.Args[0].String()
	if !ok {
		return resp.Error("key must be simple string"), nil
	}

	value, ok := h.storage.Get(key)
	if !ok {
		result = resp.NullArray()
	} else {
		result = resp.BulkString([]byte(value))
	}
	return result, nil
}

func (h *Handler) set(ctx context.Context, cmd *Command) (resp.Value, error) {
	if len(cmd.Args) != 2 {
		return resp.Error("command accepts exactly two arguments"), nil
	}

	key, ok := cmd.Args[0].String()
	if !ok {
		return resp.Error("key must be simple string"), nil
	}

	value, ok := cmd.Args[1].Bytes()
	if !ok {
		return resp.Error("value must be a string"), nil
	}
	h.storage.Set(key, string(value))
	return resp.SimpleString("OK"), nil
}
