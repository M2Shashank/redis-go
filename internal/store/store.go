package store

import (
	"fmt"
	"os"
	"sync"
	"time"

	RDB_Parser "github.com/codecrafters-io/redis-starter-go/internal/store/rdb_parser"
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

func (s *Store) LoadFromRDBFile(dir string, dbfilename string) {
	addr := dir + "/" + dbfilename

	data, err := os.ReadFile(addr)
	if err != nil {
		fmt.Println("Error reading RDB file: ", err.Error())
		return
	}
	fmt.Println("Read ", len(data), " bytes from RDB file: ", addr)

	result := RDB_Parser.ParseRDBFile(&data)

	for key, valueType := range result {
		_, value, expireTime := valueType.Type, valueType.Value, valueType.ExpireTimeMS

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
