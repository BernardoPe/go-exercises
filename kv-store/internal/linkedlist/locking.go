package linkedlist

import (
	"sync"
)

type Node[K comparable, V any] struct {
	key   K
	value V
	next  *Node[K, V]
}

type LinkedList[K comparable, V any] struct {
	head *Node[K, V]
	tail *Node[K, V]
	size int
	sync.RWMutex
}

func New[K comparable, V any]() *LinkedList[K, V] {
	return &LinkedList[K, V]{}
}

func (l *LinkedList[K, V]) Get(key K) (V, bool) {
	l.RWMutex.RLock()
	defer l.RWMutex.RUnlock()

	curr := l.head
	for curr != nil {
		if curr.key == key {
			return curr.value, true
		}
		curr = curr.next
	}

	var zero V
	return zero, false
}

func (l *LinkedList[K, V]) Set(key K, value V) error {
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

	newNode := &Node[K, V]{key: key, value: value}
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

func (l *LinkedList[K, V]) Delete(key K) error {
	l.RWMutex.Lock()
	defer l.RWMutex.Unlock()

	var prev *Node[K, V]
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

func (l *LinkedList[K, V]) Size() int {
	l.RWMutex.RLock()
	defer l.RWMutex.RUnlock()
	return l.size
}

func (l *LinkedList[K, V]) ForEach(fn func(key K, value V) bool) {
	l.RWMutex.RLock()
	defer l.RWMutex.RUnlock()

	current := l.head
	for current != nil {
		if !fn(current.key, current.value) {
			return
		}
		current = current.next
	}
}
