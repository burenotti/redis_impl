package cmd

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func Get(keys ...string) Command {
	return &get{
		Keys: keys,
	}
}

type get struct {
	baseCommand
	Keys []string
}

func (g *get) Execute(ctx context.Context, c Client) (*Result, error) {
	result := make([]interface{}, 0, len(g.Keys))
	storage := c.Storage()
	for _, key := range g.Keys {
		val, err := storage.Get(ctx, key)
		if err != nil {
			if errors.Is(err, ErrKeyNotFound) {
				result = append(result, NilString())
				continue
			}
			return &Result{Values: result}, err
		}
		result = append(result, val.Value())
	}
	return &Result{Values: result}, nil
}

func (g *get) Name() string {
	return GET
}

func (g *get) Args() []interface{} {
	res := []interface{}{GET}
	for _, key := range g.Keys {
		res = append(res, key)
	}
	return res
}

type ExistsOpt string

const (
	NotExists ExistsOpt = "NX"
	Exists    ExistsOpt = "XX"
)

type SetOpt func(*set) error

func IfExists(opt ExistsOpt) SetOpt {
	return func(s *set) error {
		if s.exists != "" {
			return fmt.Errorf("%w: Exists opt provided more than once", ErrInvalidOpt)
		}
		s.exists = opt
		return nil
	}
}

func SetGetPrevious() SetOpt {
	return func(s *set) error {
		s.get = true
		return nil
	}
}

func KeepTTL() SetOpt {
	return func(s *set) error {
		s.keepTTL = true
		return nil
	}
}

func ExpiresAt(exp time.Time) SetOpt {
	return func(s *set) error {
		if !s.expiresAt.IsZero() {
			return fmt.Errorf("%w: expiration provided more than once", ErrInvalidOpt)
		}
		s.expiresAt = &exp
		return nil
	}
}

func TTL(ttl time.Duration) SetOpt {
	return ExpiresAt(time.Now().Add(ttl))
}

func Set(key string, value interface{}, opts ...SetOpt) (Command, error) {
	s := &set{
		key:   key,
		value: value,
	}

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}
	return s, nil
}

type set struct {
	modifyingCommand
	key       string
	value     interface{}
	exists    ExistsOpt
	expiresAt *time.Time
	get       bool
	keepTTL   bool
}

func (s *set) Name() string {
	return SET
}

func (s *set) Execute(ctx context.Context, c Client) (*Result, error) {
	prev, err := c.Storage().Get(ctx, s.key)
	keyNotFound := prev == nil
	if err != nil && !keyNotFound {
		return nil, err
	}

	if s.exists == NotExists && !keyNotFound {
		return nil, ErrKeyExists
	}

	if s.exists == Exists && keyNotFound {
		return nil, ErrKeyNotFound
	}
	newExpiry := s.expiresAt
	if s.keepTTL && prev != nil {
		newExpiry = prev.ExpiresAt()
	}

	_, err = c.Storage().Set(ctx, s.key, s.value, newExpiry)
	if err != nil {
		return nil, err
	}

	if s.get {
		return NewResult(prev), nil
	}

	return OkResult(), nil
}

func (s *set) Args() []interface{} {
	res := []interface{}{SET, s.key, s.value}

	if s.get {
		res = append(res, "GET")
	}

	if s.exists == NotExists {
		res = append(res, "NX")
	}
	if s.exists == Exists {
		res = append(res, "XX")
	}

	if s.expiresAt != nil {
		res = append(res, "PXAT", s.expiresAt.Unix())
	}

	if s.keepTTL {
		res = append(res, "KEEPTTL")
	}

	return res
}
