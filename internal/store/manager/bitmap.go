package store

import (
	Error "github.com/codecrafters-io/redis-starter-go/internal/errors"
	"strings"

	store_model "github.com/codecrafters-io/redis-starter-go/internal/store/models"
)

func (s *Store) SetBit(key string, offset int, value int) (int, error) {
	redisObj, exists := s.data[key]
	if !exists {
		s.data[key] = createStringObject()
		redisObj = s.data[key]
	}
	if redisObj.Type != StringType {
		return 0, Error.ERR_WRONGTYPE("STRING")
	}

	stringObj := redisObj.Value.(*store_model.String)
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

	stringObj := redisObj.Value.(*store_model.String)
	byteIndex, bitIndex := offset/8, offset%8
	return stringObj.GetBitVal(byteIndex, bitIndex)
}

func (s *Store) StrLen(key string) int {
	redisObj, exists := s.data[key]
	if !exists || redisObj.Type != StringType {
		return 0
	}

	stringObj := redisObj.Value.(*store_model.String)
	return stringObj.Length
}

func (s *Store) BitCount(key string, start, end int) int {
	redisObj, exists := s.data[key]
	if !exists || redisObj.Type != StringType {
		return 0
	}

	stringObj := redisObj.Value.(*store_model.String)
	count := stringObj.CountSetBits(start, end)
	return count
}

func (s *Store) BitopHandler(operation string, destKey string, keys []string) int {
	destObj, exists := s.data[destKey]
	if !exists {
		s.data[destKey] = createStringObject()
		destObj = s.data[destKey]
	}

	var missing bool = false
	validKeys := []*store_model.String{}
	for _, key := range keys {
		redisObj, exists := s.data[key]
		if exists && redisObj.Type == StringType {
			data := redisObj.Value.(*store_model.String)
			validKeys = append(validKeys, data)
		} else {
			missing = true
		}
	}

	var lenDest int
	switch strings.ToUpper(operation) {
	case "AND":
		lenDest = destObj.Value.(*store_model.String).AndAll(validKeys)
		if missing {
			clear(destObj.Value.(*store_model.String).Data)
		}

	case "OR":
		lenDest = destObj.Value.(*store_model.String).OrAll(validKeys)
	}

	return lenDest
}
