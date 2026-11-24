package handlers

import (
	"github.com/codecrafters-io/redis-starter-go/internal/resp"

	store"github.com/codecrafters-io/redis-starter-go/internal/store/manager"
	Error "github.com/codecrafters-io/redis-starter-go/internal/errors"
)

func General(cmd *resp.Command, dataStore *store.Store) resp.RESPType {
	switch cmd.Com {

	case "ECHO":
		return resp.BulkString(cmd.Args[0])

	case "TYPE":
		var key string = cmd.Args[0]
		var typeStr string = dataStore.TypeOf(key)

		return resp.SimpleString(typeStr)

	case "KEYS":
		// Only handling for *
		var pattern string = cmd.Args[0]
		var keys []string = dataStore.GetKeysOfPattern(pattern)

		var response resp.RESPType = resp.ArrayToBulkRESP(keys)
		return response

	}

	return resp.SimpleError(Error.ERR_INVALID.Error())
}
