package command

import (
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

var SimpleOK resp.RESPType = resp.SimpleString("OK")

func ExecuteCommand(cmd *resp.Command, dataStore *store.Store) resp.RESPType {
	switch cmd.Com {

	case "PING", "ECHO", "TYPE", "INFO":
		return generalHandler(cmd, dataStore)

	case "SET", "GET", "INCR":
		return stringHandler(cmd, dataStore)

	case "RPUSH", "LPUSH", "LRANGE", "LLEN", "LPOP", "BLPOP":
		return listHandler(cmd, dataStore)

	case "XADD", "XRANGE", "XREAD":
		return streamHandler(cmd, dataStore)

	default:
		return resp.SimpleError("ERR Invalid or Unsupported Command")

	}
}
