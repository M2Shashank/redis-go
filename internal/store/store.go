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

func (s *Store) Get(key string) (string, bool) {
	s.dataMutex.RLock()
	defer s.dataMutex.RUnlock()

	obj, exists := s.data[key]
	if !exists || obj.Type != StringType {
		return "", false
	}

	return obj.Value.(string), exists
}

func (s *Store) Set(key string, typ ObjectType, value string) int {
	s.dataMutex.Lock()
	defer s.dataMutex.Unlock()

	obj, exists := s.data[key]
	if exists && obj.Type != typ {
		return -1
	}

	if !exists {
		if typ == ListType {
			s.data[key] = &RedisObject{ typ, []string{value} }
		} else if typ == StringType {
			s.data[key] = &RedisObject{ typ, value }
		}
		return 1
	}

	var result int = 1
	if typ == StringType {
		s.data[key].Value = value
	} else if typ == ListType {
		listObj := s.data[key]
		listObj.Value = append(listObj.Value.([]string), value)
		result = len(listObj.Value.([]string))
	}
	return result
}

func (s *Store) Delete(key string) {
	s.dataMutex.Lock()
	defer s.dataMutex.Unlock()

	delete(s.data, key)
}
