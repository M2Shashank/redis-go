package model

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/internal/collections"
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
	Key     string
	Entries *collections.DoublyLL[*StreamEntry]

	// list of channels waiting for new entries in this stream. BXREAD
	// Why list of channels ? more than 1 client can be waiting for new entries in the same stream
	Notify []chan *StreamNotification
}

type StreamNotification struct {
    Key   string
    Field *StreamEntry
}

// For Response Conversion Only. Not used in Store
// Used in XREAD and BXREAD commands.
type ModStream struct {
	Key     string
	Entries []*StreamEntry
}


// --------------- Stream Methods ----------------
func (s *Stream) Validate(entry *StreamEntry) error {
	lastEntry, err := s.Entries.Back()
	if err != nil {
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

func (s *Stream) CreateStreamEntry(ID string, fields map[string]string) (*StreamEntry, error) {
	var nums []string = strings.Split(ID, "-")
	var millisecondsTime, sequenceNumber int64
	var err1, err2 error

	if len(nums) != 2 && ID != "*" {
		return nil, errors.New("ERR Invalid Arguments for XADD command")
	}

	if ID == "*" {
		millisecondsTime = time.Now().UnixMilli()
	} else {
		millisecondsTime, err1 = strconv.ParseInt(nums[0], 10, 64)
		if err1 != nil {
			return nil, errors.New("ERR Invalid Arguments for XADD command")
		}
	}

	if ID == "*" || nums[1] == "*" {
		var lastEntry, err = s.Entries.Back()
		if err != nil {
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

	newID := strconv.FormatInt(millisecondsTime, 10) + "-" + strconv.FormatInt(sequenceNumber, 10)

	var entry *StreamEntry = &StreamEntry{
		ID:               newID,
		millisecondsTime: millisecondsTime,
		sequenceNumber:   sequenceNumber,
		Fields:           fields,
	}

	var err error = s.Validate(entry)
	if err != nil {
		return nil, err
	}

	return entry, nil
}

func (s *Stream) FilterEntries(startMST, startSeq, endMST, endSeq int64) []*StreamEntry {
	var result []*StreamEntry = make([]*StreamEntry, 0, 1)

	s.Entries.ForEach(func(curr *StreamEntry) {
		afterStart := curr.millisecondsTime > startMST || (curr.millisecondsTime == startMST && curr.sequenceNumber >= startSeq)
		beforeEnd := curr.millisecondsTime < endMST || (curr.millisecondsTime == endMST && (endSeq != -1 && curr.sequenceNumber <= endSeq || endSeq == -1))
		if endMST == -1 {
			beforeEnd = true
		}

		if afterStart && beforeEnd {
			result = append(result, curr)
		}
	})

	return result
}

func (s *Stream) LastEntry() (*StreamEntry, error) {
	return s.Entries.Back()
}

// Should be called with dataMutex locked
func (s *Stream) RemoveChannel(ch chan *StreamNotification) {
	var newNotify []chan *StreamNotification
	for _, c := range s.Notify {
		if c != ch {
			newNotify = append(newNotify, c)
		}
	}
	s.Notify = newNotify
}

// Should be called with dataMutex locked
func (s *Stream) NotifyNewEntry(entry *StreamEntry) {
	for _, ch := range s.Notify {
		ch <- &StreamNotification{
			Key:   s.Key,
			Field: entry,
		}
	}
}


//* Notification Helpers

func (n *StreamNotification) CheckValidNotification(mstMp map[string]int64, seqMp map[string]int64, firstMst int64) bool {
	return n.Field.millisecondsTime > firstMst || (n.Field.millisecondsTime == mstMp[n.Key] && n.Field.sequenceNumber > seqMp[n.Key])
}

func (s *StreamEntry) IsIDEqual(mst, seq int64) bool {
	return s.millisecondsTime == mst && s.sequenceNumber == seq
}