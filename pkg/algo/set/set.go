package set

import "cmp"

type Less[T any] func(T, T) bool

type Iterator[T any] func(T) bool

type Item[T any] interface {
	Less(other T) bool
}

type SortedSet[T any] struct {
	root *node[T]
	size int
	less Less[T]
}

func WithLess[T any](less Less[T], items ...T) *SortedSet[T] {
	s := SortedSet[T]{
		root: nil,
		size: 0,
		less: less,
	}
	for _, item := range items {
		s.Add(item)
	}
	return &s
}

func OfItems[T Item[T]](items ...T) *SortedSet[T] {
	less := func(a, b T) bool {
		return a.Less(b)
	}
	return WithLess[T](less, items...)
}

func Of[T cmp.Ordered](items ...T) *SortedSet[T] {
	less := func(a, b T) bool {
		return a < b
	}
	return WithLess[T](less, items...)
}

func (s *SortedSet[T]) Add(val T) {
	s.size++
	s.root = insertNode(s.root, val, s.less)
}

func (s *SortedSet[T]) Remove(val T) bool {
	if find(s.root, val, s.less) == nil {
		return false
	}
	s.root = deleteNode(s.root, val, s.less)
	s.size--
	return true
}

func (s *SortedSet[T]) Min() (min T, ok bool) {
	if s.root == nil {
		return
	}
	return findMin(s.root).val, true
}

func (s *SortedSet[T]) Max() (max T, ok bool) {
	if s.root == nil {
		return
	}
	return findMax(s.root).val, true
}

func (s *SortedSet[T]) Has(val T) bool {
	if n := find(s.root, val, s.less); n != nil {
		return true
	}
	return false
}

func (s *SortedSet[T]) Size() int {
	return s.size
}

func (s *SortedSet[T]) Ascend(iter Iterator[T]) {
	visitAscend(s.root, iter)
}

func (s *SortedSet[T]) Descend(iter Iterator[T]) {
	visitDescend(s.root, iter)
}

type node[T any] struct {
	val    T
	left   *node[T]
	right  *node[T]
	height int
}

func height[T any](n *node[T]) int {
	if n == nil {
		return 0
	}
	return n.height
}

// Creates a new node structure.
func newNode[T any](val T) *node[T] {
	return &node[T]{
		val:    val,
		left:   nil,
		right:  nil,
		height: 1,
	}
}

func rightRotate[T any](y *node[T]) *node[T] {
	x := y.left
	T2 := x.right
	x.right = y
	y.left = T2
	y.height = max(height(y.left), height(y.right)) + 1
	x.height = max(height(x.left), height(x.right)) + 1
	return x
}

// Performs a left rotation on the node.
func leftRotate[T any](x *node[T]) *node[T] {
	y := x.right
	T2 := y.left
	y.left = x
	x.right = T2
	x.height = max(height(x.left), height(x.right)) + 1
	y.height = max(height(y.left), height(y.right)) + 1
	return y
}

func balance[T any](n *node[T]) int {
	if n == nil {
		return 0
	}
	return height(n.left) - height(n.right)
}

// Inserts a new node into the AVL Tree.
func insertNode[T any](node *node[T], val T, less Less[T]) *node[T] {
	if node == nil {
		return newNode(val)
	}
	if less(val, node.val) { //nolint:gocritic // can't be rewritten with switch
		node.left = insertNode(node.left, val, less)
	} else if less(node.val, val) {
		node.right = insertNode(node.right, val, less)
	} else {
		return node
	}

	node.height = 1 + max(height(node.left), height(node.right))
	balanceFactor := balance(node)

	if balanceFactor > 1 {
		if less(val, node.left.val) {
			return rightRotate(node)
		} else if less(node.left.val, val) {
			node.left = leftRotate(node.left)
			return rightRotate(node)
		}
	}

	if balanceFactor < -1 {
		if less(val, node.right.val) {
			return leftRotate(node)
		} else if less(node.right.val, val) {
			node.right = rightRotate(node.right)
			return leftRotate(node)
		}
	}

	return node
}

func findMin[T any](node *node[T]) *node[T] {
	current := node
	for current.left != nil {
		current = current.left
	}
	return current
}

func findMax[T any](node *node[T]) *node[T] {
	current := node
	for current.right != nil {
		current = current.right
	}
	return current
}

// Deletes a node from the AVL Tree.
func deleteNode[T any](root *node[T], val T, less Less[T]) *node[T] {
	// Searching node
	if root == nil {
		return root
	}
	if less(val, root.val) { //nolint:gocritic,nestif // gocritic proposes nonsense. complexity is required
		root.left = deleteNode(root.left, val, less)
	} else if less(root.val, val) {
		root.right = deleteNode(root.right, val, less)
	} else {
		if root.left == nil || root.right == nil {
			temp := root.left
			if temp == nil {
				temp = root.right
			}
			if temp == nil {
				root = nil
			} else {
				*root = *temp
			}
		} else {
			temp := findMin(root.right)
			root.val = temp.val
			root.right = deleteNode(root.right, temp.val, less)
		}
	}
	if root == nil {
		return root
	}
	root.height = 1 + max(height(root.left), height(root.right))
	balanceFactor := balance(root)

	if balanceFactor > 1 {
		if balance(root.left) >= 0 {
			return rightRotate(root)
		}
		root.left = leftRotate(root.left)
		return rightRotate(root)
	}
	if balanceFactor < -1 {
		if balance(root.right) <= 0 {
			return leftRotate(root)
		}
		root.right = rightRotate(root.right)
		return leftRotate(root)
	}
	return root
}

func find[T any](n *node[T], val T, less Less[T]) *node[T] {
	current := n
	for current != nil {
		//nolint:gocritic // gocritic proposes nonsense
		if less(val, current.val) {
			current = current.left
		} else if less(current.val, val) {
			current = current.right
		} else {
			return current
		}
	}
	return nil
}

func visitAscend[T any](n *node[T], iter Iterator[T]) bool {
	if n == nil {
		return true
	}

	if ok := visitAscend(n.left, iter); !ok {
		return false
	}
	if ok := iter(n.val); !ok {
		return false
	}
	return visitAscend(n.right, iter)
}

func visitDescend[T any](n *node[T], iter Iterator[T]) bool {
	if n == nil {
		return true
	}

	if ok := visitDescend(n.right, iter); !ok {
		return false
	}
	if ok := iter(n.val); !ok {
		return false
	}
	return visitDescend(n.left, iter)
}
