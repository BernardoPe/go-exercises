package linkedlist

import (
	"sync/atomic"
)

type AtomicNode[T comparable] struct {
	key   T
	value T
	next  atomic.Pointer[AtomicNode[T]]
}

type AtomicLinkedList[T comparable] struct {
	head atomic.Pointer[AtomicNode[T]]
	size atomic.Int64
}

func NewAtomicLinkedList[T comparable]() *AtomicLinkedList[T] {
	return &AtomicLinkedList[T]{}
}

func (l *AtomicLinkedList[T]) Get(key T) (T, bool) {
	curr := l.head.Load()
	for curr != nil {
		if curr.key == key {
			return curr.value, true
		}
		curr = curr.next.Load()
	}
	var zero T
	return zero, false
}

func (l *AtomicLinkedList[T]) Set(key T, value T) error {
	for {
		head := l.head.Load()

		curr := head
		for curr != nil {
			if curr.key == key {
				curr.value = value
				return nil
			}
			curr = curr.next.Load()
		}

		newNode := &AtomicNode[T]{key: key, value: value}
		newNode.next.Store(head)

		if l.head.CompareAndSwap(head, newNode) {
			l.size.Add(1)
			return nil
		}
	}
}

func (l *AtomicLinkedList[T]) Delete(key T) error {
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

func (l *AtomicLinkedList[T]) Size() int64 {
	return l.size.Load()
}
