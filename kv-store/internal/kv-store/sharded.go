package kv_store

import (
	"kv-store/internal/interfaces"
	"sync"
)

type ShardedKVStore[K comparable, V any] struct {
	shards    []*KVStore[K, V]
	numShards uint64
}

func (S *ShardedKVStore[K, V]) getShard(key K) *KVStore[K, V] {
	hash := hashKey(key)
	shardIndex := hash % S.numShards
	return S.shards[shardIndex]
}

func (S *ShardedKVStore[K, V]) Get(key K) (V, bool) {
	return S.getShard(key).Get(key)
}

func (S *ShardedKVStore[K, V]) Set(key K, value V) error {
	return S.getShard(key).Set(key, value)
}

func (S *ShardedKVStore[K, V]) Delete(key K) error {
	return S.getShard(key).Delete(key)
}

func (S *ShardedKVStore[K, V]) Clear() {
	var wg sync.WaitGroup
	for _, shard := range S.shards {
		wg.Add(1)
		go func(s *KVStore[K, V]) {
			defer wg.Done()
			s.Clear()
		}(shard)
	}
	wg.Wait()
}

func (S *ShardedKVStore[K, V]) Size() uint64 {
	var total uint64
	for _, shard := range S.shards {
		total += shard.Size()
	}
	return total
}

func NewSharded[K comparable, V any](numShards, capacityPerShard uint64, factory interfaces.LinkedListFactory[K, V]) *ShardedKVStore[K, V] {
	shards := make([]*KVStore[K, V], numShards)
	for i := uint64(0); i < numShards; i++ {
		shards[i] = New[K, V](capacityPerShard, factory)
	}

	return &ShardedKVStore[K, V]{
		shards:    shards,
		numShards: numShards,
	}
}
