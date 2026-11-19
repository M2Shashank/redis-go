package command

import (
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
	cmd_handlers "github.com/codecrafters-io/redis-starter-go/internal/cmd_router/handlers"
)

var SimpleOK resp.RESPType = resp.SimpleString("OK")

// handles all store related commands.

func StoreCommandHandler(cmd *resp.Command, dataStore *store.Store) resp.RESPType {
	var result resp.RESPType
	switch cmd.Com {

	case "ECHO", "TYPE", "KEYS":
		result = cmd_handlers.General(cmd, dataStore)

	case "SET", "GET", "INCR":
		result = cmd_handlers.String(cmd, dataStore)

	case "RPUSH", "LPUSH", "LRANGE", "LLEN", "LPOP", "BLPOP":
		result = cmd_handlers.List(cmd, dataStore)

	case "XADD", "XRANGE", "XREAD":
		result = cmd_handlers.Stream(cmd, dataStore)

	case "ZADD", "ZRANK", "ZRANGE", "ZCARD", "ZSCORE", "ZREM":
		result = cmd_handlers.Zset(cmd, dataStore)

	case "SETBIT", "GETBIT", "STRLEN", "BITCOUNT", "BITOP":
		result = cmd_handlers.Bitmap(cmd, dataStore)

	case "GEOADD", "GEOPOS", "GEODIST":
		result = cmd_handlers.Geospacial(cmd, dataStore)

	default:
		result = resp.SimpleError("ERR Invalid or Unsupported Command")

	}

	return result
}
