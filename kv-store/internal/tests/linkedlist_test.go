package tests

import (
	"sync"
	"testing"

	"kv-store/internal/interfaces"
	"kv-store/internal/linkedlist"
)

// Generic test functions to reduce code duplication

func testBasicOperations[T interfaces.ILinkedList[int]](t *testing.T, list T, testName string) {
	// Test Set
	if err := list.Set(1, 100); err != nil {
		t.Fatalf("%s Set failed: %v", testName, err)
	}
	if err := list.Set(2, 200); err != nil {
		t.Fatalf("%s Set failed: %v", testName, err)
	}

	// Test Get
	value, found := list.Get(1)
	if !found || value != 100 {
		t.Fatalf("%s Get failed: expected 100, got %v", testName, value)
	}
	value, found = list.Get(2)
	if !found || value != 200 {
		t.Fatalf("%s Get failed: expected 200, got %v", testName, value)
	}
	value, found = list.Get(3)
	if found {
		t.Fatalf("%s Get failed: expected not found, got %v", testName, value)
	}

	// Test Delete
	if err := list.Delete(1); err != nil {
		t.Fatalf("%s Delete failed: %v", testName, err)
	}
	_, found = list.Get(1)
	if found {
		t.Fatalf("%s Get after Delete failed: expected not found", testName)
	}
}

func testEmptyList[T interfaces.ILinkedList[string]](t *testing.T, list T, testName string) {
	// Test Get on empty list
	value, found := list.Get("nonexistent")
	if found {
		t.Fatalf("%s Get on empty list: expected not found, got %v", testName, value)
	}
	if value != "" {
		t.Fatalf("%s Get on empty list: expected zero value, got %v", testName, value)
	}

	// Test Delete on empty list
	if err := list.Delete("nonexistent"); err != nil {
		t.Fatalf("%s Delete on empty list failed: %v", testName, err)
	}
}

func testSingleElement[T interfaces.ILinkedList[string]](t *testing.T, list T, testName string) {
	// Add single element
	if err := list.Set("key1", "value1"); err != nil {
		t.Fatalf("%s Set failed: %v", testName, err)
	}

	// Get the element
	value, found := list.Get("key1")
	if !found || value != "value1" {
		t.Fatalf("%s Get failed: expected 'value1', got %v", testName, value)
	}

	// Delete the element
	if err := list.Delete("key1"); err != nil {
		t.Fatalf("%s Delete failed: %v", testName, err)
	}

	// Verify it's gone
	_, found = list.Get("key1")
	if found {
		t.Fatalf("%s Get after Delete: expected not found", testName)
	}
}

func testUpdateExistingKey[T interfaces.ILinkedList[int]](t *testing.T, list T, testName string) {
	// Set initial value
	if err := list.Set(1, 100); err != nil {
		t.Fatalf("%s Set failed: %v", testName, err)
	}

	// Update the value
	if err := list.Set(1, 200); err != nil {
		t.Fatalf("%s Set update failed: %v", testName, err)
	}

	// Verify updated value
	value, found := list.Get(1)
	if !found || value != 200 {
		t.Fatalf("%s Get after update: expected 200, got %v", testName, value)
	}
}

func testConcurrentAccess[T interfaces.ILinkedList[int]](t *testing.T, list T, testName string) {
	const numGoroutines = 50
	const numOperations = 100

	var wg sync.WaitGroup

	// Concurrent writers
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := id*numOperations + j
				if err := list.Set(key, key*2); err != nil {
					t.Errorf("%s Concurrent Set failed: %v", testName, err)
				}
			}
		}(i)
	}

	// Concurrent readers
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := id*numOperations + j
				// Try to read, but don't fail if not found due to timing
				list.Get(key)
			}
		}(i)
	}

	// Concurrent deleters
	for i := 0; i < numGoroutines/2; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations/2; j++ {
				key := id*numOperations + j
				list.Delete(key)
			}
		}(i)
	}

	wg.Wait()

	t.Logf("%s concurrent access test completed successfully", testName)
}

func testZeroValues[T interfaces.ILinkedList[int]](t *testing.T, list T, testName string) {
	// Test setting zero value
	if err := list.Set(0, 0); err != nil {
		t.Fatalf("%s Set zero value failed: %v", testName, err)
	}

	// Test getting zero value
	value, found := list.Get(0)
	if !found || value != 0 {
		t.Fatalf("%s Get zero value failed: expected 0, got %v", testName, value)
	}

	// Test deleting zero value key
	if err := list.Delete(0); err != nil {
		t.Fatalf("%s Delete zero key failed: %v", testName, err)
	}

	// Verify deletion
	_, found = list.Get(0)
	if found {
		t.Fatalf("%s Get after Delete zero key: expected not found", testName)
	}
}

// Specific tests for each implementation

// Lock-based LinkedList Tests
func TestLockBasedLinkedList_BasicOperations(t *testing.T) {
	list := linkedlist.NewLinkedList[int]()
	testBasicOperations(t, list, "LockBased")
}

func TestLockBasedLinkedList_EmptyList(t *testing.T) {
	list := linkedlist.NewLinkedList[string]()
	testEmptyList(t, list, "LockBased")
}

func TestLockBasedLinkedList_SingleElement(t *testing.T) {
	list := linkedlist.NewLinkedList[string]()
	testSingleElement(t, list, "LockBased")
}

func TestLockBasedLinkedList_UpdateExistingKey(t *testing.T) {
	list := linkedlist.NewLinkedList[int]()
	testUpdateExistingKey(t, list, "LockBased")
}

func TestLockBasedLinkedList_ConcurrentAccess(t *testing.T) {
	list := linkedlist.NewLinkedList[int]()
	testConcurrentAccess(t, list, "LockBased")
}

func TestLockBasedLinkedList_ZeroValues(t *testing.T) {
	list := linkedlist.NewLinkedList[int]()
	testZeroValues(t, list, "LockBased")
}

// Atomic LinkedList Tests
func TestAtomicLinkedList_BasicOperations(t *testing.T) {
	list := linkedlist.NewAtomicLinkedList[int]()
	testBasicOperations(t, list, "Atomic")
}

func TestAtomicLinkedList_EmptyList(t *testing.T) {
	list := linkedlist.NewAtomicLinkedList[string]()
	testEmptyList(t, list, "Atomic")
}

func TestAtomicLinkedList_SingleElement(t *testing.T) {
	list := linkedlist.NewAtomicLinkedList[string]()
	testSingleElement(t, list, "Atomic")
}

func TestAtomicLinkedList_UpdateExistingKey(t *testing.T) {
	list := linkedlist.NewAtomicLinkedList[int]()
	testUpdateExistingKey(t, list, "Atomic")
}

func TestAtomicLinkedList_ConcurrentAccess(t *testing.T) {
	list := linkedlist.NewAtomicLinkedList[int]()
	testConcurrentAccess(t, list, "Atomic")
}

func TestAtomicLinkedList_ZeroValues(t *testing.T) {
	list := linkedlist.NewAtomicLinkedList[int]()
	testZeroValues(t, list, "Atomic")
}

// Sharded LinkedList (Lock-based) Tests
func TestShardedLinkedList_BasicOperations(t *testing.T) {
	list := linkedlist.NewShardedLinkedList[int](16)
	testBasicOperations(t, list, "ShardedLock")
}

func TestShardedLinkedList_EmptyList(t *testing.T) {
	list := linkedlist.NewShardedLinkedList[string](16)
	testEmptyList(t, list, "ShardedLock")
}

func TestShardedLinkedList_SingleElement(t *testing.T) {
	list := linkedlist.NewShardedLinkedList[string](16)
	testSingleElement(t, list, "ShardedLock")
}

func TestShardedLinkedList_UpdateExistingKey(t *testing.T) {
	list := linkedlist.NewShardedLinkedList[int](16)
	testUpdateExistingKey(t, list, "ShardedLock")
}

func TestShardedLinkedList_ConcurrentAccess(t *testing.T) {
	list := linkedlist.NewShardedLinkedList[int](16)
	testConcurrentAccess(t, list, "ShardedLock")
}

func TestShardedLinkedList_ZeroValues(t *testing.T) {
	list := linkedlist.NewShardedLinkedList[int](16)
	testZeroValues(t, list, "ShardedLock")
}

// Sharded Atomic LinkedList Tests
func TestShardedAtomicLinkedList_BasicOperations(t *testing.T) {
	list := linkedlist.NewShardedAtomicLinkedList[int](16)
	testBasicOperations(t, list, "ShardedAtomic")
}

func TestShardedAtomicLinkedList_EmptyList(t *testing.T) {
	list := linkedlist.NewShardedAtomicLinkedList[string](16)
	testEmptyList(t, list, "ShardedAtomic")
}

func TestShardedAtomicLinkedList_SingleElement(t *testing.T) {
	list := linkedlist.NewShardedAtomicLinkedList[string](16)
	testSingleElement(t, list, "ShardedAtomic")
}

func TestShardedAtomicLinkedList_UpdateExistingKey(t *testing.T) {
	list := linkedlist.NewShardedAtomicLinkedList[int](16)
	testUpdateExistingKey(t, list, "ShardedAtomic")
}

func TestShardedAtomicLinkedList_ConcurrentAccess(t *testing.T) {
	list := linkedlist.NewShardedAtomicLinkedList[int](16)
	testConcurrentAccess(t, list, "ShardedAtomic")
}

func TestShardedAtomicLinkedList_ZeroValues(t *testing.T) {
	list := linkedlist.NewShardedAtomicLinkedList[int](16)
	testZeroValues(t, list, "ShardedAtomic")
}

// Additional specific tests
func TestShardedLinkedList_DifferentShardCounts(t *testing.T) {
	testShardCounts := []int{1, 2, 4, 8, 16, 32}

	for _, shardCount := range testShardCounts {
		t.Run("ShardCount_"+string(rune(shardCount+'0')), func(t *testing.T) {
			list := linkedlist.NewShardedLinkedList[int](shardCount)
			testBasicOperations(t, list, "ShardedLock_"+string(rune(shardCount+'0')))
		})
	}
}

func TestShardedAtomicLinkedList_DifferentShardCounts(t *testing.T) {
	testShardCounts := []int{1, 2, 4, 8, 16, 32}

	for _, shardCount := range testShardCounts {
		t.Run("ShardCount_"+string(rune(shardCount+'0')), func(t *testing.T) {
			list := linkedlist.NewShardedAtomicLinkedList[int](shardCount)
			testBasicOperations(t, list, "ShardedAtomic_"+string(rune(shardCount+'0')))
		})
	}
}
