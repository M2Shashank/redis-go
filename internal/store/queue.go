package store

import (
)

type Queue[T any] struct {
	Entries *DoublyLL[T]
	size int
}

func NewQueue[T any]() *Queue[T] {
	return &Queue[T] {
		Entries: &DoublyLL[T]{},
		size: 0,
	}
}

func (q *Queue[T]) Size() int {
	return q.size
}

func (q *Queue[T]) Empty() bool {
	return q.size == 0
}

func (q *Queue[T]) Push(value T) {
	q.Entries.PushBack(value)
	q.size++
}

func (q *Queue[T]) Pop() T {
	if q.size == 0 {
		var zero T
		return zero
	}
	q.size--
	
	val ,_ := q.Entries.PopFront()
	return val
}
