package store

import (
	"errors"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/store/storage"
)

func (s *Store) SetBit(key string, offset int, value int) (int, error) {
	redisObj, exists := s.data[key]
	if !exists {
		s.data[key] = createStringObject()
		redisObj = s.data[key]
	}
	if redisObj.Type != StringType {
		return 0, errors.New("ERR WrongType. NOT A STRING")
	}

	stringObj := redisObj.Value.(*storage.String)
	byteIndex, bitIndex := offset/8, offset%8
	ans := stringObj.SetBitVal(byteIndex, bitIndex, value)
	redisObj.version++
	return ans, nil
}

func (s *Store) GetBit(key string, offset int) int {
	redisObj, exists := s.data[key]
	if !exists || redisObj.Type != StringType {
		return 0
	}

	stringObj := redisObj.Value.(*storage.String)
	byteIndex, bitIndex := offset/8, offset%8
	return stringObj.GetBitVal(byteIndex, bitIndex)
}

func (s *Store) StrLen(key string) int {
	redisObj, exists := s.data[key]
	if !exists || redisObj.Type != StringType {
		return 0
	}

	stringObj := redisObj.Value.(*storage.String)
	return stringObj.Length
}

func (s *Store) BitCount(key string, start, end int) int {
	redisObj, exists := s.data[key]
	if !exists || redisObj.Type != StringType {
		return 0
	}

	stringObj := redisObj.Value.(*storage.String)
	count := stringObj.CountSetBits(start, end)
	return count
}

func (s *Store) BitopHandler(operation string, destKey string, keys []string) (int, error) {
	if strings.ToUpper(operation) != "AND" && strings.ToUpper(operation) != "OR" {
		return 0, errors.New("ERR Unsupported BITOP operation")
	}

	destObj, exists := s.data[destKey]
	if !exists {
		s.data[destKey] = createStringObject()
		destObj = s.data[destKey]
	}

	var missing bool = false
	validKeys := []*storage.String{}
	for _, key := range keys {
		redisObj, exists := s.data[key]
		if exists && redisObj.Type == StringType {
			data := redisObj.Value.(*storage.String)
			validKeys = append(validKeys, data)
		} else {
			missing = true
		}
	}

	switch strings.ToUpper(operation) {
	case "AND":
		lenDest := destObj.Value.(*storage.String).AndAll(validKeys)
		if missing {
			clear(destObj.Value.(*storage.String).Data)
		}

		return lenDest, nil

	case "OR":
		lenDest := destObj.Value.(*storage.String).OrAll(validKeys)
		return lenDest, nil
	}

	return 0, errors.New("ERR Unsupported BITOP operation")
}
