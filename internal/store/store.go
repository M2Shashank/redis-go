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

// ----------------- String Operations ------------------
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

// ------------- End of String Operations ---------------

// ----------------- List Operations ---------------------
func (s *Store) RPush(key string, value string) int {
	s.dataMutex.Lock()
	defer s.dataMutex.Unlock()

	listObj, exists := s.data[key]
	if !exists {
		s.data[key] = &RedisObject{ListType, &DoublyLL{}}
		listObj = s.data[key]
	}

	if exists && listObj.Type != ListType {
		return -1
	}

	listObj.Value.(*DoublyLL).PushBack(value)
	return listObj.Value.(*DoublyLL).Length
}

func (s *Store) LPush(key string, value string) int {
	s.dataMutex.Lock()
	defer s.dataMutex.Unlock()

	listObj, exists := s.data[key]
	if !exists {
		s.data[key] = &RedisObject{ListType, &DoublyLL{}}
		listObj = s.data[key]
	}

	if exists && listObj.Type != ListType {
		return -1
	}

	listObj.Value.(*DoublyLL).PushFront(value)
	return listObj.Value.(*DoublyLL).Length
}

func (s *Store) LRange(key string, start int, stop int) []string {
	s.dataMutex.RLock()
	defer s.dataMutex.RUnlock()

	obj, exists := s.data[key]
	if !exists || obj.Type != ListType {
		return nil
	}

	list := obj.Value.(*DoublyLL)

	if start < 0 {
		start = list.Length + start
	}
	if stop < 0 {
		stop = list.Length + stop
	}
	start = max(start, 0)
	stop = min(stop, list.Length-1)

	if start > stop {
		return nil
	}

	var result []string = make([]string, 0, stop-start+1)
	current := list.Head
	for i := 0; i < start; i++ {
		current = current.Next
	}

	for i := start; i <= stop; i++ {
		result = append(result, current.Value)
		current = current.Next
	}
	return result
}

func (s *Store) LLen(key string) int {
	s.dataMutex.RLock()
	defer s.dataMutex.RUnlock()

	obj, exists := s.data[key]
	if !exists || obj.Type != ListType {
		return 0
	}

	return obj.Value.(*DoublyLL).Length
}

func (s *Store) LPop(key string) (string, bool) {
	s.dataMutex.Lock()
	defer s.dataMutex.Unlock()

	obj, exists := s.data[key]
	if !exists || obj.Type != ListType {
		return "", false
	}

	list := obj.Value.(*DoublyLL)
	value, done := list.PopFront()
	return value, done
}
// ---------------- end of List Operations ----------------
