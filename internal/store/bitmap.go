package store

import (
	"errors"
	"strings"
)

// type String struct {
// 	Data []byte
// 	Length int // length of the string in bytes
// }

// func createStringObject() *RedisObject {
// 	obj := &RedisObject{
// 		Type: StringType,
// 		Value: &String{
// 			Data: []byte(""),
// 			Length: 0,
// 		},
// 		version: 0,
// 	}
// 	return obj
// }

func (s *String) SetBitVal(byteIndex, bitIndex, value int) int {
	if byteIndex+1 > s.Length {
		s.Data = append(
			s.Data,
			make([]byte, byteIndex-s.Length+1)...,
		)
		s.Length = byteIndex + 1
	}

	dataByte := &s.Data[byteIndex]
	prevBit := (*dataByte >> (7 - bitIndex)) & 1

	mask := byte(1 << (7 - bitIndex))
	if value == 1 {
		*dataByte |= mask
	} else {
		*dataByte &= (^mask)
	}

	return int(prevBit)
}

func (s *String) GetBitVal(byteIndex, bitIndex int) int {
	if byteIndex+1 > s.Length {
		return 0
	}

	return int((s.Data[byteIndex] >> (7 - bitIndex)) & 1)
}

func (s *String) CountSetBits(start, end int) int {
	end = min(end, s.Length-1)
	if end == -1 {
		end = s.Length - 1
	}
	if start < 0 || end >= s.Length || start > end {
		return 0
	}

	var count int = 0
	for i := start; i <= end; i++ {
		for j := range 8 {
			count += int((s.Data[i] >> j) & 1)
		}
	}
	return count
}

func (s *String) AndOne(o *String) {
	minLength := min(s.Length, o.Length)
	for i := range minLength {
		s.Data[i] &= o.Data[i]
	}

	if o.Length > s.Length {
		s.Data = append(s.Data, make([]byte, o.Length-s.Length)...)
		s.Length = o.Length
	} else {
		clear(s.Data[minLength:s.Length])
	}
}

func (s *String) AndAll(keys []*String) int {
	if len(keys) == 0 {
		s.Data = []byte{0}
		s.Length = 1
		return 1
	}

	s.Data = append([]byte{}, keys[0].Data...)
	s.Length = keys[0].Length

	for _, key := range keys {
		s.AndOne(key)
	}
	return s.Length
}

func (s *Store) SetBit(key string, offset int, value int) (int, error) {
	redisObj, exists := s.data[key]
	if !exists {
		s.data[key] = createStringObject()
		redisObj = s.data[key]
	}
	if redisObj.Type != StringType {
		return 0, errors.New("ERR WrongType. NOT A STRING")
	}

	stringObj := redisObj.Value.(*String)
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

	stringObj := redisObj.Value.(*String)
	byteIndex, bitIndex := offset/8, offset%8
	return stringObj.GetBitVal(byteIndex, bitIndex)
}

func (s *Store) StrLen(key string) int {
	redisObj, exists := s.data[key]
	if !exists || redisObj.Type != StringType {
		return 0
	}

	stringObj := redisObj.Value.(*String)
	return stringObj.Length
}

func (s *Store) BitCount(key string, start, end int) int {
	redisObj, exists := s.data[key]
	if !exists || redisObj.Type != StringType {
		return 0
	}

	stringObj := redisObj.Value.(*String)
	count := stringObj.CountSetBits(start, end)
	return count
}

func (s *Store) BitopHandler(operation string, destKey string, keys []string) (int, error) {
	switch strings.ToUpper(operation) {
	case "AND":
		validKeys := []*String{}
		for _, key := range keys {
			redisObj, exists := s.data[key]
			if exists && redisObj.Type == StringType {
				data := redisObj.Value.(*String)
				validKeys = append(validKeys, data)
			}
		}

		destObj, exists := s.data[destKey]
		if !exists {
			s.data[destKey] = createStringObject()
			destObj = s.data[destKey]
		}

		lenDest := destObj.Value.(*String).AndAll(validKeys)
		return lenDest, nil

	default:
		return 0, errors.New("ERR Unsupported BITOP operation")
	}
}
