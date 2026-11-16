package store

import (
	"errors"

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
		if s.greater(curNode.Value, value) {
			s.SetList.PushAtIndex(value, index)
			pushed = true
		}
		curNode = curNode.Next
		s.IndexMap[curNode.Value] = index
		index++
	}
	if !pushed {
		s.SetList.PushBack(value)
		s.IndexMap[value] = index
	}

	return returnValue
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
