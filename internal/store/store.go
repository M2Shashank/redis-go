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

func (s *Store) TypeOf(key string) string {
	s.dataMutex.RLock()
	defer s.dataMutex.RUnlock()

	obj, exists := s.data[key]
	if !exists {
		return "none"
	}

	switch obj.Type {
	case StringType:
		return "string"
	case ListType:
		return "list"
	case StreamType:
		return "stream"
	default:
		return "none"
	}
}
