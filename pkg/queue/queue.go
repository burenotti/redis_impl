package queue

type node[T any] struct {
	val  T
	next *node[T]
	prev *node[T]
}

func New[T any]() *Queue[T] {
	return &Queue[T]{}
}

func Of[T any](values ...T) *Queue[T] {
	q := New[T]()
	for _, v := range values {
		q.Push(v)
	}
	return q
}

type Queue[T any] struct {
	head *node[T]
	tail *node[T]
	len  uint64
}

func (q *Queue[T]) Len() uint64 {
	return q.len
}

func (q *Queue[T]) Push(value T) {
	if q.len == 0 {
		q.tail = &node[T]{
			val:  value,
			next: nil,
			prev: nil,
		}
		q.head = q.tail
		q.len++
		return
	}
	q.tail.next = &node[T]{
		val:  value,
		next: nil,
		prev: q.tail,
	}
	q.len++
	q.tail = q.tail.next
}

func (q *Queue[T]) Pop() (T, bool) {
	if q.len == 0 {
		return *new(T), false
	}

	head := q.head

	q.head = q.head.next
	q.len--
	return head.val, true
}

func (q *Queue[T]) MustPop() T {
	if v, ok := q.Pop(); ok {
		return v
	} else {
		panic("queue is empty")
	}
}

func (q *Queue[T]) Peek() (T, bool) {
	if q.len == 0 {
		return *new(T), false
	}
	return q.tail.val, true
}
