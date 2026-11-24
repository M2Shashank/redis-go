package store

import (
	"sync"
	"time"

	rdb_parser "github.com/codecrafters-io/redis-starter-go/internal/store/rdb"
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

func (s *Store) LoadRDBIntoStore(rdbFile []byte) {
	data := rdb_parser.ParseRDBFile(&rdbFile)

	for key, valueType := range data {
		_, expireTime, value := valueType.Type, valueType.ExpireTimeMS, valueType.Value

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

func (s *Store) GetKeyVersion(key string) int64 {
	s.dataMutex.RLock()
	defer s.dataMutex.RUnlock()

	obj, exists := s.data[key]
	if !exists {
		return 0
	}

	return obj.version
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
	case SortedSetType:
		return "zset"
	default:
		return "none"
	}
}