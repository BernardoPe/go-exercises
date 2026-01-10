package benchmarks

import (
	"math/rand"
	"testing"

	"kv-store/internal/interfaces"
	"kv-store/internal/kv-store"
	"kv-store/internal/linkedlist"
)

const (
	BenchSize     = 1000
	BenchCapacity = 16
	ReadRatio     = 0.8
)

// Generic benchmark functions
func benchSet[T interfaces.IKVStore[int, int]](b *testing.B, store T, size int) {
	keys := make([]int, size)
	values := make([]int, size)
	for i := 0; i < size; i++ {
		keys[i] = rand.Intn(size * 10)
		values[i] = rand.Intn(size * 10)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = store.Set(keys[i%size], values[i%size])
	}
}

func benchGet[T interfaces.IKVStore[int, int]](b *testing.B, store T, size int) {
	for i := 0; i < size; i++ {
		_ = store.Set(i, i*2)
	}
	keys := make([]int, size)
	for i := 0; i < size; i++ {
		keys[i] = rand.Intn(size)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = store.Get(keys[i%size])
	}
}

func benchSetUpdate[T interfaces.IKVStore[int, int]](b *testing.B, store T, size int) {
	for i := 0; i < size; i++ {
		_ = store.Set(i, i*2)
	}
	keys := make([]int, size)
	values := make([]int, size)
	for i := 0; i < size; i++ {
		keys[i] = rand.Intn(size)
		values[i] = rand.Intn(size * 10)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = store.Set(keys[i%size], values[i%size])
	}
}

func benchMixed[T interfaces.IKVStore[int, int]](b *testing.B, store T, size int) {
	for i := 0; i < size/2; i++ {
		_ = store.Set(i, i*2)
	}
	keys := make([]int, size)
	values := make([]int, size)
	operations := make([]bool, size)
	for i := 0; i < size; i++ {
		keys[i] = rand.Intn(size)
		values[i] = rand.Intn(size * 10)
		operations[i] = rand.Float64() < ReadRatio
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if operations[i%size] {
			_, _ = store.Get(keys[i%size])
		} else {
			_ = store.Set(keys[i%size], values[i%size])
		}
	}
}

func benchConcurrentRead[T interfaces.IKVStore[int, int]](b *testing.B, store T, size int) {
	for i := 0; i < size; i++ {
		_ = store.Set(i, i*2)
	}

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = store.Get(rand.Intn(size))
		}
	})
}

func benchConcurrentWrite[T interfaces.IKVStore[int, int]](b *testing.B, store T, size int) {
	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			key := rand.Intn(size * 10)
			_ = store.Set(key, key*2)
		}
	})
}

func benchConcurrentMixed[T interfaces.IKVStore[int, int]](b *testing.B, store T, size int) {
	for i := 0; i < size/2; i++ {
		_ = store.Set(i, i*2)
	}

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			key := rand.Intn(size)
			if rand.Float64() < ReadRatio {
				_, _ = store.Get(key)
			} else {
				_ = store.Set(key, key*2)
			}
		}
	})
}

func NewAtomicStore() interfaces.IKVStore[int, int] {
	return kv_store.New[int, int](BenchCapacity, func() interfaces.ILinkedList[int, int] {
		return linkedlist.NewAtomic[int, int]()
	})
}

func NewLockingStore() interfaces.IKVStore[int, int] {
	return kv_store.New[int, int](BenchCapacity, func() interfaces.ILinkedList[int, int] {
		return linkedlist.New[int, int]()
	})
}

func BenchmarkLocking_Set(b *testing.B) {
	store := NewLockingStore()
	benchSet(b, store, BenchSize)
}

func BenchmarkLocking_Get(b *testing.B) {
	store := NewLockingStore()
	benchGet(b, store, BenchSize)
}

func BenchmarkLocking_SetUpdate(b *testing.B) {
	store := NewLockingStore()
	benchSetUpdate(b, store, 100)
}

func BenchmarkLocking_Mixed(b *testing.B) {
	store := NewLockingStore()
	benchMixed(b, store, BenchSize)
}

func BenchmarkLocking_ConcurrentRead(b *testing.B) {
	store := NewLockingStore()
	benchConcurrentRead(b, store, BenchSize)
}

func BenchmarkLocking_ConcurrentWrite(b *testing.B) {
	store := NewLockingStore()
	benchConcurrentWrite(b, store, BenchSize)
}

func BenchmarkLocking_ConcurrentMixed(b *testing.B) {
	store := NewLockingStore()
	benchConcurrentMixed(b, store, BenchSize)
}

func BenchmarkAtomic_Set(b *testing.B) {
	store := NewAtomicStore()
	benchSet(b, store, BenchSize)
}

func BenchmarkAtomic_Get(b *testing.B) {
	store := NewAtomicStore()
	benchGet(b, store, BenchSize)
}

func BenchmarkAtomic_SetUpdate(b *testing.B) {
	store := NewAtomicStore()
	benchSetUpdate(b, store, 100)
}

func BenchmarkAtomic_Mixed(b *testing.B) {
	store := NewAtomicStore()
	benchMixed(b, store, BenchSize)
}

func BenchmarkAtomic_ConcurrentRead(b *testing.B) {
	store := NewAtomicStore()
	benchConcurrentRead(b, store, BenchSize)
}

func BenchmarkAtomic_ConcurrentWrite(b *testing.B) {
	store := NewAtomicStore()
	benchConcurrentWrite(b, store, BenchSize)
}

func BenchmarkAtomic_ConcurrentMixed(b *testing.B) {
	store := NewAtomicStore()
	benchConcurrentMixed(b, store, BenchSize)
}

func NewShardedLockingStore() interfaces.IKVStore[int, int] {
	return kv_store.NewSharded[int, int](8, BenchCapacity/8, func() interfaces.ILinkedList[int, int] {
		return linkedlist.New[int, int]()
	})
}

func BenchmarkShardedLocking_Set(b *testing.B) {
	store := NewShardedLockingStore()
	benchSet(b, store, BenchSize)
}

func BenchmarkShardedLocking_Get(b *testing.B) {
	store := NewShardedLockingStore()
	benchGet(b, store, BenchSize)
}

func BenchmarkShardedLocking_SetUpdate(b *testing.B) {
	store := NewShardedLockingStore()
	benchSetUpdate(b, store, 100)
}

func BenchmarkShardedLocking_Mixed(b *testing.B) {
	store := NewShardedLockingStore()
	benchMixed(b, store, BenchSize)
}

func BenchmarkShardedLocking_ConcurrentRead(b *testing.B) {
	store := NewShardedLockingStore()
	benchConcurrentRead(b, store, BenchSize)
}

func BenchmarkShardedLocking_ConcurrentWrite(b *testing.B) {
	store := NewShardedLockingStore()
	benchConcurrentWrite(b, store, BenchSize)
}

func BenchmarkShardedLocking_ConcurrentMixed(b *testing.B) {
	store := NewShardedLockingStore()
	benchConcurrentMixed(b, store, BenchSize)
}

func BenchmarkShardedAtomic_Set(b *testing.B) {
	store := NewShardedStore()
	benchSet(b, store, BenchSize)
}

func BenchmarkShardedAtomic_Get(b *testing.B) {
	store := NewShardedStore()
	benchGet(b, store, BenchSize)
}

func BenchmarkShardedAtomic_SetUpdate(b *testing.B) {
	store := NewShardedStore()
	benchSetUpdate(b, store, 100)
}

func BenchmarkShardedAtomic_Mixed(b *testing.B) {
	store := NewShardedStore()
	benchMixed(b, store, BenchSize)
}

func BenchmarkShardedAtomic_ConcurrentRead(b *testing.B) {
	store := NewShardedStore()
	benchConcurrentRead(b, store, BenchSize)
}

func BenchmarkShardedAtomic_ConcurrentWrite(b *testing.B) {
	store := NewShardedStore()
	benchConcurrentWrite(b, store, BenchSize)
}

func BenchmarkShardedAtomic_ConcurrentMixed(b *testing.B) {
	store := NewShardedStore()
	benchConcurrentMixed(b, store, BenchSize)
}

func NewShardedStore() interfaces.IKVStore[int, int] {
	return kv_store.NewSharded[int, int](8, BenchCapacity/8, func() interfaces.ILinkedList[int, int] {
		return linkedlist.NewAtomic[int, int]()
	})
}

const (
	ContestedKeyCount = 4
)

func benchHighContention[T interfaces.IKVStore[int, int]](b *testing.B, store T, contestedKeys int) {
	keys := make([]int, contestedKeys)
	for i := 0; i < contestedKeys; i++ {
		keys[i] = i
		_ = store.Set(keys[i], i*2)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			key := keys[rand.Intn(contestedKeys)]
			if rand.Float64() < ReadRatio {
				_, _ = store.Get(key)
			} else {
				_ = store.Set(key, key*2)
			}
		}
	})
}

func BenchmarkHighContention_Locking(b *testing.B) {
	store := NewLockingStore()
	benchHighContention(b, store, ContestedKeyCount)
}

func BenchmarkHighContention_Atomic(b *testing.B) {
	store := NewAtomicStore()
	benchHighContention(b, store, ContestedKeyCount)
}

func BenchmarkHighContention_ShardedLocking(b *testing.B) {
	store := NewShardedLockingStore()
	benchHighContention(b, store, ContestedKeyCount)
}

func BenchmarkHighContention_ShardedAtomic(b *testing.B) {
	store := NewShardedStore()
	benchHighContention(b, store, ContestedKeyCount)
}
