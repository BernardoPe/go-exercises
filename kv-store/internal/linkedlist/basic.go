package linkedlist

import (
	"sync"
)

type Node[T comparable] struct {
	key   T
	value T
	next  *Node[T]
}

type LinkedList[T comparable] struct {
	head *Node[T]
	tail *Node[T]
	size int
	sync.RWMutex
}

func NewLinkedList[T comparable]() *LinkedList[T] {
	return &LinkedList[T]{}
}

func (l *LinkedList[T]) Get(key T) (T, bool) {
	l.RWMutex.RLock()
	defer l.RWMutex.RUnlock()

	curr := l.head
	for curr != nil {
		if curr.key == key {
			return curr.value, true
		}
		curr = curr.next
	}

	var zero T
	return zero, false
}

func (l *LinkedList[T]) Set(key T, value T) error {
	l.RWMutex.Lock()
	defer l.RWMutex.Unlock()

	curr := l.head
	for curr != nil {
		if curr.key == key {
			curr.value = value
			return nil
		}
		curr = curr.next
	}

	newNode := &Node[T]{key: key, value: value}
	if l.head == nil {
		l.head = newNode
		l.tail = newNode
	} else {
		l.tail.next = newNode
		l.tail = newNode
	}

	l.size++
	return nil
}

func (l *LinkedList[T]) Delete(key T) error {
	l.RWMutex.Lock()
	defer l.RWMutex.Unlock()

	var prev *Node[T]
	curr := l.head
	for curr != nil {
		if curr.key == key {
			if prev == nil {
				l.head = curr.next
			} else {
				prev.next = curr.next
			}
			if curr == l.tail {
				l.tail = prev
			}
			l.size--
			return nil
		}
		prev = curr
		curr = curr.next
	}

	return nil
}

func (l *LinkedList[T]) Size() int {
	l.RWMutex.RLock()
	defer l.RWMutex.RUnlock()
	return l.size
}
