package store

import (
	"strconv"
	"strings"
	"errors"
)

type StreamEntry struct {
	millisecondsTime int64
	sequenceNumber   int64
	ID               string
	Fields           map[string]string
}

// stream object looks like -- map[string]*Stream
// data[key] => &RedisObject{ StreamType, &Stream{ Entries: &DoublyLL[StreamEntry]{} }}
type Stream struct {
	Entries *DoublyLL[*StreamEntry]
}

func (s *Stream) validate(entry *StreamEntry) error {
	lastEntry, exists := s.Entries.Back()
	if !exists {
		return nil
	}

	if entry.millisecondsTime < lastEntry.millisecondsTime {
		return errors.New("ERR The ID specified in XADD is equal or smaller than the target stream top item")
	}

	if entry.millisecondsTime == lastEntry.millisecondsTime && entry.sequenceNumber <= lastEntry.sequenceNumber {
		return errors.New("ERR The ID specified in XADD is equal or smaller than the target stream top item")
	}

	return nil
}

func (s *Store) CreateStreamEntry(ID string, fields map[string]string) (*StreamEntry, error) {
	var nums []string = strings.Split(ID, "-")
	if len(nums) != 2 {
		return nil, errors.New("ERR Invalid Arguments for XADD command")
	}

	var millisecondsTime, err1 = strconv.ParseInt(nums[0], 10, 64)
	var sequenceNumber, err2 = strconv.ParseInt(nums[1], 10, 64)
	if err1 != nil || err2 != nil {
		return nil, errors.New("ERR Invalid Arguments for XADD command")
	}
	if millisecondsTime < 0 || sequenceNumber < 0 {
		return nil, errors.New("ERR Invalid Arguments for XADD command")
	}
	if millisecondsTime == 0 && sequenceNumber == 0 {
		return nil, errors.New("ERR The ID specified in XADD must be greater than 0-0")
	}

	return &StreamEntry{
		ID:               ID,
		millisecondsTime: millisecondsTime,
		sequenceNumber:   sequenceNumber,
		Fields:           fields,
	}, nil
}

func (s *Store) createStreamObject() *RedisObject {
	obj := &RedisObject{
		Type: StreamType,
		Value: &Stream{
			Entries: &DoublyLL[*StreamEntry]{},
		},
	}
	return obj
}

func (s *Store) AddStreamEntry(key string, entry *StreamEntry) error {
	redisObj, exists := s.data[key]
	if exists && redisObj.Type != StreamType {
		return errors.New("ERR WrongType Operation. NOT A STREAM")
	}

	if !exists {
		s.data[key] = s.createStreamObject()
		redisObj = s.data[key]
	}

	stream := redisObj.Value.(*Stream)

	valid := stream.validate(entry)
	if valid != nil {
		return valid
	}

	stream.Entries.PushBack(entry)
	return nil
}
