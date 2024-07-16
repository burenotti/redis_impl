package memory_test

import (
	"context"
	"github.com/burenotti/redis_impl/internal/domain/cmd"
	"github.com/burenotti/redis_impl/internal/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestStorage_canSetValues(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	storage := memory.New()
	_, err := storage.Set(ctx, "first_name", "artem", nil)
	require.NoError(t, err)

	_, err = storage.Set(ctx, "last_name", "burenin", nil)
	require.NoError(t, err)

	value, err := storage.Get(ctx, "first_name")
	require.NoError(t, err)
	require.NotNil(t, value)
	assert.Equal(t, "artem", value.Value())
	assert.EqualValues(t, 1, value.Revision())
	assert.Equal(t, "first_name", value.(*memory.Entry).Key())
	assert.Nil(t, value.ExpiresAt())

	value, err = storage.Get(ctx, "last_name")
	require.NoError(t, err)
	assert.Equal(t, "burenin", value.Value())
	assert.EqualValues(t, 1, value.Revision())
	assert.Nil(t, value.ExpiresAt())

	value, err = storage.Get(ctx, "middle_name")
	assert.Nil(t, value)
	assert.ErrorIs(t, err, cmd.ErrKeyNotFound)
}

func TestStorage_canSetExpiringValues(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	storage := memory.New()

	ttl := 5 * time.Millisecond
	expiresAt := time.Now().Add(ttl)
	_, err := storage.Set(ctx, "last_name", "burenin", &expiresAt)
	require.NoError(t, err)

	time.Sleep(2 * ttl)
	_, err = storage.Get(ctx, "last_name")
	assert.ErrorIs(t, err, cmd.ErrExpired)
}

func TestStorage_canDeleteValues(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	storage := memory.New()

	_, _ = storage.Set(ctx, "first_name", "artem", nil)

	e, err := storage.Del(ctx, "last_name")
	assert.Nil(t, e)
	require.ErrorIs(t, cmd.ErrKeyNotFound, err)
	_, err = storage.Del(ctx, "first_name")
	require.NoError(t, err)
	_, err = storage.Del(ctx, "first_name")
	require.ErrorIs(t, err, cmd.ErrKeyNotFound)
}
