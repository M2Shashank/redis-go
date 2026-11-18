package storage

import (
	"github.com/codecrafters-io/redis-starter-go/internal/collections"
)

type SortedSet struct {
	// linked list of values. Map of value to score
	SetList  *collections.DoublyLL[string]
	ScoreMap map[string]float64
	IndexMap map[string]int
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

	if left > right || left >= len || right < 0 {
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