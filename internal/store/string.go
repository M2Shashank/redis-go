package store

import (
	"strconv"

	storage "github.com/codecrafters-io/redis-starter-go/internal/store/storage"
	Error "github.com/codecrafters-io/redis-starter-go/internal/errors"
)

func createStringObject() *RedisObject {
	obj := &RedisObject{
		Type: StringType,
		Value: &storage.String{
			Data:   []byte(""),
			Length: 0,
		},
		version: 0,
	}
	return obj
}

// ----------------- String Operations ------------------
func (s *Store) Get(key string) (string, error) {
	s.dataMutex.RLock()
	defer s.dataMutex.RUnlock()

	obj, exists := s.data[key]
	if !exists {
		return "", nil
	}
	if obj.Type != StringType {
		return "", Error.ERR_WRONGTYPE("STRING")
	}

	var ans string = string(obj.Value.(*storage.String).Data)
	return ans, nil
}

func (s *Store) Set(key string, value string) error {
	s.dataMutex.Lock()
	defer s.dataMutex.Unlock()

	obj, exists := s.data[key]
	if !exists {
		s.data[key] = createStringObject()
		obj = s.data[key]
	}

	if exists && obj.Type != StringType {
		return Error.ERR_WRONGTYPE("STRING")
	}

	obj.Value.(*storage.String).SetData(value)
	s.data[key].version++
	return nil
}

func (s *Store) Delete(key string) {
	s.dataMutex.Lock()
	defer s.dataMutex.Unlock()

	delete(s.data, key)
}

func (s *Store) Incr(key string) (int, error) {
	s.dataMutex.Lock()
	defer s.dataMutex.Unlock()

	obj, exists := s.data[key]
	if !exists {
		s.data[key] = createStringObject()
		s.data[key].Value.(*storage.String).SetData("1")
		s.data[key].version++
		return 1, nil
	}
	if obj.Type != StringType {
		return 0, Error.ERR_WRONGTYPE("STRING")
	}

	stringObj := obj.Value.(*storage.String)

	value, err := strconv.Atoi(string(stringObj.Data))
	if err != nil {
		return 0, Error.ERR_NAN
	}

	stringObj.SetData(strconv.Itoa(value + 1))
	s.data[key].version++
	return value + 1, nil
}

// ------------- End of String Operations ---------------
