package store

import (
	"errors"
	"strconv"
	"strings"
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

func (s *Stream) createStreamEntry(ID string, fields map[string]string) (*StreamEntry, error) {
	var nums []string = strings.Split(ID, "-")
	if len(nums) != 2 {
		return nil, errors.New("ERR Invalid Arguments for XADD command")
	}

	var millisecondsTime, sequenceNumber int64
	var err1, err2 error

	millisecondsTime, err1 = strconv.ParseInt(nums[0], 10, 64)
	if err1 != nil {
		return nil, errors.New("ERR Invalid Arguments for XADD command")
	}

	if nums[1] == "*" {
		var lastEntry, exists = s.Entries.Back()
		if !exists {
			if millisecondsTime == 0 {
				sequenceNumber = 1
			} else {
				sequenceNumber = 0
			}
		} else {
			if millisecondsTime < lastEntry.millisecondsTime {
				return nil, errors.New("ERR The ID specified in XADD is equal or smaller than the target stream top item")
			} else if millisecondsTime == lastEntry.millisecondsTime {
				sequenceNumber = lastEntry.sequenceNumber + 1
			} else {
				sequenceNumber = 0
			}
		}
	} else {
		sequenceNumber, err2 = strconv.ParseInt(nums[1], 10, 64)
		if err2 != nil {
			return nil, errors.New("ERR Invalid Arguments for XADD command")
		}
	}

	if millisecondsTime < 0 || sequenceNumber < 0 {
		return nil, errors.New("ERR Invalid Arguments for XADD command")
	}
	if millisecondsTime == 0 && sequenceNumber == 0 {
		return nil, errors.New("ERR The ID specified in XADD must be greater than 0-0")
	}

	var entry *StreamEntry = &StreamEntry{
		ID:               ID,
		millisecondsTime: millisecondsTime,
		sequenceNumber:   sequenceNumber,
		Fields:           fields,
	}

	err := s.validate(entry)
	if err != nil {
		return nil, err
	}

	return entry, nil
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

func (s *Store) AddStreamEntry(key string, ID string, fields map[string]string) (string, error) {
	redisObj, exists := s.data[key]
	if exists && redisObj.Type != StreamType {
		return "", errors.New("ERR WrongType Operation. NOT A STREAM")
	}

	if !exists {
		s.data[key] = s.createStreamObject()
		redisObj = s.data[key]
	}

	stream := redisObj.Value.(*Stream)
	entry, err := stream.createStreamEntry(ID, fields)
	if err != nil {
		return "", err
	}

	stream.Entries.PushBack(entry)
	newID := strconv.FormatInt(entry.millisecondsTime, 10) + "-" + strconv.FormatInt(entry.sequenceNumber, 10)
	return newID, nil
}
