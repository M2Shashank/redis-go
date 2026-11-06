package store

import (
	"strconv"
	"errors"
)

// ----------------- String Operations ------------------
func (s *Store) Get(key string) (string, error) {
	s.dataMutex.RLock()
	defer s.dataMutex.RUnlock()

	obj, exists := s.data[key]
	if !exists {
		return "", nil
	}
	if obj.Type != StringType {
		return "", errors.New("ERR WrongType. NOT A STRING")
	}

	return obj.Value.(string), nil
}

func (s *Store) Set(key string, value string) error {
	s.dataMutex.Lock()
	defer s.dataMutex.Unlock()

	obj, exists := s.data[key]
	if !exists {
		s.data[key] = &RedisObject{StringType, value}
		return nil
	}

	if exists && obj.Type != StringType {
		return errors.New("ERR WrongType. NOT A STRING")
	}

	s.data[key].Value = value
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
		s.data[key] = &RedisObject{StringType, "1"}
		return 1, nil
	}
	if obj.Type != StringType {
		return 0, errors.New("ERR WrongType. NOT A STRING")
	}

	value, err := strconv.Atoi(obj.Value.(string))
	if err != nil {
		return 0, errors.New("ERR value is not an integer or out of range")
	}

	s.data[key].Value = strconv.Itoa(value + 1)
	return value + 1, nil
}

// ------------- End of String Operations ---------------