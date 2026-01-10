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

func benchmarkSet[T interfaces.ILinkedList[int, int]](b *testing.B, list T, size int) {
	keys := make([]int, size)
	for i := 0; i < size; i++ {
		keys[i] = rand.Intn(size)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := keys[i%size]
		_ = list.Set(key, key*2)
	}
}

func benchmarkGet[T interfaces.ILinkedList[int, int]](b *testing.B, list T, size int) {
	// Pre-populate the list
	for i := 0; i < size; i++ {
		_ = list.Set(i, i*2)
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

func benchmarkMixed[T interfaces.ILinkedList[int, int]](b *testing.B, list T, size int) {
	// Pre-populate the list
	for i := 0; i < size/2; i++ {
		_ = list.Set(i, i*2)
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
			_ = list.Set(key, key*2)
		}
	}
}

func benchmarkConcurrentRead[T interfaces.ILinkedList[int, int]](b *testing.B, list T, size int) {
	// Pre-populate the list
	for i := 0; i < size; i++ {
		_ = list.Set(i, i*2)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			key := rand.Intn(size)
			list.Get(key)
		}
	})
}

func benchmarkConcurrentWrite[T interfaces.ILinkedList[int, int]](b *testing.B, list T, size int) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			key := rand.Intn(size)
			_ = list.Set(key, key*2)
		}
	})
}

func benchmarkConcurrentMixed[T interfaces.ILinkedList[int, int]](b *testing.B, list T, size int) {
	// Pre-populate the list
	for i := 0; i < size/2; i++ {
		_ = list.Set(i, i*2)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			key := rand.Intn(size)
			if rand.Float64() < WorkloadMixed {
				list.Get(key)
			} else {
				_ = list.Set(key, key*2)
			}
		}
	})
}

// Locking LinkedList Benchmarks

func BenchmarkLockingLinkedList_Set(b *testing.B) {
	list := linkedlist.New[int, int]()
	benchmarkSet(b, list, BenchmarkSize)
}

func BenchmarkLockingLinkedList_Get(b *testing.B) {
	list := linkedlist.New[int, int]()
	benchmarkGet(b, list, BenchmarkSize)
}

func BenchmarkLockingLinkedList_Mixed(b *testing.B) {
	list := linkedlist.New[int, int]()
	benchmarkMixed(b, list, BenchmarkSize)
}

func BenchmarkLockingLinkedList_ConcurrentMixed(b *testing.B) {
	list := linkedlist.New[int, int]()
	benchmarkConcurrentMixed(b, list, BenchmarkSize)
}

func BenchmarkLockingLinkedList_ConcurrentRead(b *testing.B) {
	list := linkedlist.New[int, int]()
	benchmarkConcurrentRead(b, list, BenchmarkSize)
}

func BenchmarkLockingLinkedList_ConcurrentWrite(b *testing.B) {
	list := linkedlist.New[int, int]()
	benchmarkConcurrentWrite(b, list, BenchmarkSize)
}

// Atomic LinkedList Benchmarks

func BenchmarkLinkedList_Set(b *testing.B) {
	list := linkedlist.NewAtomic[int, int]()
	benchmarkSet(b, list, BenchmarkSize)
}

func BenchmarkLinkedList_Get(b *testing.B) {
	list := linkedlist.NewAtomic[int, int]()
	benchmarkGet(b, list, BenchmarkSize)
}

func BenchmarkLinkedList_Mixed(b *testing.B) {
	list := linkedlist.NewAtomic[int, int]()
	benchmarkMixed(b, list, BenchmarkSize)
}

func BenchmarkLinkedList_ConcurrentMixed(b *testing.B) {
	list := linkedlist.NewAtomic[int, int]()
	benchmarkConcurrentMixed(b, list, BenchmarkSize)
}

func BenchmarkLinkedList_ConcurrentRead(b *testing.B) {
	list := linkedlist.NewAtomic[int, int]()
	benchmarkConcurrentRead(b, list, BenchmarkSize)
}

func BenchmarkLinkedList_ConcurrentWrite(b *testing.B) {
	list := linkedlist.NewAtomic[int, int]()
	benchmarkConcurrentWrite(b, list, BenchmarkSize)
}

// Sharded LinkedList Benchmarks

func BenchmarkShardedLinkedList_Set(b *testing.B) {
	list := linkedlist.NewSharded[int, int](16)
	benchmarkSet(b, list, BenchmarkSize)
}

func BenchmarkShardedLinkedList_Get(b *testing.B) {
	list := linkedlist.NewSharded[int, int](16)
	benchmarkGet(b, list, BenchmarkSize)
}

func BenchmarkShardedLinkedList_Mixed(b *testing.B) {
	list := linkedlist.NewSharded[int, int](16)
	benchmarkMixed(b, list, BenchmarkSize)
}

func BenchmarkShardedLinkedList_ConcurrentMixed(b *testing.B) {
	list := linkedlist.NewSharded[int, int](16)
	benchmarkConcurrentMixed(b, list, BenchmarkSize)
}

func BenchmarkShardedLinkedList_ConcurrentRead(b *testing.B) {
	list := linkedlist.NewSharded[int, int](16)
	benchmarkConcurrentRead(b, list, BenchmarkSize)
}

func BenchmarkShardedLinkedList_ConcurrentWrite(b *testing.B) {
	list := linkedlist.NewSharded[int, int](16)
	benchmarkConcurrentWrite(b, list, BenchmarkSize)
}

// Sharded Atomic LinkedList Benchmarks

func BenchmarkShardedAtomicLinkedList_Set(b *testing.B) {
	list := linkedlist.NewShardedAtomic[int, int](16)
	benchmarkSet(b, list, BenchmarkSize)
}

func BenchmarkShardedAtomicLinkedList_Get(b *testing.B) {
	list := linkedlist.NewShardedAtomic[int, int](16)
	benchmarkGet(b, list, BenchmarkSize)
}

func BenchmarkShardedAtomicLinkedList_Mixed(b *testing.B) {
	list := linkedlist.NewShardedAtomic[int, int](16)
	benchmarkMixed(b, list, BenchmarkSize)
}

func BenchmarkShardedAtomicLinkedList_ConcurrentMixed(b *testing.B) {
	list := linkedlist.NewShardedAtomic[int, int](16)
	benchmarkConcurrentMixed(b, list, BenchmarkSize)
}

func BenchmarkShardedAtomicLinkedList_ConcurrentRead(b *testing.B) {
	list := linkedlist.NewShardedAtomic[int, int](16)
	benchmarkConcurrentRead(b, list, BenchmarkSize)
}

func BenchmarkShardedAtomicLinkedList_ConcurrentWrite(b *testing.B) {
	list := linkedlist.NewShardedAtomic[int, int](16)
	benchmarkConcurrentWrite(b, list, BenchmarkSize)
}

// Comparative benchmarks with different shard counts

func BenchmarkShardedLinkedList_DifferentShardCounts(b *testing.B) {
	shardCounts := []int{1, 2, 4, 8, 16, 32, 64}

	for _, shardCount := range shardCounts {
		b.Run(fmt.Sprintf("Shards_%d", shardCount), func(b *testing.B) {
			list := linkedlist.NewSharded[int, int](shardCount)
			benchmarkConcurrentMixed(b, list, BenchmarkSize)
		})
	}
}

func BenchmarkShardedAtomicLinkedList_DifferentShardCounts(b *testing.B) {
	shardCounts := []int{1, 2, 4, 8, 16, 32, 64}

	for _, shardCount := range shardCounts {
		b.Run(fmt.Sprintf("Shards_%d", shardCount), func(b *testing.B) {
			list := linkedlist.NewShardedAtomic[int, int](shardCount)
			benchmarkConcurrentMixed(b, list, BenchmarkSize)
		})
	}
}

// Comparative benchmarks under high contention

func BenchmarkHighContention_ShardedAtomicLinkedList(b *testing.B) {
	list := linkedlist.NewShardedAtomic[int, int](16)

	contestedKeyCount := ConcurrentWorkers / 8

	keys := make([]int, contestedKeyCount)
	for i := 0; i < contestedKeyCount; i++ {
		keys[i] = i
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			key := keys[rand.Intn(contestedKeyCount)]
			if rand.Float64() < WorkloadMixed {
				list.Get(key)
			} else {
				_ = list.Set(key, key*2)
			}
		}
	})
}

func BenchmarkHighContention_ShardedLinkedList(b *testing.B) {
	list := linkedlist.NewSharded[int, int](16)

	contestedKeyCount := ConcurrentWorkers / 8

	keys := make([]int, contestedKeyCount)
	for i := 0; i < contestedKeyCount; i++ {
		keys[i] = i
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			key := keys[rand.Intn(contestedKeyCount)]
			if rand.Float64() < WorkloadMixed {
				list.Get(key)
			} else {
				_ = list.Set(key, key*2)
			}
		}
	})
}
