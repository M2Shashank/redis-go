package server

import (
	"github.com/codecrafters-io/redis-starter-go/internal/resp"

	Error "github.com/codecrafters-io/redis-starter-go/internal/errors"
)

func (server *Server) PersistenceHandler(cmd *resp.Command) resp.RESPType {
	switch cmd.Com {
	case "CONFIG":
		// Implemented only config get for now.
		temp := server.Persistence.HandleConfigGET(cmd.Args[1])
		return resp.ArrayToBulkRESP(temp)
	}

	return resp.SimpleError(Error.ERR_INVALID.Error())
}

func (server *Server) ReplayAOFToStore(aofFileData []byte) {
	if len(aofFileData) == 0 || aofFileData == nil {
		return
	}

	// Nil connection client for executing commands
	replayClient := NewClient(nil, server)

	for {
		n, cmd := resp.ParseRESPCommand(aofFileData)
		if cmd == nil || n == 0 {
			break
		}

		_ = replayClient.ExecuteCommand(cmd)
		aofFileData = aofFileData[n:]
	}
}