package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/burenotti/redis_impl/internal/domain"
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

func (g get) Execute(ctx context.Context, storage Storage) (*Result, error) {
	result := make([]interface{}, 0, len(g.Keys))
	for _, key := range g.Keys {
		val, err := storage.Get(ctx, key)
		if err != nil {
			if errors.Is(err, domain.ErrKeyNotFound) {
				return NewResult([]byte(nil)), nil
			}
			return &Result{Values: result}, err
		}
		result = append(result, val.Value())
	}
	return &Result{Values: result}, nil
}

func (g get) Name() string {
	return GET
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
		if s.expiresAt != nil {
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

func (s *set) Execute(ctx context.Context, storage Storage) (*Result, error) {
	prev, err := storage.Get(ctx, s.key)
	keyNotFound := prev == nil
	if err != nil && !keyNotFound {
		return nil, err
	}

	if s.exists == NotExists && !keyNotFound {
		return nil, domain.ErrKeyExists
	}

	if s.exists == Exists && keyNotFound {
		return nil, domain.ErrKeyNotFound
	}

	var newExpiry *time.Time
	if s.keepTTL && prev != nil {
		newExpiry = prev.ExpiresAt()
	}
	newExpiry = s.expiresAt

	err = storage.Set(ctx, s.key, s.value, newExpiry)

	if err != nil {
		return nil, err
	}

	if s.get {
		return NewResult(prev), nil
	} else {
		return EmptyResult(), nil
	}
}
