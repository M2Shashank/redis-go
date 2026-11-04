package store

import (
	"sync"
)

type Store struct {
	data      map[string]RedisObject
	dataMutex sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		data: make(map[string]RedisObject),
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

func (s *Store) Set(key string, typ ObjectType, value string) {
	s.dataMutex.Lock()
	defer s.dataMutex.Unlock()

	if typ == StringType {
		s.data[key] = RedisObject{
			Type:  typ,
			Value: value,
		}
	} else if typ == ListType {
		listObj, exists := s.data[key]
		if !exists {
			s.data[key] = RedisObject{
				Type:  typ,
				Value: []string{value},
			}
		} else {
			listObj.Value = append(listObj.Value.([]string), value)
		}
	}
}

func (s *Store) Delete(key string) {
	s.dataMutex.Lock()
	defer s.dataMutex.Unlock()

	delete(s.data, key)
}
