package interfaces

type IKVStore[K comparable, V comparable] interface {
	Get(key K) (V, bool)
	Set(key K, value V) error
	Delete(key K) error
	Clear()
	Size() int
	Keys() []K
}
