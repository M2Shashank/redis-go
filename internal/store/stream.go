package store

import (
	"errors"
	"time"

	"github.com/codecrafters-io/redis-starter-go/internal/collections"
	"github.com/codecrafters-io/redis-starter-go/internal/store/storage"
)

func createStreamObject(key string) *RedisObject {
	obj := &RedisObject{
		Type: StreamType,
		Value: &storage.Stream{
			Key:     key,
			Entries: collections.NewDoublyLL[*storage.StreamEntry](),
			Notify: []chan *storage.StreamNotification{},
		},
	}
	return obj
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

	stream := redisObj.Value.(*storage.Stream)
	entry, err := stream.CreateStreamEntry(ID, fields)
	if err != nil {
		return "", err
	}

	stream.Entries.PushBack(entry)
	stream.NotifyNewEntry(entry)
	redisObj.version++
	return entry.ID, nil
}

func (s *Store) XRange(key string, startMST int64, startSeq int64, endMST int64, endSeq int64) ([]*storage.StreamEntry, error) {
	s.dataMutex.RLock()
	defer s.dataMutex.RUnlock()

	redisObj, exists := s.data[key]
	if !exists {
		return []*storage.StreamEntry{}, nil
	}

	if redisObj.Type != StreamType {
		return nil, errors.New("ERR WrongType Operation. NOT A STREAM")
	}

	stream := redisObj.Value.(*storage.Stream)
	result := stream.FilterEntries(startMST, startSeq, endMST, endSeq)
	
	return result, nil
}

func (s *Store) XReadStreams(keys []string, mst []int64, seq []int64) ([]*storage.ModStream, bool) {
	s.dataMutex.RLock()
	defer s.dataMutex.RUnlock()

	var response []*storage.ModStream = make([]*storage.ModStream, 0, len(keys))
	for i := 0; i < len(keys); i++ {
		if mst[i] == -1 && seq[i] == -1 {
			continue
		}

		entries, err := s.XRange(keys[i], mst[i], seq[i], -1, -1)
		if err != nil {
			return nil, false
		}

		if len(entries) != 0 && entries[0].IsIDEqual(mst[i], seq[i]) {
			entries = entries[1:]
		}

		if len(entries) == 0 {
			continue
		}

		response = append(response, &storage.ModStream{
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
func (s *Store) BXReadStreams(keys []string, mst []int64, seq []int64, expiryTime int64) ([]*storage.ModStream, error) {
	var expiryTimeMS time.Duration = time.Duration(expiryTime) * time.Millisecond

	var mstMp map[string]int64 = make(map[string]int64, len(keys))
	var seqMp map[string]int64 = make(map[string]int64, len(keys))

	for i := range keys {
		mstMp[keys[i]] = mst[i]
		seqMp[keys[i]] = seq[i]
	}

	var result []*storage.ModStream = make([]*storage.ModStream, 0, 1)

	waiterChannel := make(chan *storage.StreamNotification, 1)

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

		var notifyChList = &redisObj.Value.(*storage.Stream).Notify
		*notifyChList = append(*notifyChList, waiterChannel)
	}
	s.dataMutex.Unlock()

	// func to remove the channel from the Notify list of all its keys streams
	removeChFromAll := func() {
		for i := 0; i < len(keys); i++ {
			redisObj, _ := s.data[keys[i]]
			redisObj.Value.(*storage.Stream).RemoveChannel(waiterChannel)
		}
	}

	if expiryTimeMS == 0 {
		msg := <-waiterChannel
		removeChFromAll()
		result = append(result, &storage.ModStream{
			Key:     msg.Key,
			Entries: []*storage.StreamEntry{msg.Field},
		})
		return result, nil
	}

	timer := time.NewTimer(expiryTimeMS)
	for {
		select {
		case msg := <-waiterChannel:
			if msg.CheckValidNotification(mstMp, seqMp, mst[0]) {
				result = append(result, &storage.ModStream{
					Key:     msg.Key,
					Entries: []*storage.StreamEntry{msg.Field},
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
