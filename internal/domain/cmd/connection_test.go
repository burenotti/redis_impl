package cmd

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type mockValue struct {
	value     interface{}
	expiresAt *time.Time
	revision  uint64
}

func (m *mockValue) Value() interface{} {
	return m.value
}

func (m *mockValue) ExpiresAt() *time.Time {
	return m.expiresAt
}

func (m *mockValue) Revision() uint64 {
	return m.revision
}

func TestPing(t *testing.T) {
	ctx := context.Background()
	storage := NewMockStorage(t)

	cmd := Ping()
	assert.Equal(t, PING, cmd.Name())
	res, err := cmd.Execute(ctx, storage)
	assert.NoError(t, err)
	assert.Equal(t, res, Pong)
}

func TestGet(t *testing.T) {
	ctx := context.Background()
	storage := NewMockStorage(t)
	firstName := &mockValue{value: []byte("artem")}
	lastName := &mockValue{value: []byte("burenin")}
	expected := NewResult([]byte("artem"), []byte("burenin"), NilString)

	storage.On("Get", ctx, "first_name").Return(firstName, nil).Once()
	storage.On("Get", ctx, "last_name").Return(lastName, nil).Once()
	storage.On("Get", ctx, "middle_name").Return(nil, ErrKeyNotFound).Once()

	cmd := Get("first_name", "last_name", "middle_name")
	res, err := cmd.Execute(ctx, storage)
	require.NoError(t, err)
	assert.Equal(t, expected, res)
}
