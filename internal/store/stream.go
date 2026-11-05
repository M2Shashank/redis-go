package store

type StreamEntry struct {
	ID     string
	Fields map[string]string
}

// stream object looks like -- map[string]*Stream 
// data[key] => &RedisObject{ StreamType, &Stream{ Entries: &DoublyLL[StreamEntry]{} }}
type Stream struct {
	Entries *DoublyLL[*StreamEntry]
}

func (s *Store) CreateStreamEntry(ID string, fields map[string]string) *StreamEntry {
	return &StreamEntry{
		ID:     ID,
		Fields: fields,
	}
}

func (s *Store) createStreamObject(key string) *RedisObject {
	obj := &RedisObject{
		Type: StreamType, 
		Value: &Stream{
			Entries: &DoublyLL[*StreamEntry]{},
		},
	}
	return obj
}

func (s *Store) AddStreamEntry(key string, entry *StreamEntry) bool {
	redisObj, exists := s.data[key]
	if exists && redisObj.Type != StreamType {
		return false
	}

	if !exists {
		s.data[key] = s.createStreamObject(key)
		redisObj = s.data[key]
	}

	stream := redisObj.Value.(*Stream)
	stream.Entries.PushBack(entry)
	return true
}