package handler

import (
	"fmt"
	"github.com/burenotti/redis_impl/internal/domain/cmd"
	"strings"
	"time"
)

func parseGet(args []interface{}) (cmd.Command, error) {
	parsed := make([]string, len(args))
	ok := true
	for i, arg := range args {
		if parsed[i], ok = asString(arg); !ok {
			return nil, fmt.Errorf("%w: all arguments must be strings", ErrSyntax)
		}
	}
	return cmd.Get(parsed...), nil
}

func parseSet(args []interface{}) (cmd.Command, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("%w: not enough arguments", ErrSyntax)
	}

	if len(args) > 6 {
		return nil, fmt.Errorf("%w: too many arguments", ErrSyntax)
	}

	opts := make([]cmd.SetOpt, 0, 3)
	key, ok := asString(args[0])
	if !ok {
		return nil, fmt.Errorf("%w: key must be a string", ErrSyntax)
	}
	value := args[1]

	for i := 2; i < len(args); i++ {
		val, ok := asString(args[i])
		if !ok {
			return nil, fmt.Errorf("%w: bad syntax of command set", ErrSyntax)
		}
		val = strings.ToUpper(val)
		var opt cmd.SetOpt
		switch val {
		case "NX":
			opt = cmd.IfExists(cmd.NotExists)
		case "XX":
			opt = cmd.IfExists(cmd.Exists)
		case "GET":
			opt = cmd.SetGetPrevious()
		case "KEEPTTL":
			opt = cmd.KeepTTL()
		case "EX", "PX", "EXAT", "PXAT":
			if i == len(args)-1 {
				return nil, fmt.Errorf("%w: need value for %s", ErrSyntax, val)
			}
			expiry, ok := args[i+1].(int64)
			if !ok {
				return nil, fmt.Errorf("%w: %s argument must be an integer", ErrSyntax, val)
			}
			switch val {
			case "EX":
				opt = cmd.TTL(time.Duration(expiry) * time.Second)
			case "PX":
				opt = cmd.TTL(time.Duration(expiry) * time.Millisecond)
			case "EXAT":
				opt = cmd.ExpiresAt(time.Unix(expiry, 0))
			case "PXAT":
				opt = cmd.ExpiresAt(time.UnixMilli(expiry))
			}
		default:
			return nil, fmt.Errorf("%w: invalid argument %d", ErrSyntax, i+1)
		}
		opts = append(opts, opt)
	}

	return cmd.Set(key, value, opts...)
}

func parseNoArgs(command cmd.Command, args []interface{}) (cmd.Command, error) {
	if len(args) > 0 {
		return nil, fmt.Errorf("%w: %s does not accept any arguments", ErrSyntax, command.Name())
	}
	return command, nil

}

func parsePing(args []interface{}) (cmd.Command, error) {
	return parseNoArgs(cmd.Ping(), args)
}

func parseDiscard(args []interface{}) (cmd.Command, error) {
	return parseNoArgs(cmd.Discard(), args)
}

func parseMulti(args []interface{}) (cmd.Command, error) {
	return parseNoArgs(cmd.Multi(), args)
}

func parseExec(args []interface{}) (cmd.Command, error) {
	return parseNoArgs(cmd.Exec(), args)
}

func parseUnwatch(args []interface{}) (cmd.Command, error) {
	return parseNoArgs(cmd.Unwatch(), args)
}

func parseWatch(args []interface{}) (cmd.Command, error) {
	parsed := make([]string, len(args))
	ok := true
	for i, arg := range args {
		if parsed[i], ok = asString(arg); !ok {
			return nil, fmt.Errorf("%w: all arguments must be strings", ErrSyntax)
		}
	}
	return cmd.Watch(parsed...), nil
}

func asString(i interface{}) (string, bool) {
	if bytes, ok := i.([]byte); ok {
		return string(bytes), true
	} else {
		return "", false
	}
}
