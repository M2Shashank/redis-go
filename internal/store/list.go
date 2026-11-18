package store

import (
	"errors"
	"time"

	"github.com/codecrafters-io/redis-starter-go/internal/collections"
	"github.com/codecrafters-io/redis-starter-go/internal/store/storage"
)

func createListObject() *RedisObject {
	obj := &RedisObject{
		Type: ListType,
		Value: &storage.List{
			Entries: &collections.DoublyLL[string]{},
			Waiters: []chan string{},
		},
	}
	return obj
}

// ------------------------------------------------------

func (s *Store) RPush(key string, values []string) (int, error) {
	s.dataMutex.Lock()
	defer s.dataMutex.Unlock()

	listObj, exists := s.data[key]
	if !exists {
		s.data[key] = createListObject()
		listObj = s.data[key]
	}
	if exists && listObj.Type != ListType {
		return 0, errors.New("ERR WrongType. NOT A LIST")
	}

	list := listObj.Value.(*storage.List)
	var result = list.Entries.Length + len(values)

	for _, value := range values {
		list.Entries.PushBack(value)
	}
	listObj.version++

	list.ServeWaiters()
	return result, nil
}

func (s *Store) LPush(key string, values []string) (int, error) {
	s.dataMutex.Lock()
	defer s.dataMutex.Unlock()

	listObj, exists := s.data[key]
	if !exists {
		s.data[key] = createListObject()
		listObj = s.data[key]
	}
	if exists && listObj.Type != ListType {
		return 0, errors.New("ERR WrongType. NOT A LIST")
	}

	list := listObj.Value.(*storage.List)
	var result = list.Entries.Length + len(values)

	for _, value := range values {
		list.Entries.PushFront(value)
	}
	listObj.version++

	list.ServeWaiters()
	return result, nil
}

func (s *Store) LRange(key string, start int, stop int) []string {
	s.dataMutex.RLock()
	defer s.dataMutex.RUnlock()

	obj, exists := s.data[key]
	if !exists || obj.Type != ListType {
		return []string{}
	}

	list := obj.Value.(*storage.List).Entries

	if start < 0 {
		start = list.Length + start
	}
	if stop < 0 {
		stop = list.Length + stop
	}
	start = max(start, 0)
	stop = min(stop, list.Length-1)

	if start > stop {
		return []string{}
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

func (s *Store) LLen(key string) (int, error) {
	s.dataMutex.RLock()
	defer s.dataMutex.RUnlock()

	obj, exists := s.data[key]
	if !exists {
		return 0, nil
	}
	if obj.Type != ListType {
		return 0, errors.New("ERR WrongType. NOT A LIST")
	}

	return obj.Value.(*storage.List).Entries.Length, nil
}

func (s *Store) LPop(key string) (string, error) {
	s.dataMutex.Lock()
	defer s.dataMutex.Unlock()

	obj, exists := s.data[key]
	if !exists {
		return "", nil
	}
	if obj.Type != ListType {
		return "", errors.New("ERR WrongType. NOT A LIST")
	}

	list := obj.Value.(*storage.List)
	value, err := list.Entries.PopFront()
	obj.version++

	return value, err
}

func (s *Store) LPopN(key string, count int) ([]string, error) {
	s.dataMutex.Lock()
	defer s.dataMutex.Unlock()

	obj, exists := s.data[key]
	if !exists {
		return nil, nil
	}
	if obj.Type != ListType {
		return []string{}, errors.New("ERR WrongType. NOT A LIST")
	}

	list := obj.Value.(*storage.List)
	var result []string = make([]string, 0, count)

	for i := 0; i < count; i++ {
		value, err := list.Entries.PopFront()
		if err != nil {
			break // Stop popping if the list is empty
		}
		result = append(result, value)
	}
	obj.version++

	return result, nil
}

func (s *Store) BLPop(key string, timeout float64) (string, error) {
	s.dataMutex.Lock()
	timeoutDuration := time.Duration(timeout * float64(time.Second))

	obj, exists := s.data[key]
	if !exists {
		s.data[key] = createListObject()
		obj = s.data[key]
	}
	if obj.Type != ListType {
		s.dataMutex.Unlock()
		return "", errors.New("ERR WrongType. NOT A LIST")
	}

	list := obj.Value.(*storage.List)
	if list.Entries.Length > 0 {
		value, _ := list.Entries.PopFront()
		obj.version++
		s.dataMutex.Unlock()
		return value, nil
	}

	// each waiter has its channel
	ch := make(chan string, 1)
	list.Waiters = append(list.Waiters, ch)

	s.dataMutex.Unlock()

	incrVersion := func() {
		s.dataMutex.Lock()
		obj.version++
		s.dataMutex.Unlock()
	}

	if timeoutDuration <= 0 {
		// No timeout, block indefinitely
		value := <-ch
		incrVersion()
		return value, nil
	}

	select {
	// block and wait for incoming value
	case value := <-ch:
		incrVersion()
		return value, nil

	// timeout after specified duration. Remove from waiting list
	case <-time.After(timeoutDuration):
		s.dataMutex.Lock()

		for i, waiter := range list.Waiters {
			if waiter == ch {
				list.Waiters = append(list.Waiters[:i], list.Waiters[i+1:]...)
				break
			}
		}

		s.dataMutex.Unlock()

		return "", nil
	}
}
