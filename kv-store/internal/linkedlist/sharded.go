package linkedlist

import (
	"fmt"
	"hash/fnv"

	"kv-store/internal/interfaces"
)

const DefaultShardCount = 16

type ShardedLinkedList[T comparable] struct {
	shards    []interfaces.ILinkedList[T]
	shardMask uint64
}

func NewGenericShardedLinkedList[T comparable](shardCount int, factory interfaces.LinkedListFactory[T]) *ShardedLinkedList[T] {
	if shardCount <= 0 {
		shardCount = DefaultShardCount
	}

	shardCount = nextPowerOf2(shardCount)

	shards := make([]interfaces.ILinkedList[T], shardCount)
	for i := 0; i < shardCount; i++ {
		shards[i] = factory()
	}

	return &ShardedLinkedList[T]{
		shards:    shards,
		shardMask: uint64(shardCount - 1),
	}
}

func NewShardedLinkedList[T comparable](shardCount int) *ShardedLinkedList[T] {
	factory := func() interfaces.ILinkedList[T] {
		return NewLinkedList[T]()
	}
	return NewGenericShardedLinkedList(shardCount, factory)
}

func NewShardedAtomicLinkedList[T comparable](shardCount int) *ShardedLinkedList[T] {
	factory := func() interfaces.ILinkedList[T] {
		return NewAtomicLinkedList[T]()
	}
	return NewGenericShardedLinkedList(shardCount, factory)
}

func (sl *ShardedLinkedList[T]) getShard(key T) interfaces.ILinkedList[T] {
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

func (sl *ShardedLinkedList[T]) Get(key T) (T, bool) {
	shard := sl.getShard(key)
	return shard.Get(key)
}

func (sl *ShardedLinkedList[T]) Set(key T, value T) error {
	shard := sl.getShard(key)
	return shard.Set(key, value)
}

func (sl *ShardedLinkedList[T]) Delete(key T) error {
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
