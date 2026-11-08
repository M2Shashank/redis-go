package command

import (
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

func generalHandler(cmd *resp.Command, dataStore *store.Store) resp.RESPType {
	switch cmd.Com {

	case "PING":
		return resp.SimpleString("PONG")

	case "ECHO":
		var result resp.RESPType
		if len(cmd.Args) != 1 {
			result = resp.SimpleError("ERR Invalid command syntax for ECHO command")
		} else {
			result = resp.BulkString(cmd.Args[0])
		}
		return result

	case "TYPE":
		if len(cmd.Args) != 1 {
			return resp.SimpleError("ERR Invalid command syntax for TYPE command")
		}

		var key string = cmd.Args[0]
		var typeStr string = dataStore.TypeOf(key)

		return resp.SimpleString(typeStr)

	case "INFO":
		result := resp.BulkString("# Replication\r\nrole:master\r\n")
		return result

	default:
		return resp.SimpleError("ERR Invalid or Unsupported Command")

	}
}
