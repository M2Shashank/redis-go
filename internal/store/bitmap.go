package store

import (
	"errors"
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
	if byteIndex + 1 > s.Length {
		s.Data = append(
			s.Data, 
			make([]byte, byteIndex - s.Length + 1)...,
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
	if byteIndex + 1 > s.Length {
		return 0
	}
	
	return int((s.Data[byteIndex] >> (7 - bitIndex)) & 1)
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
	byteIndex, bitIndex := offset / 8, offset % 8
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
	byteIndex, bitIndex := offset / 8, offset % 8
	return stringObj.GetBitVal(byteIndex, bitIndex)
}