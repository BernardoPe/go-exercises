package linkedlist

import (
	"sync/atomic"
)

type AtomicNode[K comparable, V any] struct {
	key   K
	value atomic.Value
	next  atomic.Pointer[AtomicNode[K, V]]
}

type AtomicLinkedList[K comparable, V any] struct {
	head atomic.Pointer[AtomicNode[K, V]]
	size atomic.Int64
}

func NewAtomic[K comparable, V any]() *AtomicLinkedList[K, V] {
	return &AtomicLinkedList[K, V]{}
}

func (l *AtomicLinkedList[K, V]) Get(key K) (V, bool) {
	curr := l.head.Load()
	for curr != nil {
		if curr.key == key {
			return curr.value.Load().(V), true
		}
		curr = curr.next.Load()
	}
	var zero V
	return zero, false
}

func (l *AtomicLinkedList[K, V]) Set(key K, value V) error {
	for {
		head := l.head.Load()

		curr := head
		for curr != nil {
			if curr.key == key {
				curr.value.Store(value)
				return nil
			}
			curr = curr.next.Load()
		}

		newNode := &AtomicNode[K, V]{key: key}
		newNode.value.Store(value)
		newNode.next.Store(head)

		if l.head.CompareAndSwap(head, newNode) {
			l.size.Add(1)
			return nil
		}
	}
}

func (l *AtomicLinkedList[K, V]) Delete(key K) error {
	for {
		head := l.head.Load()
		if head == nil {
			return nil
		}

		if head.key == key {
			next := head.next.Load()
			if l.head.CompareAndSwap(head, next) {
				l.size.Add(-1)
				return nil
			}
			continue
		}

		curr := head
		for curr.next.Load() != nil {
			next := curr.next.Load()
			if next.key == key {
				nextNext := next.next.Load()
				if curr.next.CompareAndSwap(next, nextNext) {
					l.size.Add(-1)
					return nil
				}
				break
			}
			curr = next
		}

		if curr.next.Load() == nil {
			return nil
		}
	}
}

func (l *AtomicLinkedList[K, V]) Size() int64 {
	return l.size.Load()
}

func (l *AtomicLinkedList[K, V]) ForEach(fn func(key K, value V) bool) {
	current := l.head.Load()
	for current != nil {
		val := current.value.Load().(V)
		if !fn(current.key, val) {
			return
		}
		current = current.next.Load()
	}
}
