package server

import (
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

func (s *Server) AuthHandler(clientName string, cmd *resp.Command) resp.RESPType {
	switch cmd.Com {
	case "ACL":
		if len(cmd.Args) < 1 {
			return resp.SimpleError("ERR wrong number of arguments for 'ACL' command")
		}

		switch cmd.Args[0] {
		case "WHOAMI":
			return resp.BulkString(clientName)
		case "GETUSER":
			if len(cmd.Args) != 2 {
				return resp.SimpleError("ERR wrong number of arguments for 'GETUSER' command")
			}

			var flagsResp resp.Array = resp.Array{}
			for _, flag := range s.Auth.GetUserFlags(cmd.Args[1]) {
				flagsResp = append(flagsResp, resp.BulkString(flag))
			}
			var pswdsResp resp.Array = resp.Array{}
			for _, pswd := range s.Auth.GetUserPswds(cmd.Args[1]) {
				pswdsResp = append(pswdsResp, resp.BulkString(pswd))
			}

			result := resp.Array{
				resp.BulkString("flags"),
				flagsResp,
				resp.BulkString("passwords"),
				pswdsResp,
			}
			return result
		case "SETUSER":
			var username string = cmd.Args[1]
			var pswd string = cmd.Args[2]
			if pswd[0] == '>' {
				s.Auth.SetUserPassword(username, pswd[1:])
				return resp.SimpleString("OK")
			}
		}
	}
	return resp.SimpleError("ERR unknown command '" + cmd.Com + "'")
}
