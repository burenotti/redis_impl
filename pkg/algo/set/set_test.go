package set_test

import (
	"github.com/burenotti/redis_impl/pkg/algo/set"
	"github.com/stretchr/testify/assert"
	"slices"
	"sort"
	"testing"
)

func TestSortedSet_Min(t *testing.T) {
	t.Run("should return false if set is empty", func(t *testing.T) {
		s := set.Of[int]()
		_, ok := s.Min()
		assert.False(t, ok)
	})

	t.Run("should return min value", func(t *testing.T) {
		s := set.Of[int](6, 3, 1, 2, 4)
		v, ok := s.Min()
		assert.True(t, ok)
		assert.Equal(t, 1, v)
	})
}

func TestSortedSet_Max(t *testing.T) {
	t.Run("should return false if set is empty", func(t *testing.T) {
		s := set.Of[int]()
		_, ok := s.Max()
		assert.False(t, ok)
	})

	t.Run("should return max value", func(t *testing.T) {
		s := set.Of[int](6, 3, 1, 2, 4)
		v, ok := s.Max()
		assert.True(t, ok)
		assert.Equal(t, 6, v)
	})
}

func TestSortedSet_Add(t *testing.T) {
	s := set.Of[int](6, 3, 1, 2, 4)
	assert.True(t, s.Has(3))
	s.Add(3)
	assert.True(t, s.Has(3))

	assert.False(t, s.Has(7))
	s.Add(7)
	assert.True(t, s.Has(7))

	s.Add(-5)
	assert.True(t, s.Has(-5))

	s.Add(100)
	assert.True(t, s.Has(100))
}

func TestSortedSet_Ascend(t *testing.T) {
	items := []int{6, 3, 2, 100, 1, 4, -5, 0, 10}

	s := set.Of[int](items...)

	sort.Ints(items)

	t.Run("should traverse items in ascend order", func(t *testing.T) {
		i := 0
		s.Ascend(func(v int) bool {
			assert.Equal(t, v, items[i])
			i++
			return true
		})
		assert.Equal(t, len(items), i)
	})

	t.Run("should stop traverse if iterator returns false", func(t *testing.T) {
		i := 0
		s.Ascend(func(v int) bool {
			assert.Equal(t, v, items[i])
			i++
			return v != 6
		})
		assert.Equal(t, 7, i)
	})
}

func TestSortedSet_Descend(t *testing.T) {
	items := []int{6, 3, 2, 100, 1, 4, -5, 0, 10}

	s := set.Of[int](items...)

	slices.SortFunc(items, func(a, b int) int {
		return b - a
	})

	t.Run("should traverse items in ascend order", func(t *testing.T) {
		i := 0
		s.Descend(func(v int) bool {
			assert.Equal(t, v, items[i])
			i++
			return true
		})
		assert.Equal(t, len(items), i)
	})

	t.Run("should stop traverse if iterator returns false", func(t *testing.T) {
		i := 0
		s.Descend(func(v int) bool {
			assert.Equal(t, v, items[i])
			i++
			return v != 6
		})
		assert.Equal(t, 3, i)
	})
}

func TestSortedSet_Remove(t *testing.T) {
	items := []int{6, 3, 2, 100, 1, 4, -5, 0, 10}

	t.Run("should delete items", func(t *testing.T) {
		s := set.Of(items...)
		for i, item := range items {
			assert.Equal(t, len(items)-i, s.Size())
			assert.True(t, s.Has(item))
			assert.True(t, s.Remove(item))
			assert.False(t, s.Has(item))
		}

		assert.False(t, s.Remove(100))
	})

	t.Run("should not panic if item is not present in set", func(t *testing.T) {
		s := set.Of(items...)
		assert.False(t, s.Remove(200))
		assert.False(t, s.Remove(-100))
		assert.False(t, s.Remove(11))
		assert.Equal(t, len(items), s.Size())
		for _, item := range items {
			assert.True(t, s.Has(item))
		}
	})
}
