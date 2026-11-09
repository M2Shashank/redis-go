package command

import (
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

var SimpleOK resp.RESPType = resp.SimpleString("OK")

func ExecuteCommand(cmd *resp.Command, dataStore *store.Store) (resp.RESPType, bool) {
	var result resp.RESPType
	var writeCommand bool = false
	switch cmd.Com {

	case "PING", "ECHO", "TYPE", "INFO":
		result = generalHandler(cmd, dataStore)

	case "SET", "GET", "INCR":
		if cmd.Com != "GET" { writeCommand = true }
		result = stringHandler(cmd, dataStore)

	case "RPUSH", "LPUSH", "LRANGE", "LLEN", "LPOP", "BLPOP":
		if cmd.Com != "LRANGE" && cmd.Com != "LLEN" { writeCommand = true }
		result = listHandler(cmd, dataStore)

	case "XADD", "XRANGE", "XREAD":
		if cmd.Com == "XADD" { writeCommand = true }
		result = streamHandler(cmd, dataStore)

	default:
		result = resp.SimpleError("ERR Invalid or Unsupported Command")

	}

	return result, writeCommand
}
