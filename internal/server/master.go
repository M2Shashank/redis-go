package server

import (
	"strings"
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

func (s *Server) addMasterReplicationInfo() {
	s.replicationInfo["role"] = "master"
	s.replicationInfo["master_replid"] = "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
	s.replicationInfo["master_repl_offset"] = "0"
}

func InfoHandler(cmd *resp.Command, server *Server) resp.RESPType {
	info := strings.Builder{}
	fmt.Fprintf(&info, "# Replication\r\n")
	for key, value := range server.replicationInfo {
		fmt.Fprintf(&info, "%s:%s\r\n", key, value)
	}
	return resp.BulkString(info.String())
}

func (s *Server) ServerCommandHandler(cmd *resp.Command) resp.RESPType {
	switch cmd.Com {

	case "REPLCONF":
		return resp.SimpleString("OK")
	case "PSYNC":
		masterId := s.replicationInfo["master_replid"]
		masterOffset := s.replicationInfo["master_repl_offset"]
		str := "FULLRESYNC " + masterId + " " + masterOffset
		return resp.SimpleString(str)
	case "INFO":
		return InfoHandler(cmd, s)
	default:
		return resp.SimpleError("ERR unknown command '" + cmd.Com + "'")
		
	}
}