package handler

import (
	"bufio"
	"context"
	"github.com/burenotti/redis_impl/pkg/resp"
	"io"
	"strings"
)

type Storage interface {
	Get(key string) (string, bool)
	Set(key string, value string)
}

type CommandHandler interface {
	Handle(ctx context.Context, cmd *Command) (resp.Value, error)
}

type CommandFunc func(ctx context.Context, cmd *Command) (resp.Value, error)

func (c CommandFunc) Handle(ctx context.Context, cmd *Command) (resp.Value, error) {
	return c(ctx, cmd)
}

type Command struct {
	Name string
	Args []resp.Value
}

type Handler struct {
	storage  Storage
	handlers map[string]CommandHandler
}

func New(store Storage) *Handler {
	h := &Handler{
		storage: store,
	}
	h.handlers = map[string]CommandHandler{
		"PING": CommandFunc(h.ping),
		"ECHO": CommandFunc(h.echo),
		"INFO": CommandFunc(h.info),
		"SET":  CommandFunc(h.set),
		"GET":  CommandFunc(h.get),
	}
	return h
}

func (h *Handler) Handle(ctx context.Context, req io.Reader, res io.Writer) error {
	reader := bufio.NewReader(req)
	cmd := resp.NullArray()

	if err := cmd.Unmarshal(reader); err != nil {
		return resp.Marshal(res, resp.Error(err.Error()))
	}

	arr, _ := cmd.Array()
	if len(arr) == 0 {
		return resp.Marshal(res, resp.Error("can't work with empty array"))
	}

	name, ok := arr[0].String()
	if !ok {
		return resp.Marshal(res, resp.Error("command name must not be empty"))
	}

	command := Command{
		Name: strings.ToUpper(name),
		Args: arr[1:],
	}

	handler, ok := h.handlers[command.Name]

	if !ok {
		return resp.Marshal(res, resp.Error("unknown command"))
	}

	result, err := handler.Handle(ctx, &command)
	if err != nil {
		return resp.Marshal(res, resp.Error("unexpected error"))
	}

	return resp.Marshal(res, result)
}
