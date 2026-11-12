package command

import (
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

var SimpleOK resp.RESPType = resp.SimpleString("OK")

// handles all store related commands.

func StoreCommandHandler(cmd *resp.Command, dataStore *store.Store) resp.RESPType {
	var result resp.RESPType
	switch cmd.Com {

	case "PING", "ECHO", "TYPE", "KEYS":
		result = generalHandler(cmd, dataStore)

	case "SET", "GET", "INCR":
		result = stringHandler(cmd, dataStore)

	case "RPUSH", "LPUSH", "LRANGE", "LLEN", "LPOP", "BLPOP":
		result = listHandler(cmd, dataStore)

	case "XADD", "XRANGE", "XREAD":
		result = streamHandler(cmd, dataStore)

	default:
		result = resp.SimpleError("ERR Invalid or Unsupported Command")

	}

	return result
}
