package store

import "time"

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
		return []string{}
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
			return nil, false
		}
		result = append(result, value)
	}

	return result, len(result) == count
}

func (s *Store) BLPop(key string, timeout float64) (string, bool) {
	s.dataMutex.Lock()
	timeoutDuration := time.Duration(timeout * float64(time.Second))

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

	if timeoutDuration <= 0 {
		// No timeout, block indefinitely
		value := <-ch
		return value, true
	}

	select {
	// block and wait for incoming value
	case value := <-ch:
		return value, true

	// timeout after specified duration. Remove from waiting list
	case <-time.After(timeoutDuration):
		s.dataMutex.Lock()

		waiters := s.waiting[key]

		for i, waiter := range waiters {
			if waiter == ch {
				waiters = append(waiters[:i], waiters[i+1:]...)
				break
			}
		}

		if len(waiters) == 0 {
			delete(s.waiting, key)
		} else {
			s.waiting[key] = waiters
		}

		s.dataMutex.Unlock()

		return "", false
	}
}

// ---------------- end of List Operations ----------------
