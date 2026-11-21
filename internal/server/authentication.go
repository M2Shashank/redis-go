package server

import (
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

func AuthHandler(client *Client, cmd *resp.Command) resp.RESPType {
	switch cmd.Com {
	case "ACL":
		if len(cmd.Args) < 1 {
			return resp.SimpleError("ERR wrong number of arguments for 'ACL' command")
		}

		return resp.BulkString(client.Username)

	default:
		return resp.SimpleError("ERR unknown command '" + cmd.Com + "'")
	}
}
