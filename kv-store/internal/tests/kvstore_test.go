package tests

import (
	"fmt"
	"kv-store/internal/interfaces"
	"kv-store/internal/kv-store"
	"kv-store/internal/linkedlist"
	"math/rand"
	"sync"
	"testing"
)

func newLockingStore[K comparable, V any](capacity int) interfaces.IKVStore[K, V] {
	factory := func() interfaces.ILinkedList[K, V] {
		return linkedlist.New[K, V]()
	}
	return kv_store.New[K, V](uint64(capacity), factory)
}

func newAtomicStore[K comparable, V any](capacity int) interfaces.IKVStore[K, V] {
	factory := func() interfaces.ILinkedList[K, V] {
		return linkedlist.NewAtomic[K, V]()
	}
	return kv_store.New[K, V](uint64(capacity), factory)
}

func newShardedStore[K comparable, V any](capacity int) interfaces.IKVStore[K, V] {
	factory := func() interfaces.ILinkedList[K, V] {
		return linkedlist.NewAtomic[K, V]()
	}
	return kv_store.NewSharded[K, V](8, uint64(capacity/8), factory)
}

func assertNoError(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", msg, err)
	}
}

func assertEqual[T comparable](t *testing.T, got, want T, msg string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: expected %v, got %v", msg, want, got)
	}
}

func assertTrue(t *testing.T, condition bool, msg string) {
	t.Helper()
	if !condition {
		t.Errorf("%s", msg)
	}
}

func TestNew(t *testing.T) {
	store := newLockingStore[int, int](10)
	assertTrue(t, store != nil, "Factory returned nil")
	assertEqual(t, store.Size(), uint64(0), "size")
}

func TestSetAndGet(t *testing.T) {
	store := newLockingStore[string, int](10)

	assertNoError(t, store.Set("key1", 100), "Set failed")

	value, found := store.Get("key1")
	assertTrue(t, found, "Get() did not find the key")
	assertEqual(t, value, 100, "value")

	_, found = store.Get("nonexistent")
	assertTrue(t, !found, "Get() found a non-existent key")
}

func TestSetMultiple(t *testing.T) {
	store := newLockingStore[int, string](5)
	testData := map[int]string{1: "one", 2: "two", 3: "three", 4: "four", 5: "five"}

	for key, value := range testData {
		assertNoError(t, store.Set(key, value), fmt.Sprintf("Set(%d, %s)", key, value))
	}

	for key, expectedValue := range testData {
		value, found := store.Get(key)
		assertTrue(t, found, fmt.Sprintf("Get(%d) did not find the key", key))
		assertEqual(t, value, expectedValue, fmt.Sprintf("Get(%d)", key))
	}
}

func TestSetOverwrite(t *testing.T) {
	store := newLockingStore[string, int](10)

	assertNoError(t, store.Set("key", 100), "Set initial value")
	assertNoError(t, store.Set("key", 200), "Set overwrite")

	value, found := store.Get("key")
	assertTrue(t, found, "Get() did not find the key after overwrite")
	assertEqual(t, value, 200, "value after overwrite")
}

func TestDelete(t *testing.T) {
	store := newLockingStore[string, int](10)

	assertNoError(t, store.Set("key1", 100), "Set")

	_, found := store.Get("key1")
	assertTrue(t, found, "Key not found before delete")

	assertNoError(t, store.Delete("key1"), "Delete")

	_, found = store.Get("key1")
	assertTrue(t, !found, "Key still exists after delete")
	assertEqual(t, store.Size(), uint64(0), "size after delete")
}

func TestDeleteNonExistent(t *testing.T) {
	store := newLockingStore[string, int](10)

	err := store.Delete("nonexistent")
	assertTrue(t, err != nil, "Expected error when deleting from empty bucket")
}

func TestSize(t *testing.T) {
	store := newLockingStore[int, int](10)

	assertEqual(t, store.Size(), uint64(0), "initial size")

	for i := 1; i <= 5; i++ {
		assertNoError(t, store.Set(i, i*10), fmt.Sprintf("Set(%d)", i))
		assertEqual(t, store.Size(), uint64(i), fmt.Sprintf("size after %d insertions", i))
	}

	for i := 1; i <= 5; i++ {
		assertNoError(t, store.Delete(i), fmt.Sprintf("Delete(%d)", i))
		assertEqual(t, store.Size(), uint64(5-i), fmt.Sprintf("size after %d deletions", i))
	}
}

func TestClear(t *testing.T) {
	store := newLockingStore[int, string](10)

	for i := 1; i <= 10; i++ {
		assertNoError(t, store.Set(i, fmt.Sprintf("value%d", i)), fmt.Sprintf("Set(%d)", i))
	}

	assertEqual(t, store.Size(), uint64(10), "size before clear")
	store.Clear()
	assertEqual(t, store.Size(), uint64(0), "size after clear")

	assertNoError(t, store.Set(1, "new"), "Set after clear")
	value, found := store.Get(1)
	assertTrue(t, found && value == "new", "Get after clear and re-add")
}

func TestResize(t *testing.T) {
	initialCapacity := 4
	store := newLockingStore[int, int](initialCapacity)

	numItems := initialCapacity*2 + 5
	for i := 0; i < numItems; i++ {
		assertNoError(t, store.Set(i, i*10), fmt.Sprintf("Set(%d)", i))
	}

	// Verify all items are still accessible after resize
	for i := 0; i < numItems; i++ {
		value, found := store.Get(i)
		assertTrue(t, found, fmt.Sprintf("Key %d not found after resize", i))
		assertEqual(t, value, i*10, fmt.Sprintf("Key %d value", i))
	}
}

func TestEmptyStore(t *testing.T) {
	store := newLockingStore[int, int](10)

	_, found := store.Get(1)
	assertTrue(t, !found, "Found key in empty store")
	assertEqual(t, store.Size(), uint64(0), "Empty store size")

	err := store.Delete(1)
	assertTrue(t, err != nil, "Expected error deleting from empty store")

	store.Clear() // Should not panic
}

func TestConcurrentReads(t *testing.T) {
	store := newLockingStore[int, int](100)

	// Pre-populate the store
	for i := 0; i < 100; i++ {
		assertNoError(t, store.Set(i, i*10), fmt.Sprintf("Set(%d)", i))
	}

	var wg sync.WaitGroup
	errors := make(chan error, 1000)

	for g := 0; g < 10; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				key := rand.Intn(100)
				value, found := store.Get(key)
				if !found {
					errors <- fmt.Errorf("key %d not found", key)
					return
				}
				if value != key*10 {
					errors <- fmt.Errorf("key %d: expected %d, got %d", key, key*10, value)
					return
				}
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Error(err)
	}
}

func TestConcurrentWrites(t *testing.T) {
	store := newLockingStore[int, int](100)

	var wg sync.WaitGroup
	numGoroutines := 10
	writesPerGoroutine := 100

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for i := 0; i < writesPerGoroutine; i++ {
				key := goroutineID*1000 + i
				if err := store.Set(key, key*10); err != nil {
					t.Errorf("Concurrent write error: %v", err)
					return
				}
			}
		}(g)
	}

	wg.Wait()

	// Verify all writes succeeded
	for g := 0; g < numGoroutines; g++ {
		for i := 0; i < writesPerGoroutine; i++ {
			key := g*1000 + i
			value, found := store.Get(key)
			assertTrue(t, found, fmt.Sprintf("Key %d not found", key))
			assertEqual(t, value, key*10, fmt.Sprintf("Key %d value", key))
		}
	}
}

func TestConcurrentMixed(t *testing.T) {
	store := newLockingStore[int, int](100)

	// Pre-populate
	for i := 0; i < 50; i++ {
		assertNoError(t, store.Set(i, i*10), fmt.Sprintf("Set(%d)", i))
	}

	var wg sync.WaitGroup
	for g := 0; g < 10; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				key := rand.Intn(100)
				switch {
				case rand.Float64() < 0.7:
					store.Get(key)
				case rand.Float64() < 0.9:
					store.Set(key, key*10)
				default:
					store.Delete(key)
				}
			}
		}()
	}

	wg.Wait()

	// Verify the store is still functional
	assertNoError(t, store.Set(999, 9990), "Store functional after mixed ops")
	value, found := store.Get(999)
	assertTrue(t, found && value == 9990, "Store working correctly after mixed ops")
}

func TestDifferentValueTypes(t *testing.T) {
	t.Run("StringValues", func(t *testing.T) {
		store := newLockingStore[int, string](10)
		assertNoError(t, store.Set(1, "hello"), "Set string")
		value, found := store.Get(1)
		assertTrue(t, found && value == "hello", "Get string value")
	})

	t.Run("StructValues", func(t *testing.T) {
		type Person struct {
			Name string
			Age  int
		}
		factory := func(capacity int) interfaces.IKVStore[string, Person] {
			llFactory := func() interfaces.ILinkedList[string, Person] {
				return linkedlist.New[string, Person]()
			}
			return kv_store.New[string, Person](uint64(capacity), llFactory)
		}
		store := factory(10)
		person := Person{Name: "Alice", Age: 30}
		assertNoError(t, store.Set("alice", person), "Set struct")
		value, found := store.Get("alice")
		assertTrue(t, found && value.Name == "Alice" && value.Age == 30, "Get struct value")
	})

	t.Run("PointerValues", func(t *testing.T) {
		factory := func(capacity int) interfaces.IKVStore[int, *string] {
			llFactory := func() interfaces.ILinkedList[int, *string] {
				return linkedlist.New[int, *string]()
			}
			return kv_store.New[int, *string](uint64(capacity), llFactory)
		}
		store := factory(10)
		str := "test"
		assertNoError(t, store.Set(1, &str), "Set pointer")
		value, found := store.Get(1)
		assertTrue(t, found && *value == "test", "Get pointer value")
	})
}

func TestLargeDataset(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large dataset test in short mode")
	}

	store := newLockingStore[int, int](100)
	numItems := 10000

	for i := 0; i < numItems; i++ {
		assertNoError(t, store.Set(i, i*10), fmt.Sprintf("Insert item %d", i))
	}

	// Verify random samples
	for i := 0; i < 100; i++ {
		key := rand.Intn(numItems)
		value, found := store.Get(key)
		assertTrue(t, found, fmt.Sprintf("Key %d not found", key))
		assertEqual(t, value, key*10, fmt.Sprintf("Key %d value", key))
	}
}

func TestAtomicKVStore(t *testing.T) {
	// Run all the same tests for atomic implementation
	testCases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{"New", func(t *testing.T) {
			store := newAtomicStore[int, int](10)
			assertTrue(t, store != nil, "Factory returned nil")
			assertEqual(t, store.Size(), uint64(0), "size")
		}},
		{"SetAndGet", func(t *testing.T) {
			store := newAtomicStore[string, int](10)
			assertNoError(t, store.Set("key1", 100), "Set failed")
			value, found := store.Get("key1")
			assertTrue(t, found, "Get() did not find the key")
			assertEqual(t, value, 100, "value")
		}},
		{"ConcurrentMixed", func(t *testing.T) {
			store := newAtomicStore[int, int](100)
			for i := 0; i < 50; i++ {
				assertNoError(t, store.Set(i, i*10), fmt.Sprintf("Set(%d)", i))
			}
			var wg sync.WaitGroup
			for g := 0; g < 10; g++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for i := 0; i < 100; i++ {
						key := rand.Intn(100)
						switch {
						case rand.Float64() < 0.7:
							store.Get(key)
						case rand.Float64() < 0.9:
							store.Set(key, key*10)
						default:
							store.Delete(key)
						}
					}
				}()
			}
			wg.Wait()
			assertNoError(t, store.Set(999, 9990), "Store functional after mixed ops")
			value, found := store.Get(999)
			assertTrue(t, found && value == 9990, "Store working correctly after mixed ops")
		}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, tc.fn)
	}
}

func TestShardedKVStore(t *testing.T) {
	testCases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{"New", func(t *testing.T) {
			store := newShardedStore[int, int](80)
			assertTrue(t, store != nil, "Factory returned nil")
			assertEqual(t, store.Size(), uint64(0), "size")
		}},
		{"SetAndGet", func(t *testing.T) {
			store := newShardedStore[string, int](80)
			assertNoError(t, store.Set("key1", 100), "Set failed")
			value, found := store.Get("key1")
			assertTrue(t, found, "Get() did not find the key")
			assertEqual(t, value, 100, "value")
		}},
		{"MultipleShards", func(t *testing.T) {
			store := newShardedStore[int, int](80)
			for i := 0; i < 100; i++ {
				assertNoError(t, store.Set(i, i*10), fmt.Sprintf("Set(%d)", i))
			}
			for i := 0; i < 100; i++ {
				value, found := store.Get(i)
				assertTrue(t, found, fmt.Sprintf("Key %d not found", i))
				assertEqual(t, value, i*10, fmt.Sprintf("Key %d value", i))
			}
		}},
		{"ConcurrentWritesDifferentShards", func(t *testing.T) {
			store := newShardedStore[int, int](80)
			var wg sync.WaitGroup
			numGoroutines := 16
			writesPerGoroutine := 100

			for g := 0; g < numGoroutines; g++ {
				wg.Add(1)
				go func(goroutineID int) {
					defer wg.Done()
					for i := 0; i < writesPerGoroutine; i++ {
						key := goroutineID*1000 + i
						if err := store.Set(key, key*10); err != nil {
							t.Errorf("Concurrent write error: %v", err)
							return
						}
					}
				}(g)
			}
			wg.Wait()
			for g := 0; g < numGoroutines; g++ {
				for i := 0; i < writesPerGoroutine; i++ {
					key := g*1000 + i
					value, found := store.Get(key)
					assertTrue(t, found, fmt.Sprintf("Key %d not found", key))
					assertEqual(t, value, key*10, fmt.Sprintf("Key %d value", key))
				}
			}
		}},
		{"ConcurrentMixed", func(t *testing.T) {
			store := newShardedStore[int, int](80)
			for i := 0; i < 50; i++ {
				assertNoError(t, store.Set(i, i*10), fmt.Sprintf("Set(%d)", i))
			}
			var wg sync.WaitGroup
			for g := 0; g < 16; g++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for i := 0; i < 100; i++ {
						key := rand.Intn(100)
						switch {
						case rand.Float64() < 0.7:
							store.Get(key)
						case rand.Float64() < 0.9:
							store.Set(key, key*10)
						default:
							store.Delete(key)
						}
					}
				}()
			}
			wg.Wait()
			assertNoError(t, store.Set(999, 9990), "Store functional after mixed ops")
			value, found := store.Get(999)
			assertTrue(t, found && value == 9990, "Store working correctly after mixed ops")
		}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, tc.fn)
	}
}
