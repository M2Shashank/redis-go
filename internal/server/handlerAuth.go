package server

import (
	"github.com/codecrafters-io/redis-starter-go/internal/resp"

	Error "github.com/codecrafters-io/redis-starter-go/internal/errors"
)

func (s *Server) AuthHandler(clientName string, cmd *resp.Command) resp.RESPType {
	switch cmd.Com {
	case "ACL":
		switch cmd.Args[0] {
		case "WHOAMI":
			return resp.BulkString(clientName)
		case "GETUSER":
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

	case "AUTH":
		valid := s.Auth.ValidateCredentials(cmd.Args[0], cmd.Args[1])
		if !valid {
			return resp.SimpleError(Error.ERR_WRONGPASS.Error())
		}

		return resp.SimpleString("OK")
	}

	return resp.SimpleError(Error.ERR_INVALID.Error())
}

func (s *Server) AuthVerdict(username string) bool {
	return s.Auth.IsAuthenticated(username)
}
