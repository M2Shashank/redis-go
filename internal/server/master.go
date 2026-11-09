package server

import (
	"fmt"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

var emptyRDBFileHex string = "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2"

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
