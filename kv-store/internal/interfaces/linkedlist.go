package interfaces

type Hashable interface {
	Hash() uint64
}

type ILinkedList[K comparable, V any] interface {
	Get(key K) (V, bool)
	Set(key K, value V) error
	Delete(key K) error
	ForEach(func(key K, value V) bool)
}

type LinkedListFactory[K comparable, V any] func() ILinkedList[K, V]
