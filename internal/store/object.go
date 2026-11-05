package store

// RedisObject represents a Redis object with its type and value.
// Represents how Redis stores different data types (string, list, stream) in memory.

type ObjectType int

const StringType ObjectType = 0
const ListType ObjectType = 1
const StreamType ObjectType = 2

type RedisObject struct {
	Type  ObjectType
	Value any
}

