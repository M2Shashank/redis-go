package store

import (
	"errors"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/internal/collections"
)

type SortedSet struct {
	// linked list of values. Map of value to score
	SetList  *collections.DoublyLL[string]
	ScoreMap map[string]float64
	IndexMap map[string]int
}

func (s *Store) CreateSortedSetObject(key string) *RedisObject {
	obj := &RedisObject{
		Type: SortedSetType,
		Value: &SortedSet{
			SetList:  &collections.DoublyLL[string]{},
			ScoreMap: make(map[string]float64),
			IndexMap: make(map[string]int),
		},
		version: 0,
	}
	return obj
}

func (s *SortedSet) greater(s1, s2 string) bool {
	if s.ScoreMap[s1] == s.ScoreMap[s2] {
		return s1 > s2
	}
	return s.ScoreMap[s1] > s.ScoreMap[s2]
}

func (s *SortedSet) rebuildIndexMap() {
	s.IndexMap = make(map[string]int)

	index := 0
	for node := s.SetList.Head; node != nil; node = node.Next {
		s.IndexMap[node.Value] = index
		index++
	}
}

func (s *SortedSet) Add(score float64, value string) bool {
	var returnValue bool = true

	_, exists := s.ScoreMap[value]
	if exists {
		if score == s.ScoreMap[value] {
			return false // no change
		}

		returnValue = false
		s.SetList.PopBasedOnValue(value, func(s1, s2 string) bool {
			return s1 == s2
		})
	}
	s.ScoreMap[value] = score

	curNode := s.SetList.Head
	index, pushed := 0, false
	for curNode != nil {
		if !pushed && s.greater(curNode.Value, value) {
			s.SetList.PushAtIndex(value, index)
			pushed = true
			break
		}
		s.IndexMap[curNode.Value] = index
		curNode = curNode.Next
		index++
	}
	if !pushed {
		s.SetList.PushBack(value)
		s.IndexMap[value] = index
	}
	s.rebuildIndexMap()

	return returnValue
}

func (s *SortedSet) GetRange(left, right int) []string {
	var len int = s.SetList.Length
	if left < 0 {
		left = len + left
	}
	if right < 0 {
		right = len + right
	}

	if left >= right || left >= len || right < 0 {
		return []string{}
	}

	var listEntries []string = []string{}
	var index int = 0

	cur := s.SetList.Head
	for index < left && cur != nil {
		cur = cur.Next
		index++
	}

	for index <= right && cur != nil {
		listEntries = append(listEntries, cur.Value)
		cur = cur.Next
		index++
	}

	return listEntries
}

func (s *SortedSet) RemoveValue(value string) bool {
	_, exists := s.ScoreMap[value]
	if !exists {
		return false
	}

	s.SetList.PopBasedOnValue(value, func(s1, s2 string) bool {
		return s1 == s2
	})
	delete(s.ScoreMap, value)
	s.rebuildIndexMap()
	return true
}

func (s *Store) ZADD(key string, score float64, value string) (bool, error) {
	s.dataMutex.Lock()
	defer s.dataMutex.Unlock()

	redisObj, exists := s.data[key]
	if exists && redisObj.Type != SortedSetType {
		return false, errors.New("ERR WrongType Operation. NOT A SORTED SET")
	}

	if !exists {
		s.data[key] = s.CreateSortedSetObject(key)
		redisObj = s.data[key]
	}

	sortedSet := redisObj.Value.(*SortedSet)
	added := sortedSet.Add(score, value)
	redisObj.version++
	return added, nil
}

func (s *Store) ZRANK(key, value string) (int, error) {
	s.dataMutex.RLock()
	defer s.dataMutex.RUnlock()

	redisObj, exists := s.data[key]
	if !exists {
		return 0, errors.New("ERR Key does not exist")
	}

	index, exists := redisObj.Value.(*SortedSet).IndexMap[value]
	if !exists {
		return 0, errors.New("ERR Value does not exist in the sorted set")
	}

	return index, nil
}

func (s *Store) ZRANGE(key string, left, right int) []string {
	s.dataMutex.RLock()
	defer s.dataMutex.RUnlock()

	SetObj, exists := s.data[key]
	if !exists || SetObj.Type != SortedSetType {
		return []string{}
	}

	sortedSet := SetObj.Value.(*SortedSet)
	listRange := sortedSet.GetRange(left, right)
	return listRange
}

func (s *Store) ZCARD(key string) int {
	s.dataMutex.RLock()
	defer s.dataMutex.RUnlock()

	SetObj, exists := s.data[key]
	if !exists || SetObj.Type != SortedSetType {
		return 0
	}

	sortedSet := SetObj.Value.(*SortedSet)
	return sortedSet.SetList.Length
}

func (s *Store) ZSCORE(key, value string) (string, error) {
	s.dataMutex.RLock()
	defer s.dataMutex.RUnlock()

	SetObj, exists := s.data[key]
	if !exists || SetObj.Type != SortedSetType {
		return "", errors.New("ERR Key does not exist")
	}

	sortedSet := SetObj.Value.(*SortedSet)
	score, exists := sortedSet.ScoreMap[value]
	if !exists {
		return "", errors.New("ERR Value does not exist in the sorted set")
	}

	return strconv.FormatFloat(score, 'f', -1, 64), nil
}

func (s *Store) ZREM(key, value string) (bool, error) {
	s.dataMutex.Lock()
	defer s.dataMutex.Unlock()

	SetObj, exists := s.data[key]
	if !exists || SetObj.Type != SortedSetType {
		return false, errors.New("ERR Key does not exist")
	}

	sortedSet := SetObj.Value.(*SortedSet)
	done := sortedSet.RemoveValue(value)
	SetObj.version++

	return done, nil
}