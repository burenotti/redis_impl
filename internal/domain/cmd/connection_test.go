package cmd_test

import (
	"context"
	"testing"
	"time"

	"github.com/burenotti/redis_impl/internal/domain/cmd"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	t.Parallel()
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	storage := NewMockClient(ctl)

	ctx := context.Background()

	ping := cmd.Ping()
	assert.Equal(t, cmd.PING, ping.Name())
	res, err := ping.Execute(ctx, storage)
	assert.NoError(t, err)
	assert.Equal(t, res, cmd.ResultPong())
}

func TestGet(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	client := NewMockClient(ctl)
	storage := NewMockStorage(ctl)

	firstName := &mockValue{value: []byte("artem")}
	lastName := &mockValue{value: []byte("burenin")}
	expected := cmd.NewResult([]byte("artem"), []byte("burenin"), cmd.NilString())

	client.EXPECT().Storage().Return(storage)
	storage.EXPECT().Get(ctx, "first_name").Return(firstName, nil)
	storage.EXPECT().Get(ctx, "last_name").Return(lastName, nil)
	storage.EXPECT().Get(ctx, "middle_name").Return(nil, cmd.ErrKeyNotFound)

	get := cmd.Get("first_name", "last_name", "middle_name")
	res, err := get.Execute(ctx, client)
	require.NoError(t, err)
	assert.Equal(t, expected, res)
}
