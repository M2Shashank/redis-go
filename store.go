package main

import (
	"sync"
)

type Store struct {
	data map[string]string
	dataMutex sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		data: make(map[string]string),
	}
}

func (s *Store) Get(key string) (string, bool) {
	s.dataMutex.RLock()
	defer s.dataMutex.RUnlock()

	value, exists := s.data[key]
	return value, exists
}

func (s *Store) Set(key string, value string) {
	s.dataMutex.Lock()
	defer s.dataMutex.Unlock()

	s.data[key] = value
}

func (s *Store) Delete(key string) {
	s.dataMutex.Lock()
	defer s.dataMutex.Unlock()

	delete(s.data, key)
}