package store

import (
	"sync"
)

type Store struct {
	data      map[string]*RedisObject
	dataMutex sync.RWMutex

	waiting map[string][]chan string
}

func NewStore() *Store {
	return &Store{
		data:    make(map[string]*RedisObject),
		waiting: make(map[string][]chan string),
	}
}

// to be used when dataMutex is already locked
func (s *Store) serveWaiters(key string, list *DoublyLL) {
	queue, exists := s.waiting[key]
	if !exists || len(queue) == 0 {
		return
	}

	for len(queue) > 0 && list.Length > 0 {
		var ch chan string = queue[0]
		queue = queue[1:]

		value, _ := list.PopFront()
		ch <- value
	}

	if len(queue) == 0 {
		delete(s.waiting, key)
	} else {
		s.waiting[key] = queue
	}
}

