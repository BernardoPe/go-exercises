package linkedlist

import (
	"fmt"
	"hash/fnv"

	"kv-store/internal/interfaces"
)

const DefaultShardCount = 16

type ShardedLinkedList[K comparable, V any] struct {
	shards    []interfaces.ILinkedList[K, V]
	shardMask uint64
}

func NewSharded[K comparable, V any](shardCount int) *ShardedLinkedList[K, V] {
	factory := func() interfaces.ILinkedList[K, V] {
		return New[K, V]()
	}
	return newGenericShardedLinkedList(shardCount, factory)
}

func NewShardedAtomic[K comparable, V any](shardCount int) *ShardedLinkedList[K, V] {
	factory := func() interfaces.ILinkedList[K, V] {
		return NewAtomic[K, V]()
	}
	return newGenericShardedLinkedList(shardCount, factory)
}

func newGenericShardedLinkedList[K comparable, V any](shardCount int, factory interfaces.LinkedListFactory[K, V]) *ShardedLinkedList[K, V] {
	if shardCount <= 0 {
		shardCount = DefaultShardCount
	}

	shardCount = nextPowerOf2(shardCount)

	shards := make([]interfaces.ILinkedList[K, V], shardCount)
	for i := 0; i < shardCount; i++ {
		shards[i] = factory()
	}

	return &ShardedLinkedList[K, V]{
		shards:    shards,
		shardMask: uint64(shardCount - 1),
	}
}

func (sl *ShardedLinkedList[K, V]) getShard(key K) interfaces.ILinkedList[K, V] {
	var shardIndex uint64

	if hashable, ok := any(key).(interfaces.Hashable); ok {
		shardIndex = hashable.Hash()
	} else {
		switch v := any(key).(type) {
		case string:
			h := fnv.New64a()
			h.Write([]byte(v))
			shardIndex = h.Sum64()
		case int:
			shardIndex = uint64(v)
		case int32:
			shardIndex = uint64(v)
		case int64:
			shardIndex = uint64(v)
		case uint32:
			shardIndex = uint64(v)
		case uint64:
			shardIndex = v
		default:
			h := fnv.New64a()
			h.Write([]byte(fmt.Sprint(v)))
			shardIndex = h.Sum64()
		}
	}

	return sl.shards[shardIndex&sl.shardMask]
}

func (sl *ShardedLinkedList[K, V]) ForEach(func(key K, value V) bool) {
	for _, shard := range sl.shards {
		shard.ForEach(func(key K, value V) bool {
			return true
		})
	}
}

func (sl *ShardedLinkedList[K, V]) Get(key K) (V, bool) {
	shard := sl.getShard(key)
	return shard.Get(key)
}

func (sl *ShardedLinkedList[K, V]) Set(key K, value V) error {
	shard := sl.getShard(key)
	return shard.Set(key, value)
}

func (sl *ShardedLinkedList[K, V]) Delete(key K) error {
	shard := sl.getShard(key)
	return shard.Delete(key)
}

func nextPowerOf2(n int) int {
	if n <= 1 {
		return 1
	}

	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n++

	return n
}
