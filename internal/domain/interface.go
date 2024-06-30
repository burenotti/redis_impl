package domain

import (
	"errors"
	"time"
)

var (
	ErrKeyNotFound = errors.New("key not found")
	ErrKeyExists   = errors.New("key already exists")
)

type Value interface {
	Value() interface{}
	ExpiresAt() *time.Time
	Revision() uint64
}
