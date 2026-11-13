package store

import (
	"sync"
	"time"
	"encoding/binary"
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

func (s *Store) GetKeyVersion(key string) int64 {
	s.dataMutex.RLock()
	defer s.dataMutex.RUnlock()

	obj, exists := s.data[key]
	if !exists {
		return 0
	}

	return obj.version
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

func (s *Store) LoadRDBIntoStore(data map[string][]byte) {
	for key, valueType := range data {
		_, expireSlice, value := valueType[0], valueType[1:5], valueType[5:]
		expireTime := int64(binary.BigEndian.Uint32(expireSlice))

		if expireTime == 0 {
			s.Set(key, string(value))
			continue
		}

		remaining := int64(expireTime) - time.Now().UnixMilli()
		if remaining <= 0 {
			continue
		}

		s.Set(key, string(value))
		go func() {
			time.Sleep(time.Duration(remaining) * time.Millisecond)
			s.Delete(key)
		}()
	}
}

func (s *Store) GetKeysOfPattern(pattern string) []string {
	s.dataMutex.RLock()
	defer s.dataMutex.RUnlock()

	var keys []string
	for key := range s.data {
		keys = append(keys, key)
	}
	return keys
}
