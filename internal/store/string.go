package store

import (
	"strconv"
	"errors"
)

type String struct {
	Data []byte
	Length int // length of the string in bytes
}

func createStringObject(value string) *RedisObject {
	obj := &RedisObject{
		Type: StringType,
		Value: &String{
			Data: []byte(value),
			Length: len(value),
		},
		version: 0,
	}
	return obj
}

func (s *String) SetData(value string) {
	s.Data = []byte(value)
	s.Length = len(value)
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
		return "", errors.New("ERR WrongType. NOT A STRING")
	}

	var ans string = string(obj.Value.(*String).Data)
	return ans, nil
}

func (s *Store) Set(key string, value string) error {
	s.dataMutex.Lock()
	defer s.dataMutex.Unlock()

	obj, exists := s.data[key]
	if !exists {
		s.data[key] = createStringObject(value)
		return nil
	}

	if exists && obj.Type != StringType {
		return errors.New("ERR WrongType. NOT A STRING")
	}

	obj.Value.(*String).SetData(value)
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
		s.data[key] = &RedisObject{StringType, 1, "1"}
		return 1, nil
	}
	if obj.Type != StringType {
		return 0, errors.New("ERR WrongType. NOT A STRING")
	}

	stringObj := obj.Value.(*String)

	value, err := strconv.Atoi(string(stringObj.Data))
	if err != nil {
		return 0, errors.New("ERR value is not an integer or out of range")
	}

	stringObj.SetData(strconv.Itoa(value + 1))
	s.data[key].version++
	return value + 1, nil
}

// ------------- End of String Operations ---------------