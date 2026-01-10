package benchmarks

import (
	"fmt"
	"math/rand"
	"testing"

	"kv-store/internal/interfaces"
	"kv-store/internal/linkedlist"
)

const (
	BenchmarkSize     = 10000
	ConcurrentWorkers = 16
	WorkloadMixed     = 0.8 // 80% reads, 20% writes
)

func benchmarkSet[T interfaces.ILinkedList[int]](b *testing.B, list T, size int) {
	keys := make([]int, size)
	for i := 0; i < size; i++ {
		keys[i] = rand.Intn(size)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := keys[i%size]
		list.Set(key, key*2)
	}
}

func benchmarkGet[T interfaces.ILinkedList[int]](b *testing.B, list T, size int) {
	// Pre-populate the list
	for i := 0; i < size; i++ {
		list.Set(i, i*2)
	}

	keys := make([]int, size)
	for i := 0; i < size; i++ {
		keys[i] = rand.Intn(size)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := keys[i%size]
		list.Get(key)
	}
}

func benchmarkMixed[T interfaces.ILinkedList[int]](b *testing.B, list T, size int) {
	// Pre-populate the list
	for i := 0; i < size/2; i++ {
		list.Set(i, i*2)
	}

	keys := make([]int, size)
	operations := make([]bool, size) // true for read, false for write
	for i := 0; i < size; i++ {
		keys[i] = rand.Intn(size)
		operations[i] = rand.Float64() < WorkloadMixed
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := keys[i%size]
		if operations[i%size] {
			list.Get(key)
		} else {
			list.Set(key, key*2)
		}
	}
}

func benchmarkConcurrentMixed[T interfaces.ILinkedList[int]](b *testing.B, list T, size int) {
	// Pre-populate the list
	for i := 0; i < size/2; i++ {
		list.Set(i, i*2)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			key := rand.Intn(size)
			if rand.Float64() < WorkloadMixed {
				list.Get(key)
			} else {
				list.Set(key, key*2)
			}
		}
	})
}

// Basic LinkedList Benchmarks

func BenchmarkBasicLinkedList_Set(b *testing.B) {
	list := linkedlist.NewLinkedList[int]()
	benchmarkSet(b, list, BenchmarkSize)
}

func BenchmarkBasicLinkedList_Get(b *testing.B) {
	list := linkedlist.NewLinkedList[int]()
	benchmarkGet(b, list, BenchmarkSize)
}

func BenchmarkBasicLinkedList_Mixed(b *testing.B) {
	list := linkedlist.NewLinkedList[int]()
	benchmarkMixed(b, list, BenchmarkSize)
}

func BenchmarkBasicLinkedList_ConcurrentMixed(b *testing.B) {
	list := linkedlist.NewLinkedList[int]()
	benchmarkConcurrentMixed(b, list, BenchmarkSize)
}

// Atomic LinkedList Benchmarks

func BenchmarkAtomicLinkedList_Set(b *testing.B) {
	list := linkedlist.NewAtomicLinkedList[int]()
	benchmarkSet(b, list, BenchmarkSize)
}

func BenchmarkAtomicLinkedList_Get(b *testing.B) {
	list := linkedlist.NewAtomicLinkedList[int]()
	benchmarkGet(b, list, BenchmarkSize)
}

func BenchmarkAtomicLinkedList_Mixed(b *testing.B) {
	list := linkedlist.NewAtomicLinkedList[int]()
	benchmarkMixed(b, list, BenchmarkSize)
}

func BenchmarkAtomicLinkedList_ConcurrentMixed(b *testing.B) {
	list := linkedlist.NewAtomicLinkedList[int]()
	benchmarkConcurrentMixed(b, list, BenchmarkSize)
}

// Sharded LinkedList Benchmarks

func BenchmarkShardedLinkedList_Set(b *testing.B) {
	list := linkedlist.NewShardedLinkedList[int](16)
	benchmarkSet(b, list, BenchmarkSize)
}

func BenchmarkShardedLinkedList_Get(b *testing.B) {
	list := linkedlist.NewShardedLinkedList[int](16)
	benchmarkGet(b, list, BenchmarkSize)
}

func BenchmarkShardedLinkedList_Mixed(b *testing.B) {
	list := linkedlist.NewShardedLinkedList[int](16)
	benchmarkMixed(b, list, BenchmarkSize)
}

func BenchmarkShardedLinkedList_ConcurrentMixed(b *testing.B) {
	list := linkedlist.NewShardedLinkedList[int](16)
	benchmarkConcurrentMixed(b, list, BenchmarkSize)
}

// Sharded Atomic LinkedList Benchmarks

func BenchmarkShardedAtomicLinkedList_Set(b *testing.B) {
	list := linkedlist.NewShardedAtomicLinkedList[int](16)
	benchmarkSet(b, list, BenchmarkSize)
}

func BenchmarkShardedAtomicLinkedList_Get(b *testing.B) {
	list := linkedlist.NewShardedAtomicLinkedList[int](16)
	benchmarkGet(b, list, BenchmarkSize)
}

func BenchmarkShardedAtomicLinkedList_Mixed(b *testing.B) {
	list := linkedlist.NewShardedAtomicLinkedList[int](16)
	benchmarkMixed(b, list, BenchmarkSize)
}

func BenchmarkShardedAtomicLinkedList_ConcurrentMixed(b *testing.B) {
	list := linkedlist.NewShardedAtomicLinkedList[int](16)
	benchmarkConcurrentMixed(b, list, BenchmarkSize)
}

// Comparative benchmarks with different shard counts

func BenchmarkShardedLinkedList_DifferentShardCounts(b *testing.B) {
	shardCounts := []int{1, 2, 4, 8, 16, 32, 64}

	for _, shardCount := range shardCounts {
		b.Run(fmt.Sprintf("Shards_%d", shardCount), func(b *testing.B) {
			list := linkedlist.NewShardedLinkedList[int](shardCount)
			benchmarkConcurrentMixed(b, list, BenchmarkSize)
		})
	}
}

func BenchmarkShardedAtomicLinkedList_DifferentShardCounts(b *testing.B) {
	shardCounts := []int{1, 2, 4, 8, 16, 32, 64}

	for _, shardCount := range shardCounts {
		b.Run(fmt.Sprintf("Shards_%d", shardCount), func(b *testing.B) {
			list := linkedlist.NewShardedAtomicLinkedList[int](shardCount)
			benchmarkConcurrentMixed(b, list, BenchmarkSize)
		})
	}
}
