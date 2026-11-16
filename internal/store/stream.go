package store

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
	Notify []chan struct {
		key   string
		field *StreamEntry
	}
}

// For Response Conversion Only. Not used in Store
// Used in XREAD and BXREAD commands.
type ModStream struct {
	Key     string
	Entries []*StreamEntry
}

func createStreamObject(key string) *RedisObject {
	obj := &RedisObject{
		Type: StreamType,
		Value: &Stream{
			Key:     key,
			Entries: collections.NewDoublyLL[*StreamEntry](),
			Notify: []chan struct {
				key   string
				field *StreamEntry
			}{},
		},
	}
	return obj
}

// --------------- Stream Methods ----------------
func (s *Stream) validate(entry *StreamEntry) error {
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

func (s *Stream) createStreamEntry(ID string, fields map[string]string) (*StreamEntry, error) {
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

	var err error = s.validate(entry)
	if err != nil {
		return nil, err
	}

	return entry, nil
}

func (s *Stream) LastEntry() (*StreamEntry, error) {
	return s.Entries.Back()
}

// Should be called with dataMutex locked
func (s *Stream) removeChannel(ch chan struct {
	key   string
	field *StreamEntry
}) {
	var newNotify []chan struct {
		key   string
		field *StreamEntry
	}
	for _, c := range s.Notify {
		if c != ch {
			newNotify = append(newNotify, c)
		}
	}
	s.Notify = newNotify
}

// Should be called with dataMutex locked
func (s *Stream) notifyNewEntry(entry *StreamEntry) {
	for _, ch := range s.Notify {
		ch <- struct {
			key   string
			field *StreamEntry
		}{s.Key, entry}
	}
}

// -----------------------------------------------

func (s *Store) AddStreamEntry(key string, ID string, fields map[string]string) (string, error) {
	s.dataMutex.Lock()
	defer s.dataMutex.Unlock()

	redisObj, exists := s.data[key]
	if exists && redisObj.Type != StreamType {
		return "", errors.New("ERR WrongType Operation. NOT A STREAM")
	}

	if !exists {
		s.data[key] = createStreamObject(key)
		redisObj = s.data[key]
	}

	stream := redisObj.Value.(*Stream)
	entry, err := stream.createStreamEntry(ID, fields)
	if err != nil {
		return "", err
	}

	stream.Entries.PushBack(entry)
	stream.notifyNewEntry(entry)
	redisObj.version++
	return entry.ID, nil
}

func (s *Store) XRange(key string, startMST int64, startSeq int64, endMST int64, endSeq int64) ([]*StreamEntry, error) {
	s.dataMutex.RLock()
	defer s.dataMutex.RUnlock()

	redisObj, exists := s.data[key]
	if !exists {
		return []*StreamEntry{}, nil
	}

	if redisObj.Type != StreamType {
		return nil, errors.New("ERR WrongType Operation. NOT A STREAM")
	}

	var result []*StreamEntry = make([]*StreamEntry, 0, 10)
	streamEntries := redisObj.Value.(*Stream).Entries

	streamEntries.ForEach(func(curr *StreamEntry) {
		afterStart := curr.millisecondsTime > startMST || (curr.millisecondsTime == startMST && curr.sequenceNumber >= startSeq)
		beforeEnd := curr.millisecondsTime < endMST || (curr.millisecondsTime == endMST && (endSeq != -1 && curr.sequenceNumber <= endSeq || endSeq == -1))
		if endMST == -1 {
			beforeEnd = true
		}

		if afterStart && beforeEnd {
			result = append(result, curr)
		}
	})
	return result, nil
}

func (s *Store) XReadStreams(keys []string, mst []int64, seq []int64) ([]*ModStream, bool) {
	s.dataMutex.RLock()
	defer s.dataMutex.RUnlock()

	var response []*ModStream = make([]*ModStream, 0, len(keys))
	for i := 0; i < len(keys); i++ {
		if mst[i] == -1 && seq[i] == -1 {
			continue
		}

		entries, err := s.XRange(keys[i], mst[i], seq[i], -1, -1)
		if err != nil {
			return nil, false
		}

		if len(entries) != 0 && entries[0].millisecondsTime == mst[i] && entries[0].sequenceNumber == seq[i] {
			entries = entries[1:]
		}

		if len(entries) == 0 {
			continue
		}

		response = append(response, &ModStream{
			Key:     keys[i],
			Entries: entries,
		})
	}

	if len(response) == 0 {
		return nil, true
	}
	return response, false
}

// Blocking XReadStreams to be implemented
// if a call to BXREAD is made, it will create a channel of key and streamEntry and add it to Notify of all its keys streams
// Every time a new entry is made to a stream , it will notify all the channels in the Notify list.
// If an item is added to a channel, BXREAD will take the item and check if its id is strictly greater than its mst and seq.
// If it is not greater, continue waiting.
// If it is greater, return the entry and remove the channel from the Notify list of all its keys streams.
// If the timeout expires, return nil and remove the channel from the Notify list of all its keys streams.
func (s *Store) BXReadStreams(keys []string, mst []int64, seq []int64, expiryTime int64) ([]*ModStream, error) {
	var expiryTimeMS time.Duration = time.Duration(expiryTime) * time.Millisecond
	type pair struct {
		mst, seq int64
	}
	var mp map[string]pair = make(map[string]pair, len(keys))
	for i := range keys {
		mp[keys[i]] = pair{mst[i], seq[i]}
	}

	var result []*ModStream = make([]*ModStream, 0, 1)

	waiterChannel := make(chan struct {
		key   string
		field *StreamEntry
	}, 1)

	s.dataMutex.Lock()
	for i := 0; i < len(keys); i++ {
		redisObj, exists := s.data[keys[i]]
		if !exists {
			s.data[keys[i]] = createStreamObject(keys[i])
			redisObj = s.data[keys[i]]
		}
		if redisObj.Type != StreamType {
			return nil, errors.New("ERR WrongType Operation. NOT A STREAM")
		}

		var notifyChList = &redisObj.Value.(*Stream).Notify
		*notifyChList = append(*notifyChList, waiterChannel)
	}
	s.dataMutex.Unlock()

	// func to remove the channel from the Notify list of all its keys streams
	removeChFromAll := func() {
		for i := 0; i < len(keys); i++ {
			redisObj, _ := s.data[keys[i]]
			redisObj.Value.(*Stream).removeChannel(waiterChannel)
		}
	}

	if expiryTimeMS == 0 {
		msg := <-waiterChannel
		removeChFromAll()
		result = append(result, &ModStream{
			Key:     msg.key,
			Entries: []*StreamEntry{msg.field},
		})
		return result, nil
	}

	timer := time.NewTimer(expiryTimeMS)
	for {
		select {
		case msg := <-waiterChannel:
			if msg.field.millisecondsTime > mst[0] || (msg.field.millisecondsTime == mp[msg.key].mst && msg.field.sequenceNumber > mp[msg.key].seq) {
				result = append(result, &ModStream{
					Key:     msg.key,
					Entries: []*StreamEntry{msg.field},
				})

				removeChFromAll()
				return result, nil
			}

		case <-timer.C:
			removeChFromAll()
			return nil, nil
		}
	}
}
