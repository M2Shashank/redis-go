package store

import (
	"sync"
	//"time"
)

type Store struct {
	data      map[string]*RedisObject
	dataMutex sync.RWMutex

	waiting map[string][]chan string
}

func NewStore() *Store {
	return &Store{
		data:    make(map[string]*RedisObject),
		waiting: make(map[string][]chan string),
	}
}

// to be used when dataMutex is already locked
func (s *Store) serveWaiters(key string, list *DoublyLL) {
	queue, exists := s.waiting[key]
	if !exists || len(queue) == 0 {
		return
	}

	for len(queue) > 0 && list.Length > 0 {
		var ch chan string = queue[0]
		queue = queue[1:]

		value, _ := list.PopFront()
		ch <- value
	}

	if len(queue) == 0 {
		delete(s.waiting, key)
	} else {
		s.waiting[key] = queue
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
func (s *Store) RPush(key string, values []string) int {
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

	list := listObj.Value.(*DoublyLL)
	var result = list.Length + len(values)

	for _, value := range values {
		list.PushBack(value)
	}
	
	s.serveWaiters(key, list)
	return result
}

func (s *Store) LPush(key string, values []string) int {
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

	list := listObj.Value.(*DoublyLL)
	var result = list.Length + len(values)

	for _, value := range values {
		list.PushFront(value)
	}
	
	s.serveWaiters(key, list)
	return result
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

func (s *Store) LPopN(key string, count int) ([]string, bool) {
	count = min(count, s.LLen(key))
	var result []string = make([]string, 0, count)

	for i := 0; i < count; i++ {
		value, ok := s.LPop(key)
		if !ok {
			break
		}
		result = append(result, value)
	}

	return result, len(result) == count
}

func (s *Store) BLPop(key string, timeout int) (string, bool) {
	s.dataMutex.Lock()

	obj, exists := s.data[key]
	if exists {
		if obj.Type != ListType {
			s.dataMutex.Unlock()
			return "", false
		}

		list := obj.Value.(*DoublyLL)
		if list.Length > 0 {
			value, _ := list.PopFront()
			s.dataMutex.Unlock()
			return value, true
		}
	}

	// each waiter has its channel
	ch := make(chan string, 1)

	s.waiting[key] = append(s.waiting[key], ch)

	s.dataMutex.Unlock()

	// block and wait for incoming value
	value := <-ch

	return value, true
}

// ---------------- end of List Operations ----------------
