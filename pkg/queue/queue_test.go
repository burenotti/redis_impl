package queue

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestQueue(t *testing.T) {
	q := Of[int](1, 2, 3, 4)
	assert.Equal(t, uint64(4), q.Len())

	v, ok := q.Pop()
	assert.True(t, ok)
	assert.Equal(t, 1, v)
	assert.EqualValues(t, 3, q.Len())
	q.Push(5)
	assert.EqualValues(t, 4, q.Len())

	v, ok = q.Pop()
	assert.True(t, ok)
	assert.Equal(t, 2, v)

	v, ok = q.Pop()
	assert.True(t, ok)
	assert.Equal(t, 3, v)

	v, ok = q.Pop()
	assert.True(t, ok)
	assert.Equal(t, 4, v)

	v, ok = q.Pop()
	assert.True(t, ok)
	assert.Equal(t, 5, v)

	assert.EqualValues(t, 0, q.Len())
	v, ok = q.Pop()

	assert.False(t, ok)

	q.Push(10)

	v, ok = q.Peek()
	assert.EqualValues(t, 10, v)
	assert.True(t, ok)

	assert.NotPanics(t, func() {
		assert.Equal(t, 10, q.MustPop())

		_, ok := q.Peek()
		assert.False(t, ok)
	})

	assert.Panics(t, func() {
		q.MustPop()
	})
}
