package interfaces

type Hashable interface {
	Hash() uint64
}

type ILinkedList[T comparable] interface {
	Get(key T) (T, bool)
	Set(key T, value T) error
	Delete(key T) error
}

type LinkedListFactory[T comparable] func() ILinkedList[T]
