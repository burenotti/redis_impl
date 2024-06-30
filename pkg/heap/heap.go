package heap

import (
	"cmp"
	"container/heap"
)

type Lesser[T any] interface {
	Less(other T) bool
}

type Heap[T any] struct {
	data container[T]
}

func WithLess[T any](less func(T, T) bool, items ...T) *Heap[T] {
	c := container[T]{
		data: items,
		less: less,
	}
	heap.Init(&c)
	return &Heap[T]{
		data: c,
	}
}

func OfOrdered[T cmp.Ordered](items ...T) *Heap[T] {
	return WithLess(cmp.Less[T], items...)
}

func OfLessers[T Lesser[T]](items ...T) *Heap[T] {
	less := func(a T, b T) bool {
		return a.Less(b)
	}
	return WithLess(less, items...)
}

func (h *Heap[T]) Reserve(size int) {
	newSlice := make([]T, size)
	copy(newSlice, h.data.data)
	h.data.data = newSlice
}

func (h *Heap[T]) Push(value T) {
	heap.Push(&h.data, value)
}

func (h *Heap[T]) Pop() (T, bool) {
	if h.Len() == 0 {
		var null T
		return null, false
	}
	return heap.Pop(&h.data).(T), true
}

func (h *Heap[T]) MustPop() T {
	if val, ok := h.Pop(); ok {
		return val
	} else {
		panic("can't pop element: heap is empty")
	}
}

func (h *Heap[T]) Len() int {
	return h.data.Len()
}

type container[T any] struct {
	data []T
	less func(a, b T) bool
}

func (h *container[T]) Len() int { return len(h.data) }

func (h *container[T]) Less(i, j int) bool {
	return h.less(h.data[i], h.data[j])
}

func (h *container[T]) Swap(i, j int) {
	h.data[i], h.data[j] = h.data[j], h.data[i]
}

func (h *container[T]) Push(value any) {
	h.data = append(h.data, value.(T))
}

func (h *container[T]) Pop() any {
	var null T
	old := h.data
	n := len(old)
	item := old[n-1]
	old[n-1] = null
	h.data = old[0 : n-1]
	return item
}
