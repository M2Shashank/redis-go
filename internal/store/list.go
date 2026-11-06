package store

import (
	"errors"
	"time"
)

type List struct {
	Entries *DoublyLL[string]
	Waiters []chan string
}

func (s *Store) createListObject() *RedisObject {
	obj := &RedisObject{
		Type: ListType,
		Value: &List{
			Entries: &DoublyLL[string]{},
			Waiters: []chan string{},
		},
	}
	return obj
}

// ----------------- List Operations ---------------------

// to be used when dataMutex is already locked
// Used for List operations Only. BLPOP
func (list *List) serveWaiters() {
	entries, waiters := list.Entries, list.Waiters

	for len(waiters) > 0 && entries.Length > 0 {
		var ch chan string = waiters[0]
		waiters = waiters[1:]

		value, _ := entries.PopFront()
		ch <- value
	}
}

// ------------------------------------------------------

func (s *Store) RPush(key string, values []string) (int, error) {
	s.dataMutex.Lock()
	defer s.dataMutex.Unlock()

	listObj, exists := s.data[key]
	if !exists {
		s.data[key] = s.createListObject()
		listObj = s.data[key]
	}
	if exists && listObj.Type != ListType {
		return 0, errors.New("ERR WrongType. NOT A LIST")
	}

	list := listObj.Value.(*List)
	var result = list.Entries.Length + len(values)

	for _, value := range values {
		list.Entries.PushBack(value)
	}

	list.serveWaiters()
	return result, nil
}

func (s *Store) LPush(key string, values []string) (int, error) {
	s.dataMutex.Lock()
	defer s.dataMutex.Unlock()

	listObj, exists := s.data[key]
	if !exists {
		s.data[key] = s.createListObject()
		listObj = s.data[key]
	}
	if exists && listObj.Type != ListType {
		return 0, errors.New("ERR WrongType. NOT A LIST")
	}

	list := listObj.Value.(*List)
	var result = list.Entries.Length + len(values)

	for _, value := range values {
		list.Entries.PushFront(value)
	}

	list.serveWaiters()
	return result, nil
}

func (s *Store) LRange(key string, start int, stop int) []string {
	s.dataMutex.RLock()
	defer s.dataMutex.RUnlock()

	obj, exists := s.data[key]
	if !exists || obj.Type != ListType {
		return []string{}
	}

	list := obj.Value.(*List).Entries

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

	return obj.Value.(*List).Entries.Length, nil
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

	list := obj.Value.(*List)
	value, err := list.Entries.PopFront()
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

	list := obj.Value.(*List)
	var result []string = make([]string, 0, count)

	for i := 0; i < count; i++ {
		value, err := list.Entries.PopFront()
		if err != nil {
			break // Stop popping if the list is empty
		}
		result = append(result, value)
	}

	return result, nil
}

func (s *Store) BLPop(key string, timeout float64) (string, error) {
	s.dataMutex.Lock()
	timeoutDuration := time.Duration(timeout * float64(time.Second))

	obj, exists := s.data[key]
	if !exists {
		s.data[key] = s.createListObject()
		obj = s.data[key]
	}
	if obj.Type != ListType {
		s.dataMutex.Unlock()
		return "", errors.New("ERR WrongType. NOT A LIST")
	}

	list := obj.Value.(*List)
	if list.Entries.Length > 0 {
		value, _ := list.Entries.PopFront()
		s.dataMutex.Unlock()
		return value, nil
	}

	// each waiter has its channel
	ch := make(chan string, 1)
	list.Waiters = append(list.Waiters, ch)

	s.dataMutex.Unlock()

	if timeoutDuration <= 0 {
		// No timeout, block indefinitely
		value := <-ch
		return value, nil
	}

	select {
	// block and wait for incoming value
	case value := <-ch:
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
