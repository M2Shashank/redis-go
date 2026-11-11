package server

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

var emptyRDBFileHex string = "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2"

func (s *Server) MasterServerCommandHandler(cmd *resp.Command) resp.RESPType {
	if s.Role != "master" {
		return resp.SimpleError("ERR unknown command '" + cmd.Com + "'")
	}

	var result resp.RESPType
	switch cmd.Com {

	case "PSYNC":
		s.ReplicationManager.(*MasterManager).HandlePSYNC()

	case "WAIT":
		if len(cmd.Args) != 2 {
			return resp.SimpleError("ERR wrong number of arguments for 'WAIT' command")
		}
		numSlaves, err1 := strconv.Atoi(cmd.Args[0])
		timeLimit, err2 := strconv.Atoi(cmd.Args[1])
		if err1 != nil || err2 != nil {
			return resp.SimpleError("ERR invalid arguments for 'WAIT' command")
		}

		var noOfValidAcks int = s.ReplicationManager.(*MasterManager).HandleWait(numSlaves, timeLimit)
		result = resp.Integer(noOfValidAcks)

	default:
		result = resp.SimpleError("ERR unknown command '" + cmd.Com + "'")

	}
	return result
}

func (s *Server) HandleWait(numSlaves int, timeLimit int) resp.RESPType {
	var result int = s.ReplicationManager.(*MasterManager).HandleWait(numSlaves, timeLimit)
	return resp.Integer(result)
}

func (s *Server) AddSlave(client *Client) {
	s.ReplicationManager.(*MasterManager).AddSlave(client)
}

func (s *Server) SendRDBFile(client *Client) {
	rdbFile, _ := hex.DecodeString(emptyRDBFileHex)
	str := fmt.Sprintf("$%d\r\n", len(rdbFile))
	(client.Conn).Write([]byte(str))
	(client.Conn).Write(rdbFile)
}

func (s *Server) GetMasterAckChannel() *chan *ChannelItem {
	return &s.ReplicationManager.(*MasterManager).AckChannel
}

func (s *Server) DrainAckChannel() {
	s.ReplicationManager.(*MasterManager).DrainAckChannel()
}

// Synchronous for now
func (s *Server) WriteCommandToSlaves(rawCommand []byte) {
	s.ReplicationManager.(*MasterManager).WriteCommandToReplicas(rawCommand)
}

