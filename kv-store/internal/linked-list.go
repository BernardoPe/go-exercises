package internal

type Node[T comparable] struct {
	key   T
	value T
	next  *Node[T]
}

type LinkedList[T comparable] struct {
	head *Node[T]
	tail *Node[T]
	size int
}

func (l *LinkedList[T]) Get(key T) (*T, error) {
	panic("implement me")
}

func (l *LinkedList[T]) Set(key T, value T) error {
	panic("implement me")
}

func (l *LinkedList[T]) Delete(key T) error {
	panic("implement me")
}

func NewLinkedList[T comparable]() *LinkedList[T] {
	return &LinkedList[T]{}
}
