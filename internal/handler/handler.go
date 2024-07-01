package handler

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/burenotti/redis_impl/internal/domain/cmd"
	"github.com/burenotti/redis_impl/internal/service"
	"github.com/burenotti/redis_impl/pkg/resp"
	"io"
	"strings"
)

var (
	ErrSyntax = errors.New("syntax error")
)

type Handler struct {
	createController func() *service.Controller
	commands         map[string]func([]interface{}) (cmd.Command, error)
}

func New(createController func() *service.Controller) *Handler {
	h := &Handler{
		createController: createController,
		commands: map[string]func([]interface{}) (cmd.Command, error){
			cmd.GET:     parseGet,
			cmd.SET:     parseSet,
			cmd.PING:    parsePing,
			cmd.MULTI:   parseMulti,
			cmd.EXEC:    parseExec,
			cmd.DISCARD: parseDiscard,
			cmd.WATCH:   parseWatch,
			cmd.UNWATCH: parseUnwatch,
			cmd.HELLO:   parseHello,
		},
	}
	return h
}

func (h *Handler) Handle(ctx context.Context, req io.Reader, res io.Writer) error {
	reader := bufio.NewReader(req)
	controller := h.createController()
	for {
		command, err := h.parseNextCommand(reader)
		if err != nil {
			if err := resp.Marshal(res, err.Error()); err != nil {
				return err
			}
			continue
		}

		result, err := controller.Run(ctx, command)

		if err != nil {
			if err := resp.Marshal(res, err.Error()); err != nil {
				return err
			}
			continue
		}

		if err := h.marshalResult(res, result); err != nil {
			return err
		}
	}
}

func (h *Handler) marshalResult(w io.Writer, result *cmd.Result) error {
	if len(result.Values) == 1 {
		return resp.Marshal(w, result.Values[0])
	}
	return resp.Marshal(w, result.Values)
}

func (h *Handler) parseNextCommand(r *bufio.Reader) (cmd.Command, error) {

	data, err := resp.Unmarshal(r)
	if err != nil {
		return nil, err
	}

	arr, ok := data.([]interface{})
	if !ok || len(arr) == 0 {
		return nil, errors.New("syntax error")
	}

	rawName, ok := arr[0].([]byte)
	if !ok || len(rawName) == 0 {
		return nil, errors.New("command name must not be empty string")
	}
	name := strings.ToUpper(string(rawName))

	parser, ok := h.commands[name]
	if !ok {
		return nil, fmt.Errorf("%w: unknown command %s", ErrSyntax, name)
	}
	return parser(arr[1:])
}
