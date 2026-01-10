package kv_store

import (
	"fmt"
	"kv-store/internal/interfaces"
	"sync"
)

type KVStore[K comparable, V any] struct {
	buckets  []interfaces.ILinkedList[K, V]
	factory  interfaces.LinkedListFactory[K, V]
	mu       sync.RWMutex
	size     uint64
	capacity uint64
	resizing bool
}

func (S *KVStore[K, V]) Get(key K) (V, bool) {
	hash := hashKey(key)

	S.mu.RLock()
	index := hash % S.capacity
	list := S.buckets[index]
	S.mu.RUnlock()

	return list.Get(key)
}

func (S *KVStore[K, V]) Set(key K, value V) error {
	hash := hashKey(key)

	S.mu.RLock()
	index := hash % S.capacity
	list := S.buckets[index]
	_, exists := list.Get(key)
	err := list.Set(key, value)
	S.mu.RUnlock()

	if err != nil {
		return err
	}

	if !exists {
		S.mu.Lock()
		S.size++
		needsResize := !S.resizing && S.size > S.capacity*2
		if needsResize {
			S.resizing = true
			S.doResize()
			S.resizing = false
		}
		S.mu.Unlock()
	}

	return nil
}

func (S *KVStore[K, V]) Delete(key K) error {
	hash := hashKey(key)

	S.mu.RLock()
	index := hash % S.capacity
	list := S.buckets[index]
	_, exists := list.Get(key)

	if !exists {
		S.mu.RUnlock()
		return fmt.Errorf("key not found")
	}

	err := list.Delete(key)
	S.mu.RUnlock()

	if err != nil {
		return err
	}

	S.mu.Lock()
	S.size--
	S.mu.Unlock()

	return nil
}

func (S *KVStore[K, V]) Clear() {
	S.mu.Lock()
	defer S.mu.Unlock()

	for i := range S.buckets {
		S.buckets[i] = S.factory()
	}

	S.size = 0
}

func (S *KVStore[K, V]) Size() uint64 {
	S.mu.RLock()
	defer S.mu.RUnlock()

	return S.size
}

func (S *KVStore[K, V]) doResize() {
	if S.size <= S.capacity*2 {
		return
	}

	newCapacity := S.capacity * 2
	newBuckets := make([]interfaces.ILinkedList[K, V], newCapacity)

	for i := uint64(0); i < newCapacity; i++ {
		newBuckets[i] = S.factory()
	}

	for _, list := range S.buckets {
		list.ForEach(func(key K, value V) bool {
			hash := hashKey(key)
			index := hash % newCapacity
			_ = newBuckets[index].Set(key, value)
			return true
		})
	}

	S.buckets = newBuckets
	S.capacity = newCapacity
}

func hashKey[K comparable](key K) uint64 {
	var hash uint64
	if hashable, ok := any(key).(interfaces.Hashable); ok {
		hash = hashable.Hash()
	} else {
		switch v := any(key).(type) {
		case string:
			for i := 0; i < len(v); i++ {
				hash = 31*hash + uint64(v[i])
			}
		case int:
			hash = uint64(v)
		case int32:
			hash = uint64(v)
		case int64:
			hash = uint64(v)
		case uint32:
			hash = uint64(v)
		case uint64:
			hash = v
		default:
			str := fmt.Sprint(v)
			for i := 0; i < len(str); i++ {
				hash = 31*hash + uint64(str[i])
			}
		}
		if hash < 0 {
			hash = -hash
		}
	}
	return hash
}

func New[K comparable, V any](capacity uint64, factory interfaces.LinkedListFactory[K, V]) *KVStore[K, V] {
	buckets := make([]interfaces.ILinkedList[K, V], capacity)

	for i := uint64(0); i < capacity; i++ {
		buckets[i] = factory()
	}

	return &KVStore[K, V]{
		buckets:  buckets,
		capacity: capacity,
		factory:  factory,
	}
}
