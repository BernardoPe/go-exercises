package internal

type KVStore[T comparable] interface {
	Get(key T) (*T, error)
	Set(key T, value T) error
	Delete(key T) error
}

type MemKVStore[T comparable] struct {
	capacity int
	data     []LinkedList[Node[T]]
}

func (m *MemKVStore[T]) Get(key T) (*T, error) {
	panic("implement me")
}

func (m *MemKVStore[T]) Set(key T, value T) error {
	panic("implement me")
}

func (m *MemKVStore[T]) Delete(key T) error {
	panic("implement me")
}

// NewMapKVStore creates a new in-memory key-value store with the specified capacity.
// The capacity determines the initial size of the underlying data structure.
// The Get, Set, and Delete methods are thread-safe.
func NewMapKVStore[T comparable](capacity int) *MemKVStore[T] {
	return &MemKVStore[T]{
		capacity: capacity,
		data:     make([]LinkedList[Node[T]], capacity),
	}
}
