package store

import (
	"errors"

	"github.com/codecrafters-io/redis-starter-go/internal/collections"
)

type SSetEntry struct {
	Score float64
	Value string
}

type SortedSet struct {
	SetList  *collections.DoublyLL[*SSetEntry]
	IndexMap map[*SSetEntry]int
	Length int
}

func (s *Store) CreateSortedSetObject(key string) *RedisObject {
	obj := &RedisObject{
		Type: SortedSetType,
		Value: &SortedSet{
			SetList:  &collections.DoublyLL[*SSetEntry]{},
			IndexMap: make(map[*SSetEntry]int),
			Length: 0,
		},
	}
	return obj
}

func (s *SortedSet) Add(score float64, value string) int {
	node := &SSetEntry{
		Score: score,
		Value: value,
	}

	_, exists := s.IndexMap[node]
	if exists {
		return s.Length
	}

	index := s.SetList.GetFirstNodeIndexMoreThan(node, func(a, b *SSetEntry) bool {
		return a.Score > b.Score || (a.Score == b.Score && a.Value > b.Value)
	})
	s.SetList.PushAtIndex(node, index)
	s.IndexMap[node] = index
	s.Length ++

	return s.Length
}












func (s *Store) ZADD(key string, score float64, value string) (int, error) {
	s.dataMutex.Lock()
	defer s.dataMutex.Unlock()

	redisObj, exists := s.data[key]
	if exists && redisObj.Type != SortedSetType {
		return 0, errors.New("ERR WrongType Operation. NOT A SORTED SET")
	}

	if !exists {
		s.data[key] = s.CreateSortedSetObject(key)
		redisObj = s.data[key]
	}

	sortedSet := redisObj.Value.(*SortedSet)
	n := sortedSet.Add(score, value)
	redisObj.version++
	return n, nil
}
