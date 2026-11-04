package store

type ObjectType int

const StringType ObjectType = 0
const ListType ObjectType = 1

type RedisObject struct {
	Type ObjectType
	Value any
}