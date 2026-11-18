package store

import (
	"errors"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/internal/collections"
	"github.com/codecrafters-io/redis-starter-go/internal/store/storage"
)

func createSortedSetObject() *RedisObject {
	obj := &RedisObject{
		Type: SortedSetType,
		Value: &storage.SortedSet{
			SetList:  &collections.DoublyLL[string]{},
			ScoreMap: make(map[string]float64),
			IndexMap: make(map[string]int),
		},
		version: 0,
	}
	return obj
}


func (s *Store) ZADD(key string, score float64, value string) (bool, error) {
	s.dataMutex.Lock()
	defer s.dataMutex.Unlock()

	redisObj, exists := s.data[key]
	if exists && redisObj.Type != SortedSetType {
		return false, errors.New("ERR WrongType Operation. NOT A SORTED SET")
	}

	if !exists {
		s.data[key] = createSortedSetObject()
		redisObj = s.data[key]
	}

	sortedSet := redisObj.Value.(*storage.SortedSet)
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

	index, exists := redisObj.Value.(*storage.SortedSet).IndexMap[value]
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

	sortedSet := SetObj.Value.(*storage.SortedSet)
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

	sortedSet := SetObj.Value.(*storage.SortedSet)
	return sortedSet.SetList.Length
}

func (s *Store) ZSCORE(key, value string) (string, error) {
	s.dataMutex.RLock()
	defer s.dataMutex.RUnlock()

	SetObj, exists := s.data[key]
	if !exists || SetObj.Type != SortedSetType {
		return "", errors.New("ERR Key does not exist")
	}

	sortedSet := SetObj.Value.(*storage.SortedSet)
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

	sortedSet := SetObj.Value.(*storage.SortedSet)
	done := sortedSet.RemoveValue(value)
	SetObj.version++

	return done, nil
}