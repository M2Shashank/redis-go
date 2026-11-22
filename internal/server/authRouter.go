package server

import (
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

func (s *Server) AuthHandler(username string, cmd *resp.Command) resp.RESPType {
	switch cmd.Com {
	case "ACL":
		if len(cmd.Args) < 1 {
			return resp.SimpleError("ERR wrong number of arguments for 'ACL' command")
		}

		switch cmd.Args[0] {
		case "WHOAMI":
			return resp.BulkString(username)
		case "GETUSER":
			if len(cmd.Args) != 2 {
				return resp.SimpleError("ERR wrong number of arguments for 'GETUSER' command")
			}

			result := resp.Array{
				resp.BulkString("flags"),
				resp.Array{},
			}
			return result
		default:
			return resp.SimpleError("ERR unknown ACL subcommand '" + cmd.Args[0] + "'")
		}

	default:
		return resp.SimpleError("ERR unknown command '" + cmd.Com + "'")
	}
}
