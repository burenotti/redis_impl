package heap_test

import (
	"github.com/burenotti/redis_impl/pkg/heap"
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

func TestHeap_Pop(t *testing.T) {
	t.Parallel()
	expectedOrder := []int{2, 3, 4, 7, 10}

	h := heap.OfOrdered(10, 3, 7, 2, 4)

	assert.Equal(t, len(expectedOrder), h.Len())
	for _, item := range expectedOrder {
		assert.Equal(t, item, h.MustPop())
	}
	assert.Equal(t, 0, h.Len())
}

func TestHeap_Push(t *testing.T) {
	t.Parallel()
	values := []int{7, 1, -5, 10, 12, 3, 0}
	expectedOrder := append(make([]int, 0, len(values)), values...)

	sort.Slice(expectedOrder, func(i, j int) bool {
		return expectedOrder[i] < expectedOrder[j]
	})

	h := heap.OfOrdered[int]()
	assert.Equal(t, 0, h.Len())
	for _, value := range values {
		h.Push(value)
	}
	assert.Equal(t, len(expectedOrder), h.Len())
	for _, item := range expectedOrder {
		assert.Equal(t, item, h.MustPop())
	}
	assert.Equal(t, 0, h.Len())
}
