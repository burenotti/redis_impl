package handler

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/burenotti/redis_impl/internal/domain/cmd"
)

func parseGet(args []interface{}) (cmd.Command, error) {
	parsed := make([]string, len(args))
	for i, arg := range args {
		var ok bool
		if parsed[i], ok = asString(arg); !ok {
			return nil, fmt.Errorf("%w: all arguments must be strings", ErrSyntax)
		}
	}
	return cmd.Get(parsed...), nil
}

//nolint:funlen // parsing functions can be long
func parseSet(args []interface{}) (cmd.Command, error) {
	if len(args) < 2 { //nolint:mnd // min amount of arguments key, value
		return nil, fmt.Errorf("%w: not enough arguments", ErrSyntax)
	}

	if len(args) > 6 { //nolint:mnd // max amount of arguments
		return nil, fmt.Errorf("%w: too many arguments", ErrSyntax)
	}

	var opts []cmd.SetOpt
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
			i++
			expiry, err := parseInt(args[i])
			if err != nil {
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

func parseInt(arg interface{}) (int64, error) {
	switch v := arg.(type) {
	case int64:
		return v, nil
	case []byte:
		return strconv.ParseInt(string(v), 10, 64)
	case string:
		return strconv.ParseInt(v, 10, 6)
	default:
		return 0, strconv.ErrSyntax
	}
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
	for i, arg := range args {
		var ok bool
		if parsed[i], ok = asString(arg); !ok {
			return nil, fmt.Errorf("%w: all arguments must be strings", ErrSyntax)
		}
	}
	return cmd.Watch(parsed...), nil
}

//nolint:unparam // need to implement interface
func parseHello(_ []interface{}) (cmd.Command, error) {
	return cmd.Hello(), nil
}

func asString(i interface{}) (string, bool) {
	if bytes, ok := i.([]byte); ok {
		return string(bytes), true
	}
	return "", false
}
