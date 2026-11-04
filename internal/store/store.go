package store

import (
	"sync"
)

type Store struct {
	data      map[string]*RedisObject
	dataMutex sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		data: make(map[string]*RedisObject),
	}
}

// ------------- String Operations ---------------
func (s *Store) Get(key string) (string, bool) {
	s.dataMutex.RLock()
	defer s.dataMutex.RUnlock()

	obj, exists := s.data[key]
	if !exists || obj.Type != StringType {
		return "", false
	}

	return obj.Value.(string), exists
}

func (s *Store) Set(key string, value string) int {
	s.dataMutex.Lock()
	defer s.dataMutex.Unlock()

	obj, exists := s.data[key]
	if !exists {
		s.data[key] = &RedisObject{StringType, value}
		return 1
	}

	if exists && obj.Type != StringType {
		return -1
	}

	s.data[key].Value = value
	return 1
}

func (s *Store) Delete(key string) {
	s.dataMutex.Lock()
	defer s.dataMutex.Unlock()

	delete(s.data, key)
}

// ------------- List Operations ---------------
func (s *Store) RPush(key string, value string) int {
	s.dataMutex.Lock()
	defer s.dataMutex.Unlock()

	obj, exists := s.data[key]
	if !exists {
		s.data[key] = &RedisObject{ListType, []string{value}}
		return 1
	}

	if exists && obj.Type != ListType {
		return -1
	}

	listObj := s.data[key]
	listObj.Value = append(listObj.Value.([]string), value)
	return len(listObj.Value.([]string))
}

func (s *Store) LRange(key string, start int, stop int) []string {
	s.dataMutex.RLock()
	defer s.dataMutex.RUnlock()

	obj, exists := s.data[key]
	if !exists || obj.Type != ListType {
		return nil
	}

	list := obj.Value.([]string)

	if start < 0 {
		start = len(list) + start
	}
	if stop < 0 {
		stop = len(list) + stop
	}

	start = max(start, 0)
	stop = min(stop, len(list)-1)
	if start > stop {
		return nil
	}

	return list[start : stop+1]
}
